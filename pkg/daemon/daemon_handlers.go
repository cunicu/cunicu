// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	"errors"

	"go.uber.org/zap"
)

func (d *Daemon) OnInterfaceAdded(i *Interface) {
	i.Daemon = d
	i.Settings = d.Config.InterfaceSettings(i.Name())

	i.logger.Info("Added interface",
		zap.Any("pk", i.PublicKey()),
		zap.Any("type", i.Device.Type()),
		zap.Int("#peers", len(i.Peers)),
	)

	i.AddModifiedHandler(i)

	for _, f := range features {
		if fi, err := f.New(i); err == nil {
			i.features[f] = fi
		} else if !errors.Is(err, ErrFeatureDeactivated) {
			d.logger.Error("Failed to create feature", zap.Error(err))
		}
	}

	if err := i.Start(); err != nil {
		d.logger.Error("Failed to start interface", zap.Error(err))
	}
}

func (d *Daemon) OnInterfaceRemoved(i *Interface) {
	if err := i.Close(); err != nil {
		d.logger.Error("Failed to close interface", zap.Any("intf", i), zap.Error(err))
	}
}
