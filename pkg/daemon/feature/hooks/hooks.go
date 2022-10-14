package hooks

import (
	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/daemon/feature/epdisc"
	"go.uber.org/zap"
)

func init() {
	daemon.RegisterFeature("hooks", "Hooks", New, 70)
}

type Hook interface {
	core.AllHandler
	epdisc.OnConnectionStateHandler
}

type Interface struct {
	*daemon.Interface

	hooks []Hook

	logger *zap.Logger
}

func New(i *daemon.Interface) (daemon.Feature, error) {
	if len(i.Settings.Hooks) == 0 {
		return nil, nil
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

		h.OnModified(hk)
		h.OnPeer(hk)

		if f, ok := h.Features["epdisc"]; ok {
			f.(*epdisc.Interface).OnConnectionStateChange(hk)
		}

		h.hooks = append(h.hooks, hk)
	}

	return h, nil
}

func (h *Interface) Start() error {
	h.logger.Info("Started hooks")

	for _, hk := range h.hooks {
		hk.OnInterfaceAdded(h.Interface.Interface)
	}

	return nil
}

func (h *Interface) Close() error {
	for _, hk := range h.hooks {
		hk.OnInterfaceRemoved(h.Interface.Interface)
	}

	return nil
}
