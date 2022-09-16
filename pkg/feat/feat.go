// Package feat contains several sub-packages each implementing a dedicated feature.
package feat

import (
	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/feat/autocfg"
	"github.com/stv0g/cunicu/pkg/feat/cfgsync"
	"github.com/stv0g/cunicu/pkg/feat/epdisc"
	"github.com/stv0g/cunicu/pkg/feat/hooks"
	"github.com/stv0g/cunicu/pkg/feat/hsync"
	"github.com/stv0g/cunicu/pkg/feat/pdisc"
	"github.com/stv0g/cunicu/pkg/feat/rtsync"
	"github.com/stv0g/cunicu/pkg/signaling"
	"github.com/stv0g/cunicu/pkg/watcher"
	"golang.zx2c4.com/wireguard/wgctrl"
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

	if cfg.DefaultInterfaceSettings.AutoConfig.Enabled {
		feats = append(feats, autocfg.New(w, cfg, c))
	}

	if cfg.DefaultInterfaceSettings.ConfigSync.Enabled {
		feats = append(feats, cfgsync.New(w, c, cfg.DefaultInterfaceSettings.ConfigSync.Path, cfg.DefaultInterfaceSettings.ConfigSync.Watch, cfg.DefaultInterfaceSettings.WireGuard.UserSpace, cfg.InterfaceFilter))
	}

	if cfg.DefaultInterfaceSettings.RouteSync.Enabled {
		feats = append(feats, rtsync.New(w, cfg.DefaultInterfaceSettings.RouteSync.Table, cfg.DefaultInterfaceSettings.RouteSync.Watch))
	}

	if cfg.DefaultInterfaceSettings.HostSync.Enabled {
		feats = append(feats, hsync.New(w, cfg.DefaultInterfaceSettings.HostSync.Domain))
	}

	if cfg.DefaultInterfaceSettings.EndpointDisc.Enabled {
		ep = epdisc.New(w, cfg, c, b)
		feats = append(feats, ep)
	}

	if cfg.DefaultInterfaceSettings.PeerDisc.Enabled && crypto.Key(cfg.DefaultInterfaceSettings.PeerDisc.Community).IsSet() {
		feats = append(feats, pdisc.New(w, c, b, cfg))
	}

	if len(cfg.Hooks) > 0 {
		feats = append(feats, hooks.New(w, cfg, ep))
	}

	return feats, ep
}
