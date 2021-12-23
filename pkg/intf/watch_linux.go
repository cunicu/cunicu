package intf

import (
	"fmt"

	nl "github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"riasc.eu/wice/pkg/netlink"
)

func WatchWireguardKernelInterfaces(events chan InterfaceEvent, errors chan error) error {
	nlu := make(chan nl.LinkUpdate, 32)

	if err := nl.LinkSubscribeWithOptions(nlu, nil, nl.LinkSubscribeOptions{
		ErrorCallback: func(err error) {
			errors <- err
		},
	}); err != nil {
		return fmt.Errorf("failed to subscribe to netlink link event group: %w", err)
	}

	logger := zap.L().Named("wireguard")

	go func() {
		for lu := range nlu {
			logger.Debug("Received netlink link update", zap.Any("update", lu))
			if lu.Link.Type() != netlink.LinkTypeWireguard {
				continue
			}

			switch lu.Header.Type {
			case unix.RTM_NEWLINK:
				events <- InterfaceEvent{
					Op:   InterfaceAdded,
					Name: lu.Attrs().Name,
				}
			case unix.RTM_DELLINK:
				events <- InterfaceEvent{
					Op:   InterfaceDeleted,
					Name: lu.Attrs().Name,
				}
			}
		}
	}()

	return nil
}
