// Package rtsync synchronizes the kernel routing table with the AllowedIPs of each WireGuard peer
package rtsync

import (
	"errors"
	"net/netip"

	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/daemon"
	"go.uber.org/zap"
)

var errNotSupported = errors.New("not supported on this platform")

func init() { //nolint:gochecknoinits
	daemon.RegisterFeature("rtsync", "Route synchronization", New, 30)
}

type Interface struct {
	*daemon.Interface

	gwMap map[netip.Addr]*core.Peer
	stop  chan struct{}

	logger *zap.Logger
}

func New(i *daemon.Interface) (daemon.Feature, error) {
	if !i.Settings.SyncRoutes {
		return nil, nil
	}

	rs := &Interface{
		Interface: i,
		gwMap:     map[netip.Addr]*core.Peer{},
		stop:      make(chan struct{}),
		logger:    zap.L().Named("rtsync").With(zap.String("intf", i.Name())),
	}

	i.OnPeer(rs)

	return rs, nil
}

func (i *Interface) Start() error {
	i.logger.Info("Started route synchronization")

	go func() {
		if i.Settings.WatchRoutes {
			if err := i.watchKernel(); err != nil {
				if errors.Is(err, errNotSupported) {
					i.logger.Warn("Watching the kernel routing table is not supported on this platform")
				} else {
					i.logger.Error("Failed to watch kernel routing table", zap.Error(err))
				}
			}
		} else {
			// Only perform initial sync, if we are not continuously watching for changes
			if err := i.syncKernel(); err != nil {
				if errors.Is(err, errNotSupported) {
					i.logger.Warn("Synchronizing the routing table is not supported on this platform")
				} else {
					i.logger.Error("Failed to sync kernel routing table", zap.Error(err))
				}
			}
		}
	}()

	return nil
}

func (i *Interface) Close() error {
	close(i.stop)

	return nil
}

func (i *Interface) Sync() error {
	if err := i.syncKernel(); err != nil {
		if errors.Is(err, errNotSupported) {
			i.logger.Warn("Synchronizing the routing table is not supported on this platform")
		} else {
			return err
		}
	}

	return nil
}
