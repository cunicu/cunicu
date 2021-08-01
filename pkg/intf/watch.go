package intf

import (
	"fmt"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"

	"github.com/fsnotify/fsnotify"
	"github.com/vishvananda/netlink"

	nl "riasc.eu/wice/pkg/netlink"
)

const (
	wireguardSockDir = "/var/run/wireguard/"

	InterfaceAdded InterfaceEventOp = iota
	InterfaceDeleted
)

type InterfaceEventOp int
type InterfaceEvent struct {
	Op   InterfaceEventOp
	Name string
}

func (ls InterfaceEventOp) String() string {
	switch ls {
	case InterfaceAdded:
		return "added"
	case InterfaceDeleted:
		return "deleted"
	default:
		return ""
	}
}

func (e InterfaceEvent) String() string {
	return fmt.Sprintf("%s %s", e.Name, e.Op)
}

func WatchWireguardInterfaces() (chan InterfaceEvent, chan error, error) {
	events := make(chan InterfaceEvent, 16)
	errors := make(chan error, 16)

	done := make(<-chan struct{})

	// Watch kernel interfaces
	chNl := make(chan netlink.LinkUpdate, 32)
	err := netlink.LinkSubscribeWithOptions(chNl, done, netlink.LinkSubscribeOptions{
		ErrorCallback: func(err error) {
			errors <- err
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to subscribe to netlink link event group: %w", err)
	}

	// Watch userspace UAPI sockets
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	if _, err := os.Stat(wireguardSockDir); !os.IsNotExist(err) {
		err = watcher.Add(wireguardSockDir)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to watch %s: %w", wireguardSockDir, err)
		}
	}

	go func() {
		for {
			select {
			// Netlink link updates
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

			// Fsnotify events
			case event := <-watcher.Events:
				log.WithField("event", event).Trace("Received fsnotify event")

				name := normalizeSocketName(event.Name)

				if event.Op&fsnotify.Create == fsnotify.Create {
					events <- InterfaceEvent{
						Op:   InterfaceAdded,
						Name: name,
					}
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					events <- InterfaceEvent{
						Op:   InterfaceDeleted,
						Name: name,
					}
				} else {
					log.Warn("Unknown fsnotify event: %+v", event)
				}

			// Fsnotify errors
			case errors <- <-watcher.Errors:
				log.Trace("Error while watching for link changes")
			}
		}
	}()

	return events, errors, nil
}

func normalizeSocketName(name string) string {
	name = path.Base(name)
	return strings.TrimSuffix(name, ".sock")
}
