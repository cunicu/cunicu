package daemon

import (
	"github.com/stv0g/cunicu/pkg/core"
	"go.uber.org/zap"
)

type InterfaceHandler interface {
	OnInterfaceAdded(i *Interface)
	OnInterfaceRemoved(i *Interface)
}

func (d *Daemon) OnInterface(h InterfaceHandler) {
	d.onInterface = append(d.onInterface, h)
}

func (d *Daemon) OnInterfaceAdded(ci *core.Interface) {
	i, err := d.NewInterface(ci)
	if err != nil {
		d.logger.Error("Failed to add interface", zap.Error(err))
	}

	d.interfaces[ci] = i

	if err := i.Start(); err != nil {
		d.logger.Error("Failed to start interface", zap.Error(err))
	}

	for _, h := range d.onInterface {
		h.OnInterfaceAdded(i)
	}
}

func (d *Daemon) OnInterfaceRemoved(ci *core.Interface) {
	i := d.interfaces[ci]

	for _, h := range d.onInterface {
		h.OnInterfaceRemoved(i)
	}

	if err := i.Close(); err != nil {
		d.logger.Error("Failed to close interface", zap.Error(err))
	}

	delete(d.interfaces, ci)
}
