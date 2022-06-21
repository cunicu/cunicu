package core

import (
	"fmt"
	"os"
	"path"
	"strings"

	"go.uber.org/zap"

	"github.com/fsnotify/fsnotify"
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

func WatchWireguardUserspaceInterfaces(events chan InterfaceEvent, errors chan error) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	if _, err := os.Stat(wireguardSockDir); !os.IsNotExist(err) {
		if err := watcher.Add(wireguardSockDir); err != nil {
			return fmt.Errorf("failed to watch %s: %w", wireguardSockDir, err)
		}
	}

	logger := zap.L().Named("wireguard")

	go func() {
		for {
			select {

			// Fsnotify events
			case event := <-watcher.Events:
				zap.L().Debug("Received fsnotify event", zap.Any("event", event))

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
					logger.Warn("Unknown fsnotify event", zap.Any("event", event))
				}

			// Fsnotify errors
			case errors <- <-watcher.Errors:
				logger.Debug("Error while watching for link changes")
			}
		}
	}()

	return nil
}

func normalizeSocketName(name string) string {
	name = path.Base(name)
	return strings.TrimSuffix(name, ".sock")
}
