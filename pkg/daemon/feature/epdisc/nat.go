// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package epdisc

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
)

func (i *Interface) setupNAT() (err error) {
	ident := fmt.Sprintf("cunicu-if%d", i.Device.Index())
	if i.nat, err = NewNAT(ident); err != nil {
		if !errors.Is(err, errNotSupported) {
			return err
		}
	}

	return i.updateNATRules()
}

func (i *Interface) updateNATRules() (err error) {
	if i.ListenPort == 0 {
		i.logger.Debug("Skipping setup of NAT rules as interface has no listen port yet")
		return nil
	}

	ports := []int{}

	// Redirect non-STUN traffic directed at UDP muxes to WireGuard interface via in-kernel port redirect / NAT
	if i.mux != nil {
		if i.natRule != nil {
			if err := i.natRule.Delete(); err != nil {
				return fmt.Errorf("failed to remove port redirect for hots UDP mux: %w", err)
			}
		}

		if i.natRule, err = i.nat.RedirectNonSTUN(i.muxPort, i.ListenPort); err != nil {
			return fmt.Errorf("failed to setup port redirect for host UDP mux: %w", err)
		}

		ports = append(ports, i.muxPort)
	}

	if i.muxSrflx != nil {
		if i.natRuleSrflx != nil {
			if err := i.natRuleSrflx.Delete(); err != nil {
				return fmt.Errorf("failed to remove port redirect for server reflexive UDP mux: %w", err)
			}
		}

		if i.natRuleSrflx, err = i.nat.RedirectNonSTUN(i.muxSrflxPort, i.ListenPort); err != nil {
			return fmt.Errorf("failed to setup port redirect for server reflexive UDP mux: %w", err)
		}

		ports = append(ports, i.muxSrflxPort)
	}

	if len(ports) > 0 {
		i.logger.Info("Setup NAT redirects for WireGuard traffic",
			zap.Ints("sports", ports),
			zap.Int("dport", i.ListenPort))
	}

	return nil
}
