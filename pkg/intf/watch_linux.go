package intf

func WatchKernelWireguardInterfaces(chan InterfaceEvent, chan error) error {
	chNl := make(chan netlink.LinkUpdate, 32)
	err := netlink.LinkSubscribeWithOptions(chNl, nil, netlink.LinkSubscribeOptions{
		ErrorCallback: func(err error) {
			errors <- err
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to subscribe to netlink link event group: %w", err)
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
