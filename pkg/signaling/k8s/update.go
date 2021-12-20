package k8s

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/util/retry"
)

type NodeCallback func(*corev1.Node) error
type AnnotationCallback func(string) (string, error)
type AnnotationCallbackMap map[string]AnnotationCallback

func (b *Backend) applyUpdates() {
	var cbs []NodeCallback
	var timer time.Timer

	for {
		select {
		case cb := <-b.updates:
			cbs = append(cbs, cb)
			timer = *time.NewTimer(50 * time.Millisecond)

		case <-timer.C:
			if len(cbs) > 0 {
				b.logger.Debugf("Applying %d batched updates", len(cbs))

				nodes := b.clientSet.CoreV1().Nodes()

				if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
					node, err := nodes.Get(context.TODO(), b.config.NodeName, metav1.GetOptions{})
					if err != nil {
						return fmt.Errorf("failed to get latest version of node %s: %w", b.config.NodeName, err)
					}

					for _, cb := range cbs {
						if err := cb(node); err != nil {
							return err
						}
					}

					_, err = nodes.Update(context.TODO(), node, metav1.UpdateOptions{})

					return err
				}); err != nil {
					b.logger.WithError(err).Error("Failed to update node")
				}

				cbs = nil
			}
		}
	}
}

func (b *Backend) updateNode(cb NodeCallback) {
	b.updates <- cb
}
