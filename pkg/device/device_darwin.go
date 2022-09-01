package device

import (
	"net"
	"os/exec"

	"riasc.eu/wice/pkg/errors"
)

func (d *BSDKernelDevice) AddRoute(dst *net.IPNet, table int) error {
	if table != 0 {
		return errors.ErrNotSupported
	}

	return exec.Command("route", "add", "-net", dst.String(), "-interface", d.Name()).Run()
}

func (d *BSDKernelDevice) DeleteRoute(dst *net.IPNet, table int) error {
	if table != 0 {
		return errors.ErrNotSupported
	}

	return exec.Command("route", "delete", "-net", dst.String(), "-interface", d.Name()).Run()
}
