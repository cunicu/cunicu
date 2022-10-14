package device

import (
	"fmt"
	"net"
	"os/exec"

	"go.uber.org/zap"
)

func (d *BSDKernelDevice) AddRoute(dst net.IPNet, gw net.IP, table int) error {
	d.logger.Debug("Add route",
		zap.String("dst", dst.String()),
		zap.String("gw", gw.String()))

	if gw == nil {
		return exec.Command("route", "add", "-net", dst.String(), "-interface", d.Name(), "-fib", fmt.Sprint(table)).Run()
	} else {
		return exec.Command("route", "add", "-net", dst.String(), gw.String(), "-fib", fmt.Sprint(table)).Run()
	}
}

func (d *BSDKernelDevice) DeleteRoute(dst net.IPNet, table int) error {
	d.logger.Debug("Delete route",
		zap.String("dst", dst.String()))

	return exec.Command("route", "delete", "-net", dst.String(), "-interface", d.Name(), "-fib", fmt.Sprint(table)).Run()
}
