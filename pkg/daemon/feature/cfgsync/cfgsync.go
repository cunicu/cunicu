// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package cfgsync

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"syscall"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/log"
)

var Get = daemon.RegisterFeature(New, 20) //nolint:gochecknoglobals

type Interface struct {
	*daemon.Interface

	logger *log.Logger
}

func New(i *daemon.Interface) (*Interface, error) {
	c := &Interface{
		Interface: i,
		logger:    log.Global.Named("cfgsync").With(zap.String("intf", i.Name())),
	}

	return c, nil
}

func (i *Interface) Start() error {
	i.logger.Info("Started config synchronization")

	// Assign static addresses
	for _, addr := range i.Settings.Addresses {
		if err := i.Device.AddAddress(addr); err != nil && !errors.Is(err, syscall.EEXIST) {
			return fmt.Errorf("failed to assign address '%s': %w", addr.String(), err)
		}
	}

	// Set MTU
	if mtu := i.Settings.MTU; mtu != 0 {
		if err := i.Device.SetMTU(mtu); err != nil {
			return fmt.Errorf("failed to set MTU: %w", err)
		}
	}

	// Set DNS
	if dns := i.Settings.DNS; len(dns) > 0 {
		var domain []string
		if i.Settings.Domain != "" {
			domain = append(domain, i.Settings.Domain)
		}

		if err := i.SetDNS(i.Settings.DNS, domain); err != nil {
			return fmt.Errorf("failed to set DNS servers: %w", err)
		}
	}

	// Set WireGuard settings
	if err := i.ConfigureWireGuard(); err != nil {
		return fmt.Errorf("failed to configure WireGuard interface: %w", err)
	}

	return nil
}

func (i *Interface) Close() error {
	// Unset DNS
	if dns := i.Settings.DNS; len(dns) > 0 {
		if err := i.UnsetDNS(); err != nil {
			i.logger.Error("Failed to restore DNS servers", zap.Error(err))
		}
	}

	return nil
}

func (i *Interface) ConfigureWireGuard() error { //nolint:gocognit
	cfg := wgtypes.Config{}

	if i.Settings.FirewallMark != 0 && i.Settings.FirewallMark != i.FirewallMark {
		cfg.FirewallMark = &i.Settings.FirewallMark
	}

	if i.Settings.ListenPort != nil && *i.Settings.ListenPort != i.ListenPort {
		cfg.ListenPort = i.Settings.ListenPort
	}

	if i.Settings.PrivateKey.IsSet() && i.Settings.PrivateKey != i.PrivateKey() {
		cfg.PrivateKey = (*wgtypes.Key)(&i.Settings.PrivateKey)
	}

	for _, p := range i.Settings.Peers {
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
		if err := i.ConfigureDevice(cfg); err != nil {
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
