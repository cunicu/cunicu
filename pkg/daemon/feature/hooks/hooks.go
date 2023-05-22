package hooks

import (
	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/daemon"
	"go.uber.org/zap"
)

var Get = daemon.RegisterFeature(New, 70) //nolint:gochecknoglobals

type Hook interface {
	daemon.AllHandler
	daemon.PeerStateChangedHandler
}

type Interface struct {
	*daemon.Interface

	hooks []Hook

	logger *zap.Logger
}

func New(i *daemon.Interface) (*Interface, error) {
	if len(i.Settings.Hooks) == 0 {
		return nil, daemon.ErrFeatureDeactivated
	}

	h := &Interface{
		Interface: i,
		logger:    zap.L().Named("hooks").With(zap.String("intf", i.Name())),
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
