// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package epdisc

import (
	"errors"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/google/nftables"
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
	"golang.org/x/sys/unix"
)

var errRuleNotFound = errors.New("rule not found")

type NAT struct {
	conn *nftables.Conn

	table        *nftables.Table
	chainEgress  *nftables.Chain
	chainIngress *nftables.Chain

	lastID atomic.Uint32
}

func NewNAT(ident string) (*NAT, error) {
	var err error

	n := &NAT{}

	if n.conn, err = nftables.New(); err != nil {
		return nil, fmt.Errorf("failed to create netlink conn: %w", err)
	}

	if err := n.setupTable(ident); err != nil {
		return nil, fmt.Errorf("failed to setup table: %w", err)
	}

	return n, nil
}

type NATRule struct {
	*nftables.Rule
	nat *NAT
}

func (n *NAT) AddRule(r *nftables.Rule, comment string) (*NATRule, error) {
	id := n.lastID.Add(1)

	r.UserData = nil
	r.UserData = NftablesUserDataPutString(r.UserData, NftablesUserDataTypeComment, comment)
	r.UserData = NftablesUserDataPutInt(r.UserData, NftablesUserDataTypeRuleID, id)

	r = n.conn.AddRule(r)

	if err := n.conn.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush: %w", err)
	}

	rs, err := n.conn.GetRules(r.Table, r.Chain)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules: %w", err)
	}

	for _, nr := range rs {
		rid, ok := NftablesUserDataGetInt(nr.UserData, NftablesUserDataTypeRuleID)
		if ok && rid == id {
			r.Handle = nr.Handle
			break
		}
	}

	if r.Handle == 0 {
		return nil, errRuleNotFound
	}

	return &NATRule{
		Rule: r,
		nat:  n,
	}, nil
}

func (nr *NATRule) Delete() error {
	return nr.nat.conn.DelRule(nr.Rule)
}

func (n *NAT) setupTable(ident string) error {
	// Ignore any previously existing table
	n.conn.DelTable(&nftables.Table{Name: ident})

	n.conn.Flush()

	n.table = n.conn.AddTable(&nftables.Table{
		Name:   ident,
		Family: nftables.TableFamilyINet,
	})

	// Ingress
	n.chainIngress = n.conn.AddChain(&nftables.Chain{
		Name:     "ingress",
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookInput,
		Priority: nftables.ChainPriorityRaw,
		Table:    n.table,
	})

	// Egress
	n.chainEgress = n.conn.AddChain(&nftables.Chain{
		Name:     "egress",
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookOutput,
		Priority: nftables.ChainPriorityRaw,
		Table:    n.table,
	})

	return n.conn.Flush()
}

func (n *NAT) Close() error {
	n.conn.DelTable(n.table)

	if err := n.conn.Flush(); err != nil && !errors.Is(err, unix.ENOENT) {
		return err
	}

	return nil
}

// RedirectNonSTUN redirects non-STUN UDP ingress traffic directed at port 'toPort' to port 'toPort'.
func (n *NAT) RedirectNonSTUN(origPort, newPort int) (*NATRule, error) {
	r := &nftables.Rule{
		Table: n.table,
		Chain: n.chainIngress,
		Exprs: []expr.Any{
			// meta l4proto udp
			&expr.Meta{
				Key:      expr.MetaKeyL4PROTO,
				Register: 1,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     []byte{unix.IPPROTO_UDP},
			},

			// udp dport origPort
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseTransportHeader,
				Offset:       2,
				Len:          2,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     binaryutil.BigEndian.PutUint16(uint16(origPort)),
			},

			// @th,96,32 != StunMagicCookie
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseTransportHeader,
				Offset:       12,
				Len:          4,
			},
			&expr.Cmp{
				Op:       expr.CmpOpNeq,
				Register: 1,
				Data:     binaryutil.BigEndian.PutUint32(StunMagicCookie),
			},

			&expr.Notrack{},

			// udp dport set newPort
			&expr.Immediate{
				Register: 1,
				Data:     binaryutil.BigEndian.PutUint16(uint16(newPort)),
			},
			&expr.Payload{
				OperationType:  expr.PayloadWrite,
				SourceRegister: 1,
				Base:           expr.PayloadBaseTransportHeader,
				Offset:         2,
				Len:            2,
			},
		},
	}

	comment := fmt.Sprintf("UDP destination port forwarding of non-STUN traffic from port %d to port %d", origPort, newPort)
	return n.AddRule(r, comment)
}

// Perform SNAT to the source port of WireGuard UDP traffic to match port of our local ICE candidate
func (n *NAT) MasqueradeSourcePort(fromPort, toPort int, dest *net.UDPAddr) (*NATRule, error) {
	var destIP []byte
	var destIPOffset, destIPLength uint32

	isIPv6 := dest.IP.To4() == nil
	if isIPv6 {
		destIP = dest.IP.To16()
		destIPOffset = 24
		destIPLength = net.IPv6len
	} else {
		destIP = dest.IP.To4()
		destIPOffset = 16
		destIPLength = net.IPv4len
	}

	r := &nftables.Rule{
		Table: n.table,
		Chain: n.chainEgress,
		Exprs: []expr.Any{
			// meta l4proto udp
			&expr.Meta{
				Key:      expr.MetaKeyL4PROTO,
				Register: 1,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     []byte{unix.IPPROTO_UDP},
			},

			// udp sport fromPort
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseTransportHeader,
				Offset:       0,
				Len:          2,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     binaryutil.BigEndian.PutUint16(uint16(fromPort)),
			},

			// udp dst dest.IP
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseNetworkHeader,
				Offset:       destIPOffset,
				Len:          destIPLength,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     destIP,
			},

			// udp dport dest.Port
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseTransportHeader,
				Offset:       2,
				Len:          2,
			},
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     binaryutil.BigEndian.PutUint16(uint16(dest.Port)),
			},

			// notrack
			&expr.Notrack{},

			// udp sport set toPort
			&expr.Immediate{
				Register: 1,
				Data:     binaryutil.BigEndian.PutUint16(uint16(toPort)),
			},
			&expr.Payload{
				OperationType:  expr.PayloadWrite,
				SourceRegister: 1,
				Base:           expr.PayloadBaseTransportHeader,
				Offset:         0,
				Len:            2,
			},
		},
	}

	comment := fmt.Sprintf("UDP source port masquerade from port %d to %d for destination %s", fromPort, toPort, dest)
	return n.AddRule(r, comment)
}
