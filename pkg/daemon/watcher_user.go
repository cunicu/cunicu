// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/wg"
)

func normalizeSocketName(name string) string {
	name = filepath.Base(name)
	return strings.TrimSuffix(name, ".sock")
}

func (w *Watcher) watchUserInterfaces() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	if _, err := os.Stat(wg.SocketPath); !os.IsNotExist(err) {
		if err := watcher.Add(wg.SocketPath); err != nil {
			return fmt.Errorf("failed to watch %s: %w", wg.SocketPath, err)
		}
	}

	go func() {
	out:
		for {
			select {
			// Fsnotify events
			case event := <-watcher.Events:
				w.logger.Debug("Received fsnotify event", zap.Any("event", event))

				name := normalizeSocketName(event.Name)

				switch {
				case event.Op&fsnotify.Create == fsnotify.Create:
					w.events <- InterfaceEvent{
						Op:   InterfaceAdded,
						Name: name,
					}
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					w.events <- InterfaceEvent{
						Op:   InterfaceDeleted,
						Name: name,
					}
				default:
					w.logger.Warn("Unknown fsnotify event", zap.Any("event", event))
				}

			// Fsnotify errors
			case w.errors <- <-watcher.Errors:
				w.logger.Debug("Error while watching for link changes")

			case <-w.stop:
				break out
			}
		}

		w.logger.Debug("Stopped watching for changes of WireGuard userspace devices")
	}()

	return nil
}
