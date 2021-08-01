package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"riasc.eu/wice/pkg/backend"
	"riasc.eu/wice/pkg/backend/base"
	"riasc.eu/wice/pkg/crypto"
)

const (
	annotationPrefix           string = "wice.riasc.eu"
	defaultAnnotationOffers    string = annotationPrefix + "/offers"
	defaultAnnotationPublicKey string = annotationPrefix + "/public-key"
)

type Backend struct {
	base.Backend
	config BackendConfig

	clientSet *kubernetes.Clientset
	informer  cache.SharedInformer

	term    chan struct{}
	updates chan NodeCallback
}

func init() {
	backend.Backends["k8s"] = &backend.BackendPlugin{
		New:         NewBackend,
		Description: "Exchange candidates via annotation in Kubernetes Node resource",
	}
}

func NewBackend(uri *url.URL, options map[string]string) (backend.Backend, error) {
	b := Backend{
		Backend: base.NewBackend(uri, options),
		term:    make(chan struct{}),
		updates: make(chan NodeCallback),
	}

	err := b.config.Parse(uri, options)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	kubeconfig := uri.Path
	var config *rest.Config
	if kubeconfig == "" {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		// if you want to change the loading rules (which files in which order), you can do so here

		configOverrides := &clientcmd.ConfigOverrides{}
		// if you want to change override values or bind them to flags, there are methods to help you

		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		config, err = kubeConfig.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	} else if kubeconfig == "incluster" {
		config, err = rest.InClusterConfig()
		if err != nil {
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
	b.Logger.Debug("Started watching node resources")

	go b.applyUpdates()
	b.Logger.Debug("Started batched updates")

	return &b, nil
}

func (b *Backend) SubscribeOffer(kp crypto.PublicKeyPair) (chan backend.Offer, error) {
	ch := b.Backend.SubscribeOffers(kp)

	// Process the node annotation at least once before we rely on the informer
	node, err := b.getNodeByPublicKey(kp.Theirs)
	if err == nil {
		b.processNode(node)
	}

	return ch, nil
}

func (b *Backend) PublishOffer(kp crypto.PublicKeyPair, offer backend.Offer) error {
	b.updateNode(func(node *corev1.Node) error {
		offerMapJson, ok := node.ObjectMeta.Annotations[b.config.AnnotationOffers]

		// Unmarshal
		var om backend.OfferMap
		if ok && offerMapJson != "" {
			err := json.Unmarshal([]byte(offerMapJson), &om)
			if err != nil {
				return err
			}
		} else {
			om = backend.OfferMap{}
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

	return b.Backend.PublishOffer(kp, offer)
}

func (b *Backend) WithdrawOffer(kp crypto.PublicKeyPair) error {
	b.updateNode(func(node *corev1.Node) error {
		delete(node.ObjectMeta.Annotations, b.config.AnnotationOffers)

		return nil
	})

	return b.Backend.WithdrawOffer(kp)
}

func (b *Backend) Close() error {
	close(b.term)

	return nil // TODO
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
