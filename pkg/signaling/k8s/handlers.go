package k8s

import (
	"encoding/json"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"

	"riasc.eu/wice/pkg/crypto"
)

func (b *Backend) onNodeAdd(obj interface{}) {
	node := obj.(*corev1.Node)

	b.logger.Debug("Node added", zap.String("name", node.ObjectMeta.Name))
	b.processNode(node)
}

func (b *Backend) onNodeUpdate(_ interface{}, new interface{}) {
	newNode := new.(*corev1.Node)

	b.logger.Debug("Node updated", zap.String("name", newNode.ObjectMeta.Name))
	b.processNode(newNode)
}

func (b *Backend) onNodeDelete(obj interface{}) {
	node := obj.(*corev1.Node)

	b.logger.Debug("Node deleted", zap.String("name", node.ObjectMeta.Name))
	b.processNode(node)
}

func (b *Backend) processNode(node *corev1.Node) {
	// Ignore ourself
	if node.ObjectMeta.Name == b.config.NodeName {
		b.logger.Debug("Ignoring ourself", zap.String("node", node.ObjectMeta.Name))
		return
	}

	// Check if required annotations are present
	offersJson, ok := node.ObjectMeta.Annotations[b.config.AnnotationOffers]
	if !ok {
		b.logger.Debug("Missing candidate annotation", zap.String("node", node.ObjectMeta.Name))
		return
	}

	keyString, ok := node.ObjectMeta.Annotations[b.config.AnnotationPublicKey]
	if !ok {
		b.logger.Debug("Missing public key annotation", zap.String("node", node.ObjectMeta.Name))
		return
	}

	var err error
	var theirPK crypto.Key
	theirPK, err = crypto.ParseKey(keyString)
	if err != nil {
		b.logger.Warn("Failed to parse public key", zap.Error(err))
	}

	var om OfferMap

	if err := json.Unmarshal([]byte(offersJson), &om); err != nil {
		b.logger.Warn("Failed to parse candidate annotation", zap.Error(err))
		return
	}

	for ourPK, o := range om {
		kp := crypto.KeyPair{
			Ours:   ourPK,
			Theirs: theirPK,
		}

		ch, ok := b.offers[kp]
		if !ok {
			b.logger.Warn("Found candidates for unknown peer", zap.Any("kp", kp))
			continue
		}

		ch <- o
	}
}
