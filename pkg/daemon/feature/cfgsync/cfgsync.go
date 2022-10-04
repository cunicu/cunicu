package cfgsync

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"syscall"

	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/wg"
	"go.uber.org/zap"
)

func init() {
	daemon.Features["cfgsync"] = &daemon.FeaturePlugin{
		New:         New,
		Description: "Config synchronization",
		Order:       20,
	}
}

type Interface struct {
	*daemon.Interface

	logger *zap.Logger
}

func New(i *daemon.Interface) (daemon.Feature, error) {
	if !i.Settings.AutoConfig.Enabled {
		return nil, nil
	}

	c := &Interface{
		Interface: i,
		logger:    zap.L().Named("cfgsync").With(zap.String("intf", i.Name())),
	}

	return c, nil
}

func (cs *Interface) Start() error {
	cs.logger.Info("Started config synchronization")

	// Assign static addresses
	for _, addr := range cs.Settings.AutoConfig.Addresses {
		if err := cs.KernelDevice.AddAddress(addr); err != nil && !errors.Is(err, syscall.EEXIST) {
			cs.logger.Error("Failed to assign address", zap.Error(err), zap.Any("addr", addr))
		}
	}

	// Set MTU
	if mtu := cs.Settings.AutoConfig.MTU; mtu != 0 {
		if err := cs.KernelDevice.SetMTU(mtu); err != nil {
			cs.logger.Error("Failed to set MTU",
				zap.Error(err),
				zap.Int("mtu", mtu))
		}
	}

	// Set DNS
	if dns := cs.Settings.AutoConfig.DNS; len(dns) > 0 {
		if err := cs.SetDNS(cs.Settings.AutoConfig.DNS); err != nil {
			cs.logger.Error("Failed to set DNS servers",
				zap.Error(err),
				zap.Any("servers", dns))
		}
	}

	// Configure Wireguard interface
	// cfg := cs.Settings.WireGuard.Config()

	return nil
}

func (cs *Interface) Close() error {
	// Unset DNS
	if dns := cs.Settings.AutoConfig.DNS; len(dns) > 0 {
		if err := cs.UnsetDNS(); err != nil {
			cs.logger.Error("Failed to restore DNS servers", zap.Error(err))
		}
	}

	return nil
}

func (cs *Interface) Configure(cfg *wg.Config) error {
	if err := cs.ConfigureDevice(cfg.Config); err != nil {
		return fmt.Errorf("failed to synchronize interface configuration: %s", err)
	}

	return nil
}

func (i *Interface) SetDNS(svrs []net.IPAddr) error {
	var cmd *exec.Cmd

	// Check if SystemD's resolvectl is available
	if resolvectl, err := exec.LookPath("resolvectl"); err == nil {
		args := []string{"dns", i.Name()}
		for _, svr := range svrs {
			args = append(args, svr.String())
		}

		cmd = exec.Command(resolvectl, args...)
	} else if resolveconf, err := exec.LookPath("resolveconf"); err != nil {
		cmd := exec.Command(resolveconf, "-a", i.Name(), "-m", "0", "-x")

		stdin := &bytes.Buffer{}
		for _, svr := range svrs {
			fmt.Fprintf(stdin, "nameserver %s\n", svr.String())
		}

		cmd.Stdin = stdin
	}

	return cmd.Run()
}

func (i *Interface) UnsetDNS() error {
	var cmd *exec.Cmd

	// Check if SystemD's resolvectl is available
	if resolvectl, err := exec.LookPath("resolvectl"); err == nil {
		cmd = exec.Command(resolvectl, "revert", i.Name())
	} else if resolveconf, err := exec.LookPath("resolveconf"); err != nil {
		cmd = exec.Command(resolveconf, "-d", i.Name())
	}

	return cmd.Run()
}
