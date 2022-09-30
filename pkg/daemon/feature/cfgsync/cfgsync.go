// Package cfgsync synchronizes existing WireGuard configuration files with the kernel/userspace WireGuard device.
package cfgsync

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/device"
	"github.com/stv0g/cunicu/pkg/wg"
	"go.uber.org/zap"
)

func init() {
	daemon.Features["cfgsync"] = &daemon.FeaturePlugin{
		New:         New,
		Description: "Config-file synchronization",
		Order:       20,
	}
}

// Interface synchronizes the WireGuard device configuration with an on-disk configuration file.
type Interface struct {
	*daemon.Interface

	path string

	logger *zap.Logger
}

// New creates a new Syncer
func New(i *daemon.Interface) (daemon.Feature, error) {
	if !i.Settings.ConfigSync.Enabled {
		return nil, nil
	}

	cs := &Interface{
		Interface: i,
		path:      filepath.Join(wg.ConfigPath, fmt.Sprintf("%s.conf", i.Name())),
		logger:    zap.L().Named("cfgsync").With(zap.String("intf", i.Name())),
	}

	if err := i.SyncConfig(cs.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		cs.logger.Fatal("Failed to sync interface configuration",
			zap.Error(err),
			zap.String("config_file", cs.path))
	}

	i.OnModified(cs)

	return cs, nil
}

func (cs *Interface) Start() error {
	cs.logger.Info("Started configuration file synchronization")

	return nil
}

func (cs *Interface) Close() error {
	return nil
}

func (cs *Interface) Sync() error {
	des, err := os.ReadDir(wg.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to list config files in '%s': %w", wg.ConfigPath, err)
	}

	for _, de := range des {
		cs.handleFsnotifyEvent(fsnotify.Event{
			Name: filepath.Join(wg.ConfigPath, de.Name()),
			Op:   fsnotify.Write,
		})
	}

	return nil
}

func (cs *Interface) handleFsnotifyEvent(event fsnotify.Event) {
	cfg := event.Name
	filename := path.Base(cfg)
	extension := path.Ext(filename)
	name := strings.TrimSuffix(filename, extension)

	if extension != ".conf" || !cs.Daemon.Config.InterfaceFilter(name) {
		return
	}

	i := cs.Daemon.InterfaceByName(name)

	if event.Op&(fsnotify.Create|fsnotify.Write) != 0 {
		if i == nil {
			var err error
			if _, err = device.NewDevice(name, cs.Settings.WireGuard.UserSpace); err != nil {
				cs.logger.Error("Failed to create new device",
					zap.Error(err),
					zap.String("config_file", cfg))
			}
		} else {
			if err := i.SyncConfig(cfg); err != nil {
				cs.logger.Error("Failed to sync interface configuration",
					zap.Error(err),
					zap.String("config_file", cfg))
			}
		}

	} else if event.Op&(fsnotify.Remove) != 0 {
		if i == nil {
			cs.logger.Warn("Ignoring unknown interface")
			return
		}

		// TODO: Do we really want to delete devices if their config file vanish?
		// Maybe make this configurable?
		if err := i.KernelDevice.Close(); err != nil {
			cs.logger.Error("Failed to close interface", zap.Error(err))
		}
	} else if event.Op&(fsnotify.Rename) != 0 {
		// TODO: This is not supported yet
		cs.logger.Warn("We do not support tracking renamed WireGuard configuration files yet")
	}
}
