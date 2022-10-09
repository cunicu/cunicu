package device

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/stv0g/cunicu/pkg/errors"
)

func (d *BSDKernelDevice) AddRoute(dst net.IPNet, gw net.IP, table int) error {
	if table != 0 {
		return errors.ErrNotSupported
	}

	args := []string{"route", "add", fmt.Sprintf("-%s", addressFamily(dst)), "-net", dst.String()}
	if gw == nil {
		args = append(args, "-interface", d.Name())
	} else {
		args = append(args, gw.String())
	}

	if out, err := run(args...); err != nil {
		return fmt.Errorf("failed to run command '%s': %w: %s", strings.Join(args, " "), err, out)
	}

	return nil
}

func (d *BSDKernelDevice) DeleteRoute(dst net.IPNet, table int) error {
	if table != 0 {
		return errors.ErrNotSupported
	}

	return exec.Command("route", "delete", "-net", dst.String(), "-interface", d.Name()).Run()
}
