package hooks

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"strings"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"riasc.eu/wice/pkg/config"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/feat/epdisc"
	icex "riasc.eu/wice/pkg/feat/epdisc/ice"
	"riasc.eu/wice/pkg/wg"
)

type ExecHook struct {
	*config.ExecHookSetting

	logger *zap.Logger
}

func (h *ExecHook) run(msg proto.Message, args ...string) {
	allArgs := []string{}
	allArgs = append(allArgs, h.Args...)
	allArgs = append(allArgs, args...)

	cmd := exec.Command(h.Command, allArgs...)

	if msg != nil && h.Stdin {
		mo := protojson.MarshalOptions{
			Multiline:       true,
			Indent:          "  ",
			UseProtoNames:   true,
			EmitUnpopulated: false,
		}

		if buf, err := mo.Marshal(msg); err == nil {
			buf = append(buf, '\n')
			cmd.Stdin = bytes.NewReader(buf)
		}
	}

	for key, value := range h.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	if err := cmd.Run(); err != nil {
		h.logger.Error("Failed to invoke exec hook", zap.Error(err))
	}
}

func (h *ExecHook) OnInterfaceAdded(i *core.Interface) {
	go h.run(i.MarshalWithPeers(nil), "added", "interface", i.Name())
}

func (h *ExecHook) OnInterfaceRemoved(i *core.Interface) {
	go h.run(i.MarshalWithPeers(nil), "removed", "interface", i.Name())
}

func (h *ExecHook) OnInterfaceModified(i *core.Interface, old *wg.Device, m core.InterfaceModifier) {
	im := i.MarshalWithPeers(nil)

	if m.Is(core.InterfaceModifiedName) {
		go h.run(im, "modified", "interface", i.Name(), "name", i.Name(), old.Name)
	}

	if m.Is(core.InterfaceModifiedType) {
		go h.run(im, "modified", "interface", i.Name(), "type", i.Type.String(), old.Type.String())
	}

	if m.Is(core.InterfaceModifiedPrivateKey) {
		go h.run(im, "modified", "interface", i.Name(), "private-key", i.PrivateKey().String(), old.PrivateKey.String())
	}

	if m.Is(core.InterfaceModifiedListenPort) {
		new := fmt.Sprint(i.ListenPort)
		old := fmt.Sprint(old.ListenPort)

		go h.run(im, "modified", "interface", i.Name(), "listen-port", new, old)
	}

	if m.Is(core.InterfaceModifiedFirewallMark) {
		new := fmt.Sprint(i.FirewallMark)
		old := fmt.Sprint(old.FirewallMark)

		go h.run(im, "modified", "interface", i.Name(), "fwmark", new, old)
	}

	if m.Is(core.InterfaceModifiedPeers) {
		go h.run(im, "modified", "interface", i.Name(), "peers")
	}
}

func (h *ExecHook) OnPeerAdded(p *core.Peer) {
	go h.run(p.Marshal(), "added", "peer", p.Interface.Name(), p.PublicKey().String())
}

func (h *ExecHook) OnPeerRemoved(p *core.Peer) {
	go h.run(p.Marshal(), "removed", "peer", p.Interface.Name(), p.PublicKey().String())
}

func (h *ExecHook) OnPeerModified(p *core.Peer, old *wgtypes.Peer, m core.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
	pm := p.Marshal()

	if m.Is(core.PeerModifiedPresharedKey) {
		go h.run(pm, "modified", "peer", p.Interface.Name(), p.PublicKey().String(), "preshared-key", p.PresharedKey().String(), old.PresharedKey.String())
	}

	if m.Is(core.PeerModifiedEndpoint) {
		go h.run(pm, "modified", "peer", p.Interface.Name(), p.PublicKey().String(), "endpoint", p.Endpoint.String(), old.Endpoint.String())
	}

	if m.Is(core.PeerModifiedKeepaliveInterval) {
		new := fmt.Sprint(p.PersistentKeepaliveInterval.Seconds())
		old := fmt.Sprint(old.PersistentKeepaliveInterval.Seconds())

		go h.run(pm, "modified", "peer", p.Interface.Name(), p.PublicKey().String(), "presistent-keepalive", new, old)
	}

	if m.Is(core.PeerModifiedHandshakeTime) {
		new := fmt.Sprint(p.LastHandshakeTime.UnixMilli())
		old := fmt.Sprint(old.LastHandshakeTime.UnixMilli())

		go h.run(pm, "modified", "peer", p.Interface.Name(), p.PublicKey().String(), "last-handshake", new, old)
	}

	if m.Is(core.PeerModifiedAllowedIPs) {
		added := []string{}
		for _, ip := range ipsAdded {
			added = append(added, ip.String())
		}

		removed := []string{}
		for _, ip := range ipsRemoved {
			removed = append(removed, ip.String())
		}

		go h.run(pm, "modified", "peer", p.Interface.Name(), p.PublicKey().String(), "allowed-ips", strings.Join(added, ","), strings.Join(removed, ","))
	}

	if m.Is(core.PeerModifiedProtocolVersion) {
		new := fmt.Sprint(p.ProtocolVersion)
		old := fmt.Sprint(old.ProtocolVersion)

		go h.run(pm, "modified", "peer", p.Interface.Name(), p.PublicKey().String(), "protocol-version", new, old)
	}

	if m.Is(core.PeerModifiedName) {
		go h.run(pm, "modified", "peer", p.Interface.Name(), p.PublicKey().String(), "name", p.Name)
	}
}

func (h *ExecHook) OnConnectionStateChange(p *epdisc.Peer, new, prev icex.ConnectionState) {
	m := p.Peer.Marshal()
	m.Ice = p.Marshal()

	go h.run(m, "changed", "peer", "connection-state", p.Interface.Name(), p.PublicKey().String(), new.String(), prev.String())
}
