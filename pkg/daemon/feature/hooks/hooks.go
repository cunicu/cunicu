// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package hooks

import (
	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/log"
)

var Get = daemon.RegisterFeature(New, 70) //nolint:gochecknoglobals

type Hook interface {
	daemon.AllHandler
	daemon.PeerStateChangedHandler
}

type Interface struct {
	*daemon.Interface

	hooks []Hook

	logger *log.Logger
}

func New(i *daemon.Interface) (*Interface, error) {
	if len(i.Settings.Hooks) == 0 {
		return nil, daemon.ErrFeatureDeactivated
	}

	h := &Interface{
		Interface: i,
		logger:    log.Global.Named("hooks").With(zap.String("intf", i.Name())),
	}

	for _, hks := range i.Settings.Hooks {
		var hk Hook
		switch hks := hks.(type) {
		case *config.ExecHookSetting:
			hk = h.NewExecHook(hks)
		case *config.WebHookSetting:
			hk = h.NewWebHook(hks)
		}

		h.AddModifiedHandler(hk)
		h.AddPeerHandler(hk)
		h.AddPeerStateChangeHandler(hk)

		h.hooks = append(h.hooks, hk)
	}

	return h, nil
}

func (i *Interface) Start() error {
	i.logger.Info("Started hooks")

	for _, hk := range i.hooks {
		hk.OnInterfaceAdded(i.Interface)
	}

	return nil
}

func (i *Interface) Close() error {
	for _, hk := range i.hooks {
		hk.OnInterfaceRemoved(i.Interface)
	}

	return nil
}
