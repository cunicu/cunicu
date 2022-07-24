package watcher

import (
	"fmt"

	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

func (w *Watcher) watchKernel() error {
	nlu := make(chan netlink.LinkUpdate)

	if err := netlink.LinkSubscribeWithOptions(nlu, nil, netlink.LinkSubscribeOptions{
		ErrorCallback: func(err error) {
			w.errors <- err
		},
	}); err != nil {
		return fmt.Errorf("failed to subscribe to netlink link event group: %w", err)
	}

	go func() {
		w.logger.Debug("Start watching for changes of Wireguard kernel devices")

	out:
		for {
			select {
			case lu := <-nlu:
				w.logger.Debug("Received netlink link update",
					zap.Any("dev", lu.Link.Attrs().Name),
					zap.Any("type", lu.Header.Type))

				_, isWg := lu.Link.(*netlink.Wireguard)
				_, isTun := lu.Link.(*netlink.Tuntap)
				if !isWg && !isTun {
					continue
				}

				switch lu.Header.Type {
				case unix.RTM_NEWLINK:
					w.events <- InterfaceEvent{
						Op:   InterfaceAdded,
						Name: lu.Attrs().Name,
					}

				case unix.RTM_DELLINK:
					w.events <- InterfaceEvent{
						Op:   InterfaceDeleted,
						Name: lu.Attrs().Name,
					}
				}

			case <-w.stop:
				break out
			}
		}

		w.logger.Debug("Stop watching for changes of Wireguard kernel devices")
	}()

	return nil
}
