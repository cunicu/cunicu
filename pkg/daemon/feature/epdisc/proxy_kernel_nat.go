// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package epdisc

import (
	"errors"
	"fmt"
	"net"

	"github.com/pion/ice/v2"
	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/log"
)

const (
	StunMagicCookie uint32 = 0x2112A442
)

var errCandidatePairCanNotBeUsedWithNAT = errors.New("candidate pair can not be used with NAT")

// CandidatePairCanBeNATted checks if a given candidate pair
// can be used with kernel-space port address translation / natting.
func CandidatePairCanBeNATted(cp *ice.CandidatePair) bool {
	return cp.Local.Type() == ice.CandidateTypeHost || cp.Local.Type() == ice.CandidateTypeServerReflexive
}

// Compile-time assertions
var (
	_ Proxy = (*KernelNATProxy)(nil)
)

type KernelNATProxy struct {
	rule *NATRule

	logger *log.Logger
}

func NewKernelNATProxy(cp *ice.CandidatePair, nat *NAT, listenPort int, logger *log.Logger) (*KernelNATProxy, *net.UDPAddr, error) {
	var err error

	if !CandidatePairCanBeNATted(cp) {
		return nil, nil, errCandidatePairCanNotBeUsedWithNAT
	}

	epAddr := &net.UDPAddr{
		IP:   net.ParseIP(cp.Remote.Address()),
		Port: cp.Remote.Port(),
	}

	p := &KernelNATProxy{
		logger: logger.Named("proxy").With(zap.String("type", "kernel")),
	}

	// Setup source port masquerading (WireGuard listen-port -> STUN port)
	if p.rule, err = nat.MasqueradeSourcePort(listenPort, cp.Local.Port(), epAddr); err != nil {
		return nil, nil, err
	}

	return p, epAddr, nil
}

func (p *KernelNATProxy) Close() error {
	// Delete old source port masquerading rule
	if err := p.rule.Delete(); err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}

	return nil
}
