// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package rtsync synchronizes the kernel routing table with the AllowedIPs of each WireGuard peer
package rtsync

import (
	"errors"
	"net/netip"

	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/log"
)

var errNotSupported = errors.New("not supported on this platform")

var Get = daemon.RegisterFeature(New, 30) //nolint:gochecknoglobals

type Interface struct {
	*daemon.Interface

	gwMap map[netip.Addr]*daemon.Peer
	stop  chan struct{}

	logger *log.Logger
}

func New(i *daemon.Interface) (*Interface, error) {
	if !i.Settings.SyncRoutes {
		return nil, daemon.ErrFeatureDeactivated
	}

	rs := &Interface{
		Interface: i,
		gwMap:     map[netip.Addr]*daemon.Peer{},
		stop:      make(chan struct{}),
		logger:    log.Global.Named("rtsync").With(zap.String("intf", i.Name())),
	}

	i.AddPeerHandler(rs)

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
