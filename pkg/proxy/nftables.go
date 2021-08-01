// +build linux

package proxy

import (
	"fmt"
	"net"
	"runtime"

	"github.com/google/nftables"
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
)

type NFTablesProxy struct {
	BaseProxy

	logger log.FieldLogger

	NFConn *nftables.Conn
	Conn   net.Conn
}

func CheckNFTablesSupport() bool {
	return runtime.GOOS == "linux"
}

func NewNFTablesProxy(ident string, listenPort int, cb UpdateEndpointCb, conn net.Conn) (*NFTablesProxy, error) {
	ns, err := netns.Get()
	if err != nil {
		return nil, err
	}

	proxy := &NFTablesProxy{
		BaseProxy: BaseProxy{
			Ident:      ident,
			ListenPort: listenPort,
		},
		logger: log.WithFields(log.Fields{
			"logger": "proxy",
			"type":   "nftables",
		}),
		NFConn: &nftables.Conn{
			NetNS: int(ns),
		},
		Conn: conn,
	}

	proxy.logger.Infof("Network namespace: %s", ns)

	proxy.setupTable()

	// Update Wireguard peer endpoint
	rAddr := proxy.Conn.RemoteAddr().(*net.UDPAddr)
	if err = cb(rAddr); err != nil {
		return nil, err
	}

	proxy.logger.Info("Configured stateless nftables port redirection")

	return proxy, nil
}

func (p *NFTablesProxy) deleteTable() error {
	tb := nftables.Table{
		Name:   "wice",
		Family: nftables.TableFamilyINet,
	}
	p.NFConn.DelTable(&tb) // Delete any previous existing table
	p.NFConn.Flush()       // We dont care about errors here...

	return nil
}

func (p *NFTablesProxy) tableName() string {
	// pkSlug := base32.HexEncoding.EncodeToString(p.WGPeer.PublicKey[:16])
	return fmt.Sprintf("wice-%s", p.Ident)
}

func (p *NFTablesProxy) setupTable() error {
	// Delete any stale tables created by WICE
	p.deleteTable()

	lAddr := p.Conn.LocalAddr().(*net.UDPAddr)
	rAddr := p.Conn.RemoteAddr().(*net.UDPAddr)

	tb := nftables.Table{
		Name:   p.tableName(),
		Family: nftables.TableFamilyINet,
	}
	p.NFConn.AddTable(&tb)

	// Ingress
	chIngress := nftables.Chain{
		Name:     "ingress",
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookInput,
		Priority: nftables.ChainPriorityRaw,
		Table:    &tb,
	}
	p.NFConn.AddChain(&chIngress)

	// Match non-STUN UDP ingress traffic directed at the port of our local ICE candidate
	// and redirect to the listen port of the Wireguard interface.
	// STUN traffic will pass to the iceConn for keepalives and connection checks.
	rDnat := nftables.Rule{
		Table: &tb,
		Chain: &chIngress,
	}
	p.NFConn.AddChain(&chIngress)

	// meta l4proto udp
	rDnat.Exprs = append(rDnat.Exprs, &expr.Meta{
		Key:      expr.MetaKeyL4PROTO,
		Register: 1,
	})
	rDnat.Exprs = append(rDnat.Exprs, &expr.Cmp{
		Op:       expr.CmpOpEq,
		Register: 1,
		Data:     []byte{unix.IPPROTO_UDP},
	})

	// udp dport lAddr.Port
	rDnat.Exprs = append(rDnat.Exprs, &expr.Payload{
		DestRegister: 1,
		Base:         expr.PayloadBaseTransportHeader,
		Offset:       2,
		Len:          2,
	})
	rDnat.Exprs = append(rDnat.Exprs, &expr.Cmp{
		Op:       expr.CmpOpEq,
		Register: 1,
		Data:     binaryutil.BigEndian.PutUint16(uint16(lAddr.Port)),
	})

	// @th,96,32 != StunMagicCookie
	rDnat.Exprs = append(rDnat.Exprs, &expr.Payload{
		DestRegister: 1,
		Base:         expr.PayloadBaseTransportHeader,
		Offset:       12,
		Len:          4,
	})
	rDnat.Exprs = append(rDnat.Exprs, &expr.Cmp{
		Op:       expr.CmpOpNeq,
		Register: 1,
		Data:     binaryutil.BigEndian.PutUint32(StunMagicCookie),
	})

	// notrack
	rDnat.Exprs = append(rDnat.Exprs, &expr.Notrack{})

	// udp dport set p.Device.ListenPort
	rDnat.Exprs = append(rDnat.Exprs, &expr.Immediate{
		Register: 1,
		Data:     binaryutil.BigEndian.PutUint16(uint16(p.ListenPort)),
	})
	rDnat.Exprs = append(rDnat.Exprs, &expr.Payload{
		OperationType:  expr.PayloadWrite,
		SourceRegister: 1,
		Base:           expr.PayloadBaseTransportHeader,
		Offset:         2,
		Len:            2,
	})

	p.NFConn.AddRule(&rDnat)

	// Egress
	chEgress := nftables.Chain{
		Name:     "egress",
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookOutput,
		Priority: nftables.ChainPriorityRaw,
		Table:    &tb,
	}
	p.NFConn.AddChain(&chEgress)

	// Perform SNAT to the source port of Wireguard UDP traffic to match port of our local ICE candidate
	rSnat := nftables.Rule{
		Table: &tb,
		Chain: &chEgress,
	}

	// meta l4proto udp
	rSnat.Exprs = append(rSnat.Exprs, &expr.Meta{
		Key:      expr.MetaKeyL4PROTO,
		Register: 1,
	})
	rSnat.Exprs = append(rSnat.Exprs, &expr.Cmp{
		Op:       expr.CmpOpEq,
		Register: 1,
		Data:     []byte{unix.IPPROTO_UDP},
	})

	// udp sport p.ListenPort
	rSnat.Exprs = append(rSnat.Exprs, &expr.Payload{
		DestRegister: 1,
		Base:         expr.PayloadBaseTransportHeader,
		Offset:       0,
		Len:          2,
	})
	rSnat.Exprs = append(rSnat.Exprs, &expr.Cmp{
		Op:       expr.CmpOpEq,
		Register: 1,
		Data:     binaryutil.BigEndian.PutUint16(uint16(p.ListenPort)),
	})

	// udp dport rAddr.Port
	rSnat.Exprs = append(rSnat.Exprs, &expr.Payload{
		DestRegister: 1,
		Base:         expr.PayloadBaseTransportHeader,
		Offset:       2,
		Len:          2,
	})
	rSnat.Exprs = append(rSnat.Exprs, &expr.Cmp{
		Op:       expr.CmpOpEq,
		Register: 1,
		Data:     binaryutil.BigEndian.PutUint16(uint16(rAddr.Port)),
	})

	// notrack
	rSnat.Exprs = append(rSnat.Exprs, &expr.Notrack{})

	// udp sport set lAddr.Port
	rSnat.Exprs = append(rSnat.Exprs, &expr.Immediate{
		Register: 1,
		Data:     binaryutil.BigEndian.PutUint16(uint16(lAddr.Port)),
	})
	rSnat.Exprs = append(rSnat.Exprs, &expr.Payload{
		OperationType:  expr.PayloadWrite,
		SourceRegister: 1,
		Base:           expr.PayloadBaseTransportHeader,
		Offset:         0,
		Len:            2,
	})

	p.NFConn.AddRule(&rSnat)

	if err := p.NFConn.Flush(); err != nil {
		return fmt.Errorf("failed setup nftables: %w", err)
	}

	return nil
}

func (p *NFTablesProxy) Close() error {
	return p.deleteTable()
}

func (bpf *NFTablesProxy) UpdateEndpoint(addr *net.UDPAddr) error {
	return nil
}
