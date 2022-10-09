package device

import (
	"net"
	"os/exec"

	"github.com/stv0g/cunicu/pkg/errors"
)

func (d *BSDKernelDevice) AddRoute(dst net.IPNet, gw net.IP, table int) error {
	if table != 0 {
		return errors.ErrNotSupported
	}

	if gw == nil {
		return exec.Command("route", "add", "-net", dst.String(), "-interface", d.Name()).Run()
	} else {
		return exec.Command("route", "add", "-net", dst.String(), gw.String()).Run()
	}
}

func (d *BSDKernelDevice) DeleteRoute(dst net.IPNet, table int) error {
	if table != 0 {
		return errors.ErrNotSupported
	}

	return exec.Command("route", "delete", "-net", dst.String(), "-interface", d.Name()).Run()
}
