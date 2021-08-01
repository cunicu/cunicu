package k8s

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"

	"riasc.eu/wice/pkg/backend"
	"riasc.eu/wice/pkg/crypto"
)

func (b *Backend) onNodeAdd(obj interface{}) {
	node := obj.(*corev1.Node)

	b.Logger.WithField("name", node.ObjectMeta.Name).Debug("Node added")
	b.processNode(node)
}

func (b *Backend) onNodeUpdate(_ interface{}, new interface{}) {
	newNode := new.(*corev1.Node)

	b.Logger.WithField("name", newNode.ObjectMeta.Name).Debug("Node updated")
	b.processNode(newNode)
}

func (b *Backend) onNodeDelete(obj interface{}) {
	node := obj.(*corev1.Node)

	b.Logger.WithField("name", node.ObjectMeta.Name).Debug("Node deleted")
	b.processNode(node)
}

func (b *Backend) processNode(node *corev1.Node) {
	// Ignore ourself
	if node.ObjectMeta.Name == b.config.NodeName {
		b.Logger.WithField("node", node.ObjectMeta.Name).Trace("Ignoring ourself")
		return
	}

	// Check if required annotations are present
	offersJson, ok := node.ObjectMeta.Annotations[b.config.AnnotationOffers]
	if !ok {
		b.Logger.WithField("node", node.ObjectMeta.Name).Trace("Missing candidate annotation")
		return
	}

	keyString, ok := node.ObjectMeta.Annotations[b.config.AnnotationPublicKey]
	if !ok {
		b.Logger.WithField("node", node.ObjectMeta.Name).Trace("Missing public key annotation")
		return
	}

	var err error
	var theirPK crypto.Key
	theirPK, err = crypto.ParseKey(keyString)
	if err != nil {
		b.Logger.WithError(err).Warn("Failed to parse public key")
	}

	var om backend.OfferMap
	err = json.Unmarshal([]byte(offersJson), &om)
	if err != nil {
		b.Logger.WithError(err).Warn("Failed to parse candidate annotation")
		return
	}

	for ourPK, o := range om {
		kp := crypto.PublicKeyPair{
			Ours:   ourPK,
			Theirs: theirPK,
		}

		ch, ok := b.Offers[kp]
		if !ok {
			b.Logger.WithField("kp", kp).Warn("Found candidates for unknown peer")
			continue
		}

		ch <- o
	}
}
