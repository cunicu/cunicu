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

	"riasc.eu/wice/pkg/config"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/feat/epdisc"
	"riasc.eu/wice/pkg/util/buildinfo"
	"riasc.eu/wice/pkg/wg"

	icex "riasc.eu/wice/pkg/feat/epdisc/ice"
	hooksproto "riasc.eu/wice/pkg/proto/feat/hooks"
	rpcproto "riasc.eu/wice/pkg/proto/rpc"
)

type WebHook struct {
	*config.WebHookSetting

	logger *zap.Logger
}

func (h *WebHook) run(msg proto.Message) {
	req := &http.Request{
		Method: h.Method,
		URL:    &h.URL.URL,
		Header: http.Header{},
	}

	req.Header.Add("user-agent", buildinfo.UserAgent())
	req.Header.Add("content-type", "application/json")

	for key, value := range h.Headers {
		req.Header.Add(key, value)
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

	if resp, err := http.DefaultClient.Do(req); err != nil {
		h.logger.Error("Failed to invoke web-hook", zap.Error(err))
	} else if resp.StatusCode != http.StatusOK {
		h.logger.Warn("Webhook endpoint responded with non-200 code",
			zap.String("status", resp.Status),
			zap.Int("status_code", resp.StatusCode))
	}
}

func (h *WebHook) OnInterfaceAdded(i *core.Interface) {
	go h.run(&hooksproto.WebHookBody{
		Type:      rpcproto.EventType_INTERFACE_ADDED,
		Interface: marshalRedactedInterface(i),
	})
}

func (h *WebHook) OnInterfaceRemoved(i *core.Interface) {
	go h.run(&hooksproto.WebHookBody{
		Type:      rpcproto.EventType_INTERFACE_REMOVED,
		Interface: marshalRedactedInterface(i),
	})
}

func (h *WebHook) OnInterfaceModified(i *core.Interface, old *wg.Device, m core.InterfaceModifier) {
	go h.run(&hooksproto.WebHookBody{
		Type:      rpcproto.EventType_INTERFACE_MODIFIED,
		Interface: marshalRedactedInterface(i),
		Modified:  m.Strings(),
	})
}

func (h *WebHook) OnPeerAdded(p *core.Peer) {
	go h.run(&hooksproto.WebHookBody{
		Type: rpcproto.EventType_PEER_ADDED,
		Peer: marshalRedactedPeer(p),
	})
}

func (h *WebHook) OnPeerRemoved(p *core.Peer) {
	go h.run(&hooksproto.WebHookBody{
		Type: rpcproto.EventType_PEER_REMOVED,
		Peer: marshalRedactedPeer(p),
	})
}

func (h *WebHook) OnPeerModified(p *core.Peer, old *wgtypes.Peer, m core.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
	go h.run(&hooksproto.WebHookBody{
		Type:     rpcproto.EventType_PEER_MODIFIED,
		Peer:     marshalRedactedPeer(p),
		Modified: m.Strings(),
	})
}

func (h *WebHook) OnConnectionStateChange(p *epdisc.Peer, new, prev icex.ConnectionState) {
	pm := marshalRedactedPeer(p.Peer)
	pm.Ice = p.Marshal()

	go h.run(&hooksproto.WebHookBody{
		Type: rpcproto.EventType_PEER_CONNECTION_STATE_CHANGED,
		Peer: pm,
	})
}
