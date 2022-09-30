package hooks

import (
	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/daemon/feature/epdisc"
	"go.uber.org/zap"
)

func init() {
	daemon.Features["hooks"] = &daemon.FeaturePlugin{
		New:         New,
		Description: "Hooks",
		Order:       70,
	}
}

type Hook interface {
	core.AllHandler
	epdisc.OnConnectionStateHandler
}

type Interface struct {
	*daemon.Interface

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

	for _, hk := range i.Settings.Hooks {
		switch hk := hk.(type) {
		case *config.ExecHookSetting:
			h.NewExecHook(hk)
		case *config.WebHookSetting:
			h.NewWebHook(hk)
		}
	}

	return h, nil
}

func (h *Interface) registerHook(j Hook) {
	h.Daemon.Watcher.OnAll(j)

	ep := h.Features["epdisc"].(*epdisc.Interface)
	ep.OnConnectionStateChange(j)
}

func (h *Interface) Start() error {
	h.logger.Info("Started hooks")

	return nil
}

func (h *Interface) Close() error {
	return nil
}
