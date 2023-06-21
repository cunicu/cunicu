// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

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
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/daemon/feature/epdisc"
	"github.com/stv0g/cunicu/pkg/log"
	"github.com/stv0g/cunicu/pkg/wg"
)

type ExecHook struct {
	*config.ExecHookSetting

	logger *log.Logger
}

func (i *Interface) NewExecHook(cfg *config.ExecHookSetting) *ExecHook {
	hk := &ExecHook{
		ExecHookSetting: cfg,
		logger: i.logger.Named("exec").With(
			zap.String("command", cfg.Command),
		),
	}

	i.logger.Debug("Created new exec hook", zap.Any("hook", hk))

	return hk
}

func (h *ExecHook) run(msg proto.Message, args ...any) {
	allArgs := []string{}
	allArgs = append(allArgs, h.Args...)

	for _, arg := range args {
		allArgs = append(allArgs, fmt.Sprintf("%v", arg))
	}

	// It the main purpose of an exec hook to run arbitrary external executables
	cmd := exec.Command(h.Command, allArgs...) //nolint:gosec

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

func (h *ExecHook) OnInterfaceAdded(i *daemon.Interface) {
	go h.run(i.MarshalWithPeers(nil), "added", "interface", i.Name())
}

func (h *ExecHook) OnInterfaceRemoved(i *daemon.Interface) {
	go h.run(i.MarshalWithPeers(nil), "removed", "interface", i.Name())
}

func (h *ExecHook) OnInterfaceModified(i *daemon.Interface, oldIntf *wg.Interface, m daemon.InterfaceModifier) {
	im := i.MarshalWithPeers(nil)

	newIntf := i.Interface

	if m.Is(daemon.InterfaceModifiedName) {
		go h.run(im, "modified", "interface", i.Name(), "name", newIntf.Name, oldIntf.Name)
	}

	if m.Is(daemon.InterfaceModifiedType) {
		go h.run(im, "modified", "interface", i.Name(), "type", newIntf.Type, oldIntf.Type)
	}

	if m.Is(daemon.InterfaceModifiedPrivateKey) {
		go h.run(im, "modified", "interface", i.Name(), "private-key", newIntf.PrivateKey, oldIntf.PrivateKey)
	}

	if m.Is(daemon.InterfaceModifiedListenPort) {
		go h.run(im, "modified", "interface", i.Name(), "listen-port", newIntf.ListenPort, oldIntf.ListenPort)
	}

	if m.Is(daemon.InterfaceModifiedFirewallMark) {
		go h.run(im, "modified", "interface", i.Name(), "fwmark", newIntf.FirewallMark, oldIntf.FirewallMark)
	}

	if m.Is(daemon.InterfaceModifiedPeers) {
		go h.run(im, "modified", "interface", i.Name(), "peers")
	}
}

func (h *ExecHook) OnPeerAdded(p *daemon.Peer) {
	go h.run(p.Marshal(), "added", "peer", p.Interface.Name(), p.PublicKey())
}

func (h *ExecHook) OnPeerRemoved(p *daemon.Peer) {
	go h.run(p.Marshal(), "removed", "peer", p.Interface.Name(), p.PublicKey())
}

func (h *ExecHook) OnPeerModified(p *daemon.Peer, old *wgtypes.Peer, m daemon.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
	pm := p.Marshal()

	if m.Is(daemon.PeerModifiedPresharedKey) {
		go h.run(pm, "modified", "peer", p.Interface.Name(), p.PublicKey(), "preshared-key", p.PresharedKey(), old.PresharedKey)
	}

	if m.Is(daemon.PeerModifiedEndpoint) {
		go h.run(pm, "modified", "peer", p.Interface.Name(), p.PublicKey(), "endpoint", p.Endpoint, old.Endpoint)
	}

	if m.Is(daemon.PeerModifiedKeepaliveInterval) {
		go h.run(pm, "modified", "peer", p.Interface.Name(), p.PublicKey(), "persistent-keepalive", p.PersistentKeepaliveInterval.Seconds(), old.PersistentKeepaliveInterval.Seconds())
	}

	if m.Is(daemon.PeerModifiedHandshakeTime) {
		newTime := fmt.Sprint(p.LastHandshakeTime.UnixMilli())
		oldTime := fmt.Sprint(old.LastHandshakeTime.UnixMilli())

		go h.run(pm, "modified", "peer", p.Interface.Name(), p.PublicKey(), "last-handshake", newTime, oldTime)
	}

	if m.Is(daemon.PeerModifiedAllowedIPs) {
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

	if m.Is(daemon.PeerModifiedProtocolVersion) {
		go h.run(pm, "modified", "peer", p.Interface.Name(), p.PublicKey(), "protocol-version", p.ProtocolVersion, old.ProtocolVersion)
	}

	if m.Is(daemon.PeerModifiedName) {
		go h.run(pm, "modified", "peer", p.Interface.Name(), p.PublicKey(), "name", p.Name)
	}
}

func (h *ExecHook) OnPeerStateChanged(p *daemon.Peer, newState, prevState daemon.PeerState) {
	pm := p.Marshal().Redact()

	if epi := epdisc.Get(p.Interface); epi != nil {
		epp := epi.Peers[p]
		pm.Ice = epp.Marshal()
	}

	go h.run(pm, "changed", "peer", "connection-state", p.Interface.Name(), p.PublicKey(), newState, prevState)
}
