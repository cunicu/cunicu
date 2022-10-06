package cfgsync

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"syscall"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
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
	c := &Interface{
		Interface: i,
		logger:    zap.L().Named("cfgsync").With(zap.String("intf", i.Name())),
	}

	return c, nil
}

func (cs *Interface) Start() error {
	cs.logger.Info("Started config synchronization")

	// Assign static addresses
	for _, addr := range cs.Settings.Addresses {
		if err := cs.KernelDevice.AddAddress(net.IPNet(addr)); err != nil && !errors.Is(err, syscall.EEXIST) {
			return fmt.Errorf("failed to assign address '%s': %w", addr.String(), err)
		}
	}

	// Set MTU
	if mtu := cs.Settings.MTU; mtu != 0 {
		if err := cs.KernelDevice.SetMTU(mtu); err != nil {
			return fmt.Errorf("failed to set MTU: %w", err)
		}
	}

	// Set DNS
	if dns := cs.Settings.DNS; len(dns) > 0 {
		var domain []string
		if cs.Settings.Domain != "" {
			domain = append(domain, cs.Settings.Domain)
		}

		if err := cs.SetDNS(cs.Settings.DNS, domain); err != nil {
			return fmt.Errorf("failed to set DNS servers: %w", err)
		}
	}

	// Set WireGuard settings
	if err := cs.ConfigureWireGuard(); err != nil {
		return fmt.Errorf("failed to configure WireGuard interface: %w", err)
	}

	return nil
}

func (cs *Interface) Close() error {
	// Unset DNS
	if dns := cs.Settings.DNS; len(dns) > 0 {
		if err := cs.UnsetDNS(); err != nil {
			cs.logger.Error("Failed to restore DNS servers", zap.Error(err))
		}
	}

	return nil
}

func (cs *Interface) ConfigureWireGuard() error {
	cfg := wgtypes.Config{}

	if cs.Settings.FirewallMark != 0 && cs.Settings.FirewallMark != cs.FirewallMark {
		cfg.FirewallMark = &cs.Settings.FirewallMark
	}

	if cs.Settings.ListenPort != nil && *cs.Settings.ListenPort != cs.ListenPort {
		cfg.ListenPort = cs.Settings.ListenPort
	}

	if cs.Settings.PrivateKey.IsSet() && cs.Settings.PrivateKey != cs.PrivateKey() {
		cfg.PrivateKey = (*wgtypes.Key)(&cs.Settings.PrivateKey)
	}

	for _, p := range cs.Settings.Peers {
		pcfg := wgtypes.PeerConfig{}

		pcfg.PublicKey = wgtypes.Key(p.PublicKey)

		if p.Endpoint != "" {
			addr, err := net.ResolveUDPAddr("udp", p.Endpoint)
			if err != nil {
				return fmt.Errorf("failed to resolve peer endpoint address '%s': %w", p.Endpoint, err)
			}

			pcfg.Endpoint = addr
		}

		if p.PersistentKeepaliveInterval > 0 {
			pcfg.PersistentKeepaliveInterval = &p.PersistentKeepaliveInterval
		}

		if psk := p.PresharedKey; psk.IsSet() {
			pcfg.PresharedKey = (*wgtypes.Key)(&psk)
		} else if psk := crypto.Key(p.PresharedKeyPassphrase); psk.IsSet() {
			pcfg.PresharedKey = (*wgtypes.Key)(&psk)
		}

		if len(p.AllowedIPs) > 0 {
			pcfg.AllowedIPs = p.AllowedIPs
		}

		cfg.Peers = append(cfg.Peers, pcfg)
	}

	if cfg.FirewallMark != nil || cfg.ListenPort != nil || cfg.PrivateKey != nil || cfg.Peers != nil {
		if err := cs.ConfigureDevice(cfg); err != nil {
			return fmt.Errorf("failed to configure WireGuard interface: %w", err)
		}
	}

	return nil
}

func (i *Interface) SetDNS(svrs []net.IPAddr, domain []string) error {
	var cmd *exec.Cmd

	// Check if SystemD's resolvectl is available
	if resolvectl, err := exec.LookPath("resolvectl"); err == nil {
		if len(svrs) > 0 {
			args := []string{"dns", i.Name()}
			for _, svr := range svrs {
				args = append(args, svr.String())
			}

			cmd = exec.Command(resolvectl, args...)

			if err := cmd.Run(); err != nil {
				return err
			}
		}

		// Set DNS search domains
		if len(domain) > 0 {
			args := []string{"domain", i.Name()}
			args = append(args, domain...)

			cmd = exec.Command(resolvectl, args...)

			if err := cmd.Run(); err != nil {
				return err
			}
		}
	} else if resolveconf, err := exec.LookPath("resolveconf"); err != nil {
		if len(svrs) > 0 || len(domain) > 0 {
			cmd := exec.Command(resolveconf, "-a", i.Name(), "-m", "0", "-x")

			stdin := &bytes.Buffer{}

			for _, svr := range svrs {
				fmt.Fprintf(stdin, "nameserver %s\n", svr.String())
			}

			if len(domain) > 0 {
				fmt.Fprintf(stdin, "search %s\n", strings.Join(domain, " "))
			}

			cmd.Stdin = stdin

			if err := cmd.Run(); err != nil {
				return err
			}
		}
	}

	return nil
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
