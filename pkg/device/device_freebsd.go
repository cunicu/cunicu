package device

import (
	"fmt"
	"net"
	"os/exec"
)

func (d *BSDKernelDevice) AddRoute(dst net.IPNet, table int) error {
	return exec.Command("setfib", fmt.Sprint(table), "route", "add", "-net", dst.String(), "-interface", d.Name()).Run()
}

func (d *BSDKernelDevice) DeleteRoute(dst net.IPNet, table int) error {
	return exec.Command("setfib", fmt.Sprint(table), "route", "delete", "-net", dst.String(), "-interface", d.Name()).Run()
}
