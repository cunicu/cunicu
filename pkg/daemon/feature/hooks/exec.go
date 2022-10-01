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

	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/daemon/feature/epdisc"
	"github.com/stv0g/cunicu/pkg/wg"

	icex "github.com/stv0g/cunicu/pkg/ice"
)

type ExecHook struct {
	*config.ExecHookSetting

	logger *zap.Logger
}

func (h *Interface) NewExecHook(cfg *config.ExecHookSetting) *ExecHook {
	hk := &ExecHook{
		ExecHookSetting: cfg,
		logger: h.logger.Named("exec").With(
			zap.String("command", cfg.Command),
		),
	}

	h.logger.Debug("Created new exec hook", zap.Any("hook", hk))

	return hk
}

func (h *ExecHook) run(msg proto.Message, args ...any) {
	allArgs := []string{}
	allArgs = append(allArgs, h.Args...)

	for _, arg := range args {
		allArgs = append(allArgs, fmt.Sprintf("%v", arg))
	}

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
		go h.run(im, "modified", "interface", i.Name(), "type", i.Type, old.Type)
	}

	if m.Is(core.InterfaceModifiedPrivateKey) {
		go h.run(im, "modified", "interface", i.Name(), "private-key", i.PrivateKey(), old.PrivateKey)
	}

	if m.Is(core.InterfaceModifiedListenPort) {
		go h.run(im, "modified", "interface", i.Name(), "listen-port", i.ListenPort, old.ListenPort)
	}

	if m.Is(core.InterfaceModifiedFirewallMark) {
		go h.run(im, "modified", "interface", i.Name(), "fwmark", i.FirewallMark, old.FirewallMark)
	}

	if m.Is(core.InterfaceModifiedPeers) {
		go h.run(im, "modified", "interface", i.Name(), "peers")
	}
}

func (h *ExecHook) OnPeerAdded(p *core.Peer) {
	go h.run(p.Marshal(), "added", "peer", p.Interface.Name(), p.PublicKey())
}

func (h *ExecHook) OnPeerRemoved(p *core.Peer) {
	go h.run(p.Marshal(), "removed", "peer", p.Interface.Name(), p.PublicKey())
}

func (h *ExecHook) OnPeerModified(p *core.Peer, old *wgtypes.Peer, m core.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
	pm := p.Marshal()

	if m.Is(core.PeerModifiedPresharedKey) {
		go h.run(pm, "modified", "peer", p.Interface.Name(), p.PublicKey(), "preshared-key", p.PresharedKey(), old.PresharedKey)
	}

	if m.Is(core.PeerModifiedEndpoint) {
		go h.run(pm, "modified", "peer", p.Interface.Name(), p.PublicKey(), "endpoint", p.Endpoint, old.Endpoint)
	}

	if m.Is(core.PeerModifiedKeepaliveInterval) {
		go h.run(pm, "modified", "peer", p.Interface.Name(), p.PublicKey(), "presistent-keepalive", p.PersistentKeepaliveInterval.Seconds(), old.PersistentKeepaliveInterval.Seconds())
	}

	if m.Is(core.PeerModifiedHandshakeTime) {
		new := fmt.Sprint(p.LastHandshakeTime.UnixMilli())
		old := fmt.Sprint(old.LastHandshakeTime.UnixMilli())

		go h.run(pm, "modified", "peer", p.Interface.Name(), p.PublicKey(), "last-handshake", new, old)
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

		go h.run(pm, "modified", "peer", p.Interface.Name(), p.PublicKey(), "allowed-ips", strings.Join(added, ","), strings.Join(removed, ","))
	}

	if m.Is(core.PeerModifiedProtocolVersion) {
		go h.run(pm, "modified", "peer", p.Interface.Name(), p.PublicKey(), "protocol-version", p.ProtocolVersion, old.ProtocolVersion)
	}

	if m.Is(core.PeerModifiedName) {
		go h.run(pm, "modified", "peer", p.Interface.Name(), p.PublicKey(), "name", p.Name)
	}
}

func (h *ExecHook) OnConnectionStateChange(p *epdisc.Peer, new, prev icex.ConnectionState) {
	m := p.Peer.Marshal()
	m.Ice = p.Marshal()

	go h.run(m, "changed", "peer", "connection-state", p.Interface.Name(), p.PublicKey(), new, prev)
}
