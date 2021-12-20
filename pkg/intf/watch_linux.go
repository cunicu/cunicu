package intf

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	nl "github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

func WatchKernelWireguardInterfaces(lu chan InterfaceEvent, errors chan error) error {
	chNl := make(chan nl.LinkUpdate, 32)
	if err := nl.LinkSubscribeWithOptions(chNl, nil, nl.LinkSubscribeOptions{
		ErrorCallback: func(err error) {
			errors <- err
		},
	}); err != nil {
		return fmt.Errorf("failed to subscribe to netlink link event group: %w", err)
	}

	go func() {
		for {
			select {
			case lu := <-chNl:
				log.WithField("update", lu).Trace("Received netlink link update")
				if lu.Link.Type() != nl.LinkTypeWireguard {
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

		}
	}()

	return nil
}
