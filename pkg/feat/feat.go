// Package feat contains several sub-packages each implementing a dedicated feature.
package feat

import (
	"golang.zx2c4.com/wireguard/wgctrl"
	"riasc.eu/wice/pkg/config"
	"riasc.eu/wice/pkg/feat/autocfg"
	"riasc.eu/wice/pkg/feat/cfgsync"
	"riasc.eu/wice/pkg/feat/epdisc"
	"riasc.eu/wice/pkg/feat/hooks"
	"riasc.eu/wice/pkg/feat/hsync"
	"riasc.eu/wice/pkg/feat/pdisc"
	"riasc.eu/wice/pkg/feat/rtsync"
	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/watcher"
)

type Syncable interface {
	Sync() error
}

type Feature interface {
	Start() error
	Close() error
}

func NewFeatures(w *watcher.Watcher, cfg *config.Config, c *wgctrl.Client, b signaling.Backend) ([]Feature, *epdisc.EndpointDiscovery) {
	var ep *epdisc.EndpointDiscovery
	var feats = []Feature{}

	if cfg.AutoConfig.Enabled {
		feats = append(feats, autocfg.New(w, cfg, c))
	}

	if cfg.ConfigSync.Enabled {
		feats = append(feats, cfgsync.New(w, c, cfg.ConfigSync.Path, cfg.ConfigSync.Watch, cfg.WireGuard.Userspace, cfg.WireGuard.InterfaceFilter.MatchString))
	}

	if cfg.RouteSync.Enabled {
		feats = append(feats, rtsync.New(w, cfg.RouteSync.Table, cfg.RouteSync.Watch))
	}

	if cfg.HostSync.Enabled {
		feats = append(feats, hsync.New(w))
	}

	if cfg.EndpointDisc.Enabled {
		ep = epdisc.New(w, cfg, c, b)
		feats = append(feats, ep)
	}

	if cfg.PeerDisc.Enabled && cfg.PeerDisc.Community != "" {
		feats = append(feats, pdisc.New(w, c, b, cfg.PeerDisc.Community, cfg.PeerDisc.Whitelist))
	}

	if len(cfg.Hooks) > 0 {
		feats = append(feats, hooks.New(w, cfg, ep))
	}

	return feats, ep
}
