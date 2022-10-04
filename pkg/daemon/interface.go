package daemon

import (
	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/crypto"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Interface struct {
	*core.Interface

	Daemon   *Daemon
	Settings *config.InterfaceSettings

	Features map[string]Feature
}

func (d *Daemon) NewInterface(ci *core.Interface) (*Interface, error) {
	i := &Interface{
		Interface: ci,
		Daemon:    d,
		Settings:  d.Config.InterfaceSettings(ci.Name()),
		Features:  map[string]Feature{},
	}

	for _, fp := range SortedFeatures() {
		f, err := fp.New(i)
		if err != nil {
			return nil, err
		} else if f == nil {
			continue
		}

		i.Features[fp.Name] = f
	}

	return i, nil
}

func (i *Interface) Start() error {
	for _, f := range i.Features {
		if err := f.Start(); err != nil {
			return err
		}
	}

	return nil
}

func (i *Interface) Sync() error {
	for _, f := range i.Features {
		if s, ok := f.(SyncableFeature); ok {
			if err := s.Sync(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (i *Interface) Close() error {
	for _, feat := range i.Features {
		if err := feat.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (i *Interface) ConfigureDevice(cfg wgtypes.Config) error {
	if err := i.Daemon.client.ConfigureDevice(i.Name(), cfg); err != nil {
		return err
	}

	return i.Daemon.watcher.Sync()
}

func (i *Interface) AddPeer(pcfg *wgtypes.PeerConfig) error {
	return i.ConfigureDevice(wgtypes.Config{
		Peers: []wgtypes.PeerConfig{*pcfg},
	})
}

func (i *Interface) UpdatePeer(pcfg *wgtypes.PeerConfig) error {
	pcfg2 := *pcfg
	pcfg2.UpdateOnly = true

	return i.AddPeer(&pcfg2)
}

func (i *Interface) RemovePeer(pk crypto.Key) error {
	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey: wgtypes.Key(pk),
				Remove:    true,
			},
		},
	}

	return i.ConfigureDevice(cfg)
}
