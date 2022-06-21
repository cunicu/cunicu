//go:build !linux

package core

import (
	"golang.zx2c4.com/wireguard/wgctrl"
	"riasc.eu/wice/internal/config"
	"riasc.eu/wice/internal/errors"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"
)

func WireguardModuleExists() bool {
	return false
}

func CreateKernelInterface(name string, client *wgctrl.Client, backend signaling.Backend, events chan *pb.Event, cfg *config.Config) (Interface, error) {
	return nil, errors.ErrNotSupported
}
