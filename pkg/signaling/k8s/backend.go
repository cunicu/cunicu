package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/socket"
)

const (
	annotationPrefix           string = "wice.riasc.eu"
	defaultAnnotationOffers    string = annotationPrefix + "/offers"
	defaultAnnotationPublicKey string = annotationPrefix + "/public-key"
)

type Backend struct {
	logger log.FieldLogger
	offers map[crypto.PublicKeyPair]chan signaling.Offer

	config BackendConfig

	clientSet *kubernetes.Clientset
	informer  cache.SharedInformer

	term    chan struct{}
	updates chan NodeCallback

	server *socket.Server
}

func init() {
	signaling.Backends["k8s"] = &signaling.BackendPlugin{
		New:         NewBackend,
		Description: "Exchange candidates via annotation in Kubernetes Node resource",
	}
}

func NewBackend(uri *url.URL, server *socket.Server) (signaling.Backend, error) {
	var config *rest.Config
	var err error

	logFields := log.Fields{
		"logger":  "backend",
		"backend": uri.Scheme,
	}

	b := Backend{
		offers:  make(map[crypto.PublicKeyPair]chan signaling.Offer),
		logger:  log.WithFields(logFields),
		term:    make(chan struct{}),
		updates: make(chan NodeCallback),
		server:  server,
		config:  defaultConfig,
	}

	if err := b.config.Parse(uri); err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	kubeconfig := uri.Path
	if kubeconfig == "" {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		// if you want to change the loading rules (which files in which order), you can do so here

		configOverrides := &clientcmd.ConfigOverrides{}
		// if you want to change override values or bind them to flags, there are methods to help you

		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

		if config, err = kubeConfig.ClientConfig(); err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	} else if kubeconfig == "incluster" {

		if config, err = rest.InClusterConfig(); err != nil {
			return nil, fmt.Errorf("failed to get incluster configuration: %w", err)
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to get configuration from flags: %w", err)
		}
	}

	// Create the clientset
	b.clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	// Create the shared informer factory and use the client to connect to
	// Kubernetes
	factory := informers.NewSharedInformerFactoryWithOptions(b.clientSet, 0,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			// options.LabelSelector = b.config.AnnotationPublicKey
		}))

	// Get the informer for the right resource, in this case a Pod
	b.informer = factory.Core().V1().Nodes().Informer()

	b.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    b.onNodeAdd,
		UpdateFunc: b.onNodeUpdate,
		DeleteFunc: b.onNodeDelete,
	})

	go b.informer.Run(b.term)
	b.logger.Debug("Started watching node resources")

	go b.applyUpdates()
	b.logger.Debug("Started batched updates")

	b.server.BroadcastEvent(&pb.Event{
		Type:  "backend",
		State: "ready",
	})

	return &b, nil
}

func (b *Backend) SubscribeOffer(kp crypto.PublicKeyPair) (chan signaling.Offer, error) {
	b.logger.WithField("kp", kp).Info("Subscribe to offers from peer")

	ch, ok := b.offers[kp]
	if !ok {
		ch = make(chan signaling.Offer, 100)
		b.offers[kp] = ch
	}

	// Process the node annotation at least once before we rely on the informer
	node, err := b.getNodeByPublicKey(kp.Theirs)
	if err == nil {
		b.processNode(node)
	}

	return ch, nil
}

func (b *Backend) PublishOffer(kp crypto.PublicKeyPair, offer signaling.Offer) error {
	b.updateNode(func(node *corev1.Node) error {
		offerMapJson, ok := node.ObjectMeta.Annotations[b.config.AnnotationOffers]

		// Unmarshal
		var om signaling.OfferMap
		if ok && offerMapJson != "" {
			if err := json.Unmarshal([]byte(offerMapJson), &om); err != nil {
				return err
			}
		} else {
			om = signaling.OfferMap{}
		}

		// Update
		om[kp.Theirs] = offer

		// Marshal
		offerMapJsonNew, err := json.Marshal(&om)
		if err != nil {
			return err
		}

		node.ObjectMeta.Annotations[b.config.AnnotationOffers] = string(offerMapJsonNew)
		node.ObjectMeta.Annotations[b.config.AnnotationPublicKey] = kp.Ours.String()

		return nil
	})

	b.logger.WithField("kp", kp).WithField("offer", offer).Debug("Published offer")

	return nil
}

func (b *Backend) Close() error {
	close(b.term)

	return nil // TODO
}

func (b *Backend) Tick() {

}

func (b *Backend) getNodeByPublicKey(pk crypto.Key) (*corev1.Node, error) {
	coreV1 := b.clientSet.CoreV1()
	nodes, err := coreV1.Nodes().List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", b.config.AnnotationPublicKey, pk),
	})
	if err != nil {
		return nil, err
	}

	if len(nodes.Items) != 1 {
		return nil, fmt.Errorf("could not find node with public key: %s", pk)
	}

	return &nodes.Items[0], nil
}
