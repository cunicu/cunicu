package hooks

import (
	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/feat/epdisc"
	"github.com/stv0g/cunicu/pkg/watcher"
	"go.uber.org/zap"
)

type Hooks struct {
	config  *config.Config
	watcher *watcher.Watcher
	epdisc  *epdisc.EndpointDiscovery

	logger *zap.Logger
}

func New(w *watcher.Watcher, cfg *config.Config, ep *epdisc.EndpointDiscovery) *Hooks {
	h := &Hooks{
		config:  cfg,
		watcher: w,
		epdisc:  ep,

		logger: zap.L().Named("hooks"),
	}

	for _, hk := range cfg.Hooks {
		switch hk := hk.(type) {
		case *config.ExecHookSetting:
			h.NewExecHook(hk)
		case *config.WebHookSetting:
			h.NewWebHook(hk)
		}
	}

	return h
}

func (h *Hooks) NewExecHook(cfg *config.ExecHookSetting) {
	hk := &ExecHook{
		ExecHookSetting: cfg,
		logger: h.logger.Named("exec").With(
			zap.String("command", cfg.Command),
		),
	}

	h.logger.Debug("Created new exec hook", zap.Any("hook", hk))

	h.watcher.OnAll(hk)
	h.epdisc.OnConnectionStateChange(hk)
}

func (h *Hooks) NewWebHook(cfg *config.WebHookSetting) {
	hk := &WebHook{
		WebHookSetting: cfg,
		logger: h.logger.Named("web").With(
			zap.Any("url", cfg.URL),
		),
	}

	h.logger.Debug("Created new web hook", zap.Any("hook", hk))

	h.watcher.OnAll(hk)
	h.epdisc.OnConnectionStateChange(hk)
}

func (h *Hooks) Start() error {
	h.logger.Info("Started hooks")

	return nil
}

func (h *Hooks) Close() error {
	return nil
}
