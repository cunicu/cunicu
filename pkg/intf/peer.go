package intf

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/pion/ice/v2"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/internal/config"
	pice "riasc.eu/wice/internal/ice"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/proxy"
	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/socket"
)

type Peer struct {
	wgtypes.Peer

	Interface *BaseInterface

	ICEAgent *ice.Agent
	ICEConn  *ice.Conn

	localOffer   *pb.Offer
	remoteOffers chan *pb.Offer

	selectedCandidatePairs chan *ice.CandidatePair

	LastHandshake time.Time

	logger *zap.Logger

	client  *wgctrl.Client
	config  *config.Config
	backend signaling.Backend
	server  *socket.Server
}

func (p *Peer) Close() error {
	if err := p.ICEAgent.Close(); err != nil {
		return err
	}

	p.logger.Info("Closed peer")

	return nil
}

func (p *Peer) String() string {
	return p.PublicKey().String()
}

func (p *Peer) UpdateEndpoint(addr *net.UDPAddr) error {

	peerCfg := wgtypes.PeerConfig{
		PublicKey:         p.Peer.PublicKey,
		UpdateOnly:        true,
		ReplaceAllowedIPs: false,
		Endpoint:          addr,
	}

	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{peerCfg},
	}

	return p.client.ConfigureDevice(p.Interface.Name(), cfg)
}

func (p *Peer) isControlling() bool {
	var pkOur, pkTheir big.Int
	pkOur.SetBytes(p.Interface.Device.PublicKey[:])
	pkTheir.SetBytes(p.Peer.PublicKey[:])

	return pkOur.Cmp(&pkTheir) == -1 // the smaller PK is controlling
}

func (p *Peer) OnModified(new *wgtypes.Peer, modified PeerModifier) {
	if modified&PeerModifiedHandshakeTime > 0 {
		p.LastHandshake = new.LastHandshakeTime
		p.logger.Debug("New handshake", zap.Time("time", new.LastHandshakeTime))
	}

	p.server.BroadcastEvent(&pb.Event{
		Type:  "peer",
		State: "modified",
		Event: &pb.Event_Peer{
			Peer: &pb.PeerEvent{
				Peer: &pb.Peer{
					PublicKey:     p.PublicKey().Bytes(),
					LastHandshake: pb.Time(new.LastHandshakeTime),
				},
			},
		},
	})
}

func (p *Peer) onCandidate(c ice.Candidate) {
	if c == nil {
		p.logger.Info("Candidate gathering completed")
	} else {
		p.logger.Info("Found new local candidate", zap.Any("candidate", c))

		p.localOffer.Candidates = append(p.localOffer.Candidates, pb.NewCandidate(c))
	}

	p.localOffer.Epoch++

	if err := p.backend.PublishOffer(p.PublicKeyPair(), p.localOffer); err != nil {
		p.logger.Error("Failed to publish offer", zap.Error(err))
	}
}

func (p *Peer) onConnectionStateChange(state ice.ConnectionState) {
	stateLower := strings.ToLower(state.String())

	p.logger.Info("Connection state changed", zap.String("state", stateLower))

	p.server.BroadcastEvent(&pb.Event{
		Type:  "state",
		State: "changed",
		Event: &pb.Event_Intf{
			Intf: &pb.InterfaceEvent{
				Interface: &pb.Interface{
					Name: p.Interface.Name(),
					Peers: []*pb.Peer{
						{
							PublicKey: p.PublicKey().Bytes(),
							State:     pb.ConnectionState(state),
						},
					},
				},
			},
		},
	})

	if state == ice.ConnectionStateFailed {
		go p.restartLocal()
	}
}

func (p *Peer) onSelectedCandidatePairChange(a, b ice.Candidate) {
	cp, err := p.ICEAgent.GetSelectedCandidatePair()
	if err != nil {
		p.logger.Warn("Failed to get selected candidate pair")
	}

	p.logger.Info("Selected new candidate pair", zap.Any("pair", cp))

	p.selectedCandidatePairs <- cp
}

func (p *Peer) isRestartingOffer(o *pb.Offer) bool {
	ufrag, pwd, err := p.ICEAgent.GetRemoteUserCredentials()
	if err != nil {
		p.logger.Error("Failed to get local credentials", zap.Error(err))
	}

	credsChanged := ufrag != o.Ufrag || pwd != o.Pwd

	return credsChanged
}

func (p *Peer) onOffer(o *pb.Offer) {
	// if p.isRestartingOffer(o) {
	// 	p.restartRemote(o)
	// }

	for _, c := range o.Candidates {
		ic, err := c.ICECandidate()
		if err != nil {
			p.logger.Error("Failed to decode offer. Ignoring...", zap.Error(err))
			continue
		}

		if err := p.ICEAgent.AddRemoteCandidate(ic); err != nil {
			p.logger.Fatal("Failed to add remote candidate", zap.Error(err))
		}
		p.logger.Debug("Add remote candidate", zap.Any("candidate", c))
	}
}

func (p *Peer) restartLocal() {
	p.logger.Info("Session restart triggered locally", zap.Duration("delay", p.config.RestartInterval))

	time.Sleep(p.config.RestartInterval)

	p.localOffer = p.newOffer()

	p.backend.PublishOffer(p.PublicKeyPair(), p.localOffer)

	offer := <-p.remoteOffers // wait for remote answer

	p.restart(offer)
}

func (p *Peer) restartRemote(offer *pb.Offer) {
	p.logger.Info("Session restart triggered by remote")

	p.localOffer = p.newOffer()

	p.restart(offer)
}

// Performs an ICE restart
//
// This restart can either be triggered by a failed
// ICE connection state (Peer.onConnectionState())
// or by a remote offer which indicates a restart (Peer.onOffer())
func (p *Peer) restart(offer *pb.Offer) {
	if err := p.ICEAgent.Restart("", ""); err != nil {
		p.logger.Error("Failed to restart ICE session", zap.Error(err))
		return
	}

	if err := p.ICEAgent.SetRemoteCredentials(offer.Ufrag, offer.Pwd); err != nil {
		p.logger.Error("Failed to set remote creds", zap.Error(err))
		return
	}

	for _, c := range offer.Candidates {
		ic, err := c.ICECandidate()
		if err != nil {
			p.logger.Error("Failed to decode offer. Ignoring...", zap.Error(err))
			continue
		}

		if err = p.ICEAgent.AddRemoteCandidate(ic); err != nil {
			p.logger.Error("Failed to add remote candidate", zap.Error(err))
		}
	}

	if err := p.ICEAgent.GatherCandidates(); err != nil {
		p.logger.Error("Failed to gather candidates", zap.Error(err))
		return
	}
}

func (p *Peer) start() {
	var err error

	p.logger.Info("Starting new session")

	p.remoteOffers, err = p.backend.SubscribeOffer(p.PublicKeyPair())
	if err != nil {
		p.logger.Fatal("Failed to subscribe to offers", zap.Error(err))
	}

	// Wait for first offer from remote agent before creating ICE connection
	o := <-p.remoteOffers
	p.onOffer(o)

	// Start the ICE Agent. One side must be controlled, and the other must be controlling
	if p.isControlling() {
		p.ICEConn, err = p.ICEAgent.Dial(context.TODO(), o.Ufrag, o.Pwd)
	} else {
		p.ICEConn, err = p.ICEAgent.Accept(context.TODO(), o.Ufrag, o.Pwd)
	}
	if err != nil {
		p.logger.Fatal("Failed to establish ICE connection", zap.Error(err))
	}

	// Wait until we are ready
	var currentProxy proxy.Proxy = nil
	for {
		select {

		// New remote candidate
		case offer := <-p.remoteOffers:
			p.onOffer(offer)

		// New selected candidate pair
		case cp := <-p.selectedCandidatePairs:
			pt := p.config.ProxyType.ProxyType

			// p.logger.Infof("Conntype: %+v", reflect.ValueOf(cp.Local).Elem().Type())

			isTCPRelayCandidate := cp.Local.Type() == ice.CandidateTypeRelay
			if isTCPRelayCandidate {
				pt = proxy.TypeUser
			}

			if currentProxy != nil && currentProxy.Type() == pt {
				// Update endpoint of existing proxy
				addr := p.ICEConn.RemoteAddr().(*net.UDPAddr)
				currentProxy.UpdateEndpoint(addr)
			} else {
				// Stop old proxy
				if currentProxy != nil {
					currentProxy.Close()
				}

				ident := base64.StdEncoding.EncodeToString(p.Peer.PublicKey[:16])

				// Replace proxy
				if currentProxy, err = proxy.NewProxy(pt, ident, p.Interface.ListenPort, p.UpdateEndpoint, p.ICEConn); err != nil {
					p.logger.Fatal("Failed to setup proxy", zap.Error(err))
				}
			}

		}
	}
}

func (p *Peer) newOffer() *pb.Offer {
	offer := &pb.Offer{
		Version:        pb.OfferVersion,
		Type:           pb.Offer_OFFER,
		Implementation: pb.Offer_FULL,
		Epoch:          0,
		Candidates:     []*pb.Candidate{},
	}

	if p.isControlling() {
		offer.Role = pb.Offer_CONTROLLING
	} else {
		offer.Role = pb.Offer_CONTROLLED
	}

	ufrag, pwd, err := p.ICEAgent.GetLocalUserCredentials()
	if err != nil {
		p.logger.Error("Failed to get local user credentials", zap.Error(err))
	}

	offer.Ufrag = ufrag
	offer.Pwd = pwd

	return offer
}

func (p *Peer) PublicKey() crypto.Key {
	return crypto.Key(p.Peer.PublicKey)
}

func (p *Peer) PublicKeyPair() crypto.PublicKeyPair {
	return crypto.PublicKeyPair{
		Ours:   p.Interface.PublicKey(),
		Theirs: p.PublicKey(),
	}
}

func NewPeer(wgp *wgtypes.Peer, i *BaseInterface) (Peer, error) {
	p := Peer{
		Interface:              i,
		Peer:                   *wgp,
		client:                 i.client,
		backend:                i.backend,
		server:                 i.server,
		config:                 i.config,
		selectedCandidatePairs: make(chan *ice.CandidatePair),
		logger: zap.L().Named("peer").With(
			zap.String("intf", i.Name()),
			zap.Any("peer", wgp.PublicKey),
		),
	}

	agentConfig, err := p.config.AgentConfig()
	if err != nil {
		return Peer{}, fmt.Errorf("failed to generate ICE agent: %w", err)
	}

	agentConfig.LoggerFactory = &pice.LoggerFactory{
		Base: p.logger,
	}

	agentConfig.InterfaceFilter = func(name string) bool {
		_, err := p.client.Device(name)
		return p.config.InterfaceFilterICE.Match([]byte(name)) && err != nil
	}

	if i.config.ProxyType.ProxyType == proxy.TypeEBPF {
		if err := proxy.SetupEBPFProxy(agentConfig, p.Interface.ListenPort); err != nil {
			return Peer{}, fmt.Errorf("failed to setup proxy: %w", err)
		}
	}

	if p.ICEAgent, err = ice.NewAgent(agentConfig); err != nil {
		return Peer{}, fmt.Errorf("failed to create ICE agent: %w", err)
	}

	// create initial offer using freshly generated creds
	p.localOffer = p.newOffer()

	ufrag, pwd, err := p.ICEAgent.GetLocalUserCredentials()
	if err != nil {
		return Peer{}, fmt.Errorf("failed to get local credentials: %w", err)
	}

	p.logger.Debug("Peer credentials",
		zap.String("ufrag", ufrag),
		zap.String("pwd", pwd),
	)

	// When we have gathered a new ICE Candidate send it to the remote peer
	if err := p.ICEAgent.OnCandidate(p.onCandidate); err != nil {
		return Peer{}, fmt.Errorf("failed to setup on candidate handler: %w", err)
	}

	// When selected candidate pair changes
	if err := p.ICEAgent.OnSelectedCandidatePairChange(p.onSelectedCandidatePairChange); err != nil {
		return Peer{}, fmt.Errorf("failed to setup on selected candidate pair handler: %w", err)
	}

	// When ICE Connection state has change print to stdout
	if err := p.ICEAgent.OnConnectionStateChange(p.onConnectionStateChange); err != nil {
		return Peer{}, fmt.Errorf("failed to setup on connection state handler: %w", err)
	}

	p.logger.Info("Gathering local candidates")
	if err := p.ICEAgent.GatherCandidates(); err != nil {
		return Peer{}, fmt.Errorf("failed to gather candidates: %w", err)
	}

	go p.start()

	return p, nil
}
