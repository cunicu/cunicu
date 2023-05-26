// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package epdisc

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
)

func (i *Interface) setupNAT() error {
	var err error

	ident := fmt.Sprintf("cunicu-if%d", i.Device.Index())
	if i.nat, err = NewNAT(ident); err != nil {
		if !errors.Is(err, errNotSupported) {
			return err
		}
	}

	ports := []int{}

	// Redirect non-STUN traffic directed at UDP muxes to WireGuard interface via in-kernel port redirect / NAT
	if i.mux != nil {
		if i.natRule, err = i.nat.RedirectNonSTUN(i.muxPort, i.ListenPort); err != nil {
			return fmt.Errorf("failed to setup port redirect for server reflexive UDP mux: %w", err)
		}
		ports = append(ports, i.muxPort)
	}

	if i.muxSrflx != nil {
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
