package intf

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	nl "github.com/vishvananda/netlink"
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

	go func() {
		for lu := range nlu {
			log.WithField("update", lu).Trace("Received netlink link update")
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
