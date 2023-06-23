// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package hooks

import (
	"bytes"
	"io"
	"net"
	"net/http"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/stv0g/cunicu/pkg/buildinfo"
	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/daemon/feature/epdisc"
	"github.com/stv0g/cunicu/pkg/log"
	hooksproto "github.com/stv0g/cunicu/pkg/proto/feature/hooks"
	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
	"github.com/stv0g/cunicu/pkg/wg"
)

type WebHook struct {
	*config.WebHookSetting

	logger *log.Logger
}

func (i *Interface) NewWebHook(cfg *config.WebHookSetting) *WebHook {
	hk := &WebHook{
		WebHookSetting: cfg,
		logger: i.logger.Named("web").With(
			zap.Any("url", cfg.URL),
		),
	}

	i.logger.Debug("Created new web hook", zap.Any("hook", hk))

	return hk
}

func (h *WebHook) run(msg proto.Message) {
	req := &http.Request{
		Method: h.Method,
		URL:    &h.URL.URL,
		Header: http.Header{},
	}

	req.Header.Set("User-Agent", buildinfo.UserAgent())
	req.Header.Set("Content-Type", "application/json")

	for key, value := range h.Headers {
		req.Header.Set(key, value)
	}

	mo := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		UseProtoNames:   true,
		EmitUnpopulated: false,
	}

	if buf, err := mo.Marshal(msg); err == nil {
		req.Body = io.NopCloser(bytes.NewReader(buf))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		h.logger.Error("Failed to invoke web-hook", zap.Error(err))
	} else if resp.StatusCode != http.StatusOK {
		h.logger.Warn("Webhook endpoint responded with non-200 code",
			zap.String("status", resp.Status),
			zap.Int("status_code", resp.StatusCode))
	}

	if err := resp.Body.Close(); err != nil {
		h.logger.Error("Failed to close response body", zap.Error(err))
	}
}

func (h *WebHook) OnInterfaceAdded(i *daemon.Interface) {
	go h.run(&hooksproto.WebHookBody{
		Type:      rpcproto.EventType_INTERFACE_ADDED,
		Interface: marshalRedactedInterface(i),
	})
}

func (h *WebHook) OnInterfaceRemoved(i *daemon.Interface) {
	go h.run(&hooksproto.WebHookBody{
		Type:      rpcproto.EventType_INTERFACE_REMOVED,
		Interface: marshalRedactedInterface(i),
	})
}

func (h *WebHook) OnInterfaceModified(i *daemon.Interface, _ *wg.Interface, m daemon.InterfaceModifier) {
	go h.run(&hooksproto.WebHookBody{
		Type:      rpcproto.EventType_INTERFACE_MODIFIED,
		Interface: marshalRedactedInterface(i),
		Modified:  m.Strings(),
	})
}

func (h *WebHook) OnPeerAdded(p *daemon.Peer) {
	go h.run(&hooksproto.WebHookBody{
		Type: rpcproto.EventType_PEER_ADDED,
		Peer: p.Marshal().Redact(),
	})
}

func (h *WebHook) OnPeerRemoved(p *daemon.Peer) {
	go h.run(&hooksproto.WebHookBody{
		Type: rpcproto.EventType_PEER_REMOVED,
		Peer: p.Marshal().Redact(),
	})
}

func (h *WebHook) OnPeerModified(p *daemon.Peer, _ *wgtypes.Peer, m daemon.PeerModifier, _, _ []net.IPNet) {
	go h.run(&hooksproto.WebHookBody{
		Type:     rpcproto.EventType_PEER_MODIFIED,
		Peer:     p.Marshal().Redact(),
		Modified: m.Strings(),
	})
}

func (h *WebHook) OnPeerStateChanged(p *daemon.Peer, _, _ daemon.PeerState) {
	pm := p.Marshal().Redact()

	if epi := epdisc.Get(p.Interface); epi != nil {
		epp := epi.Peers[p]
		pm.Ice = epp.Marshal()
	}

	go h.run(&hooksproto.WebHookBody{
		Type: rpcproto.EventType_PEER_STATE_CHANGED,
		Peer: pm,
	})
}
