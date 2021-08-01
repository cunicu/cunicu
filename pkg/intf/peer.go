package intf

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/big"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/pion/ice/v2"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/args"
	"riasc.eu/wice/pkg/backend"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/proxy"
)

type Peer struct {
	wgtypes.Peer

	Interface *BaseInterface

	ICEAgent *ice.Agent
	ICEConn  *ice.Conn

	localOffer        backend.Offer
	lastRemoteOfferID int64
	remoteOffers      chan backend.Offer

	selectedCandidatePairs chan *ice.CandidatePair

	LastHandshake time.Time

	logger *log.Entry

	client  *wgctrl.Client
	args    *args.Args
	backend backend.Backend
}

func (p *Peer) Close() error {
	err := p.backend.WithdrawOffer(p.PublicKeyPair())
	if err != nil {
		return err
	}

	p.ICEAgent.Close()

	p.logger.Info("Closing peer")

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

func (p *Peer) colorizeConnectionState(state ice.ConnectionState) string {
	if runtime.GOOS != "windows" {
		var color int
		switch state {
		case ice.ConnectionStateNew:
			fallthrough
		case ice.ConnectionStateCompleted:
			fallthrough
		case ice.ConnectionStateClosed:
			color = 37
		case ice.ConnectionStateDisconnected:
			fallthrough
		case ice.ConnectionStateChecking:
			color = 33
		case ice.ConnectionStateFailed:
			color = 31
		case ice.ConnectionStateConnected:
			color = 32
		}
		return fmt.Sprintf("\x1b[%d;1m%s\x1b[0m", color, state)
	} else {
		return state.String()
	}
}

func (p *Peer) isControlling() bool {
	var pkOur, pkTheir big.Int
	pkOur.SetBytes(p.Interface.Device.PublicKey[:])
	pkTheir.SetBytes(p.Peer.PublicKey[:])

	return pkOur.Cmp(&pkTheir) == -1 // the smaller PK is controlling
}

func (p *Peer) OnModified(new *wgtypes.Peer, modified int) {
	if modified&PeerModifiedHandshakeTime > 0 {
		p.LastHandshake = new.LastHandshakeTime
		p.logger.WithField("time", new.LastHandshakeTime).Debug("New handshake")
	}
}

func (p *Peer) onCandidate(c ice.Candidate) {
	if c == nil {
		p.logger.Info("Candidate gathering completed")

		p.localOffer.EndOfCandidates = true
	} else {
		p.logger.WithField("candidate", c).Info("Found new local candidate")

		p.localOffer.Candidates = append(p.localOffer.Candidates, backend.Candidate{
			Candidate: c,
		})
	}

	p.localOffer.Version++

	if err := p.backend.PublishOffer(p.PublicKeyPair(), p.localOffer); err != nil {
		p.logger.WithError(err).Warn("Failed to publish offer")
		os.Exit(-1)
	}
}

func (p *Peer) onConnectionStateChange(state ice.ConnectionState) {

	p.logger.WithField("state", strings.ToLower(state.String())).Infof("Connection state changed: %s", p.colorizeConnectionState(state))

	if state == ice.ConnectionStateFailed {
		go p.restartLocal()
	}
}

func (p *Peer) onSelectedCandidatePairChange(a, b ice.Candidate) {
	cp, err := p.ICEAgent.GetSelectedCandidatePair()
	if err != nil {
		p.logger.Warn("Failed to get selected candidate pair")
	}

	p.logger.WithField("pair", cp).Info("Selected new candidate pair")

	p.selectedCandidatePairs <- cp
}

func (p *Peer) onOffer(o backend.Offer) {
	if p.lastRemoteOfferID > 0 && p.lastRemoteOfferID != o.ID { // && o.Version == 0 {
		p.restartRemote(o)
	}

	for _, c := range o.Candidates {
		err := p.ICEAgent.AddRemoteCandidate(c)
		if err != nil {
			p.logger.WithError(err).Fatal("Failed to add remote candidate")
		}
		p.logger.WithField("candidate", c).Debug("Add remote candidate")
	}

	p.lastRemoteOfferID = o.ID
}

func (p *Peer) restartLocal() {
	p.logger.WithField("id", p.localOffer.ID).Infof("Restarting session triggered locally in %s", p.args.RestartInterval)

	time.Sleep(p.args.RestartInterval)

	p.localOffer = backend.NewOffer()

	p.backend.PublishOffer(p.PublicKeyPair(), p.localOffer)

	offer := <-p.remoteOffers // wait for remote answer
	p.lastRemoteOfferID = offer.ID

	p.restart(offer)
}

func (p *Peer) restartRemote(offer backend.Offer) {
	p.logger.WithField("id", p.localOffer.ID).Info("Restarting session triggered locally")

	id := p.localOffer.ID
	p.localOffer = backend.NewOffer()
	p.localOffer.ID = id // we keep our offer ID when restart is triggered by remote

	p.restart(offer)
}

// Performs an ICE restart
//
// This restart can either be triggered by a failed
// ICE connection state (Peer.onConnectionState())
// or by a remote offer which indicates a restart (Peer.onOffer())
func (p *Peer) restart(offer backend.Offer) {
	var err error
	var localUfrag, localPwd, remoteUfrag, remotePwd string

	if remoteUfrag, remotePwd, err = p.RemoteCredentials(); err != nil {
		p.logger.WithError(err).Error("Failed to get remote credentials")
		return
	}
	if localUfrag, localPwd, err = p.LocalCreds(); err != nil {
		p.logger.WithError(err).Error("Failed to get remote credentials")
		return
	}

	if err := p.ICEAgent.Restart(localUfrag, localPwd); err != nil {
		p.logger.WithError(err).Error("Failed to restart ICE session")
		return
	}

	for _, cand := range offer.Candidates {
		err = p.ICEAgent.AddRemoteCandidate(cand.Candidate)
		if err != nil {
			p.logger.WithError(err).Error("Failed to add remote candidate")
		}
	}

	if err := p.ICEAgent.GatherCandidates(); err != nil {
		p.logger.WithError(err).Error("Failed to gather candidates")
		return
	}

	if err := p.ICEAgent.SetRemoteCredentials(remoteUfrag, remotePwd); err != nil {
		p.logger.WithError(err).Error("Failed to set remote creds")
		return
	}
}

func (p *Peer) start(remoteUfrag, remotePwd string) {
	var err error

	p.logger.WithField("id", p.localOffer.ID).Info("Starting new session")

	p.remoteOffers, err = p.backend.SubscribeOffer(p.PublicKeyPair())
	if err != nil {
		p.logger.WithError(err).Fatal("Failed to subscribe to offers")
	}

	// Wait for first offer from remote agent before creating ICE connection
	o := <-p.remoteOffers
	p.onOffer(o)

	// Start the ICE Agent. One side must be controlled, and the other must be controlling
	if p.isControlling() {
		p.ICEConn, err = p.ICEAgent.Dial(context.TODO(), remoteUfrag, remotePwd)
	} else {
		p.ICEConn, err = p.ICEAgent.Accept(context.TODO(), remoteUfrag, remotePwd)
	}
	if err != nil {
		p.logger.WithError(err).Fatal("Failed to establish ICE connection")
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
			pt := p.args.ProxyType

			// p.logger.Infof("Conntype: %+v", reflect.ValueOf(cp.Local).Elem().Type())

			isTCPRelayCandidate := cp.Local.Type() == ice.CandidateTypeRelay
			if isTCPRelayCandidate {
				pt = proxy.ProxyTypeUser
			}

			if currentProxy != nil && proxy.Type(currentProxy) == pt {
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
					p.logger.WithError(err).Fatal("Failed to setup proxy")
				}
			}

		}
	}
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
	var err error

	p := Peer{
		Interface:              i,
		Peer:                   *wgp,
		client:                 i.client,
		backend:                i.backend,
		lastRemoteOfferID:      -1,
		localOffer:             backend.NewOffer(),
		args:                   i.args,
		selectedCandidatePairs: make(chan *ice.CandidatePair),
		logger: log.WithFields(log.Fields{
			"intf": i.Name,
			"peer": wgp.PublicKey.String(),
		}),
	}

	agentConfig := p.args.AgentConfig

	agentConfig.InterfaceFilter = func(name string) bool {
		_, err := p.client.Device(name)
		return p.args.IceInterfaceRegex.Match([]byte(name)) && err != nil
	}

	var localUfrag, localPwd, remoteUfrag, remotePwd string
	if localUfrag, localPwd, err = p.LocalCreds(); err != nil {
		return Peer{}, fmt.Errorf("failed to get local credentials: %w", err)
	}
	if remoteUfrag, remotePwd, err = p.RemoteCredentials(); err != nil {
		return Peer{}, fmt.Errorf("failed to get remote credentials: %w", err)
	}

	p.logger.WithFields(log.Fields{
		"ufrag_local":  localUfrag,
		"pwd_local":    localPwd,
		"ufrag_remote": remoteUfrag,
		"pwd_remote":   remotePwd,
	}).Debug("Peer credentials")

	agentConfig.LocalUfrag = localUfrag
	agentConfig.LocalPwd = localPwd

	if p.args.ProxyType == proxy.ProxyTypeEBPF {
		proxy.SetupEBPFMux(&agentConfig, p.Interface.ListenPort)
	}

	if p.ICEAgent, err = ice.NewAgent(&agentConfig); err != nil {
		return Peer{}, fmt.Errorf("failed to create ICE agent: %w", err)
	}

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

	go p.start(remoteUfrag, remotePwd)

	return p, nil
}
