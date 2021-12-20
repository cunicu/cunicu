package signaling

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"

	"riasc.eu/wice/pkg/crypto"

	"github.com/pion/ice/v2"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Backend

type BackendType string // URL schemes
type BackendFactory func(*url.URL, map[string]string) (Backend, error)
type BackendPlugin struct {
	New         BackendFactory
	Description string
}

// Candidates

type jsonCandidate struct {
	Type        string  `json:"type"`
	Foundation  string  `json:"foundation"`
	Component   int     `json:"component"`
	NetworkType string  `json:"network"`
	Priority    int     `json:"priority"`
	Address     string  `json:"address"`
	Port        int     `json:"port"`
	TCPType     *string `json:"tcp_type,omitempty"`
	RelAddr     *string `json:"related_address,omitempty"`
	RelPort     *int    `json:"related_port,omitempty"`
}

type Candidate struct {
	ice.Candidate
}

func (c *Candidate) MarshalJSON() ([]byte, error) {
	jc := &jsonCandidate{
		Type:        c.Type().String(),
		Foundation:  c.Foundation(),
		Component:   int(c.Component()),
		NetworkType: c.NetworkType().String(),
		Priority:    int(c.Priority()),
		Address:     c.Address(),
		Port:        c.Port(),
	}

	if c.TCPType() != ice.TCPTypeUnspecified {
		t := c.TCPType().String()
		jc.TCPType = &t
	}

	if r := c.RelatedAddress(); r != nil && r.Address != "" && r.Port != 0 {
		jc.RelAddr = &r.Address
		jc.RelPort = &r.Port
	}

	return json.Marshal(jc)
}

func (c *Candidate) UnmarshalJSON(data []byte) error {
	var jc jsonCandidate

	if err := json.Unmarshal(data, &jc); err != nil {
		return err
	}

	relAddr := ""
	relPort := 0
	if jc.RelAddr != nil && jc.RelPort != nil {
		relAddr = *jc.RelAddr
		relPort = int(*jc.RelPort)
	}

	tcpType := ice.TCPTypeUnspecified
	if jc.TCPType != nil {
		tcpType = ice.NewTCPType(*jc.TCPType)
	}

	var ic ice.Candidate
	switch jc.Type {
	case "host":
		ic, err = ice.NewCandidateHost(&ice.CandidateHostConfig{
			CandidateID: "",
			Network:     jc.NetworkType,
			Address:     jc.Address,
			Port:        int(jc.Port),
			Component:   uint16(jc.Component),
			Priority:    uint32(jc.Priority),
			Foundation:  jc.Foundation,
			TCPType:     tcpType})
	case "srflx":
		ic, err = ice.NewCandidateServerReflexive(&ice.CandidateServerReflexiveConfig{
			CandidateID: "",
			Network:     jc.NetworkType,
			Address:     jc.Address,
			Port:        int(jc.Port),
			Component:   uint16(jc.Component),
			Priority:    uint32(jc.Priority),
			Foundation:  jc.Foundation,
			RelAddr:     relAddr,
			RelPort:     relPort,
		})
	case "prflx":
		ic, err = ice.NewCandidatePeerReflexive(&ice.CandidatePeerReflexiveConfig{
			CandidateID: "",
			Network:     jc.NetworkType,
			Address:     jc.Address,
			Port:        int(jc.Port),
			Component:   uint16(jc.Component),
			Priority:    uint32(jc.Priority),
			Foundation:  jc.Foundation,
			RelAddr:     relAddr,
			RelPort:     relPort,
		})

	case "relay":
		ic, err = ice.NewCandidateRelay(&ice.CandidateRelayConfig{
			CandidateID: "",
			Network:     jc.NetworkType,
			Address:     jc.Address,
			Port:        int(jc.Port),
			Component:   uint16(jc.Component),
			Priority:    uint32(jc.Priority),
			Foundation:  jc.Foundation,
			RelAddr:     relAddr,
			RelPort:     relPort,
			OnClose:     nil,
		})

	default:
		err = fmt.Errorf("unknown candidate type: %s", jc.Type)
	}
	if err != nil {
		return nil
	}

	c.Candidate = ic

	return nil
}

// Peers

type jsonPeer struct {
	PublicKey  crypto.Key  `json:"public_key"`
	AllowedIPs []net.IPNet `json:"allowed_ips,omitempty"`
}

type Peer struct {
	wgtypes.Peer
}

func (p *Peer) PublicKey() crypto.Key {
	return crypto.Key(p.Peer.PublicKey)
}

func (p *Peer) MarshalJSON() ([]byte, error) {
	jp := jsonPeer{
		PublicKey:  p.PublicKey(),
		AllowedIPs: p.AllowedIPs,
	}

	return json.Marshal(jp)
}

func (p *Peer) UnmarshalJSON(data []byte) error {
	var jp jsonPeer

	return json.Unmarshal(data, &jp)
}

// Offers

type ImplementationType string

const (
	ImplementationTypeFull ImplementationType = "full"
	ImplementationTypeLite ImplementationType = "lite"
)

type Role string

const (
	RoleControlled  Role = "controlled"
	RoleControlling Role = "controlling"
)

type OfferMap map[crypto.Key]Offer

// SDP-like session description
// See: https://www.rfc-editor.org/rfc/rfc8866.html
type Offer struct {
	Version            int64              `json:"version"`
	Role               Role               `json:"role"`
	Implementation     ImplementationType `json:"implementation"`
	Candidates         []Candidate        `json:"candidates"`
	Ufrag              string             `json:"ufrag"`
	Pwd                string             `json:"pwd"`
	Epoch              int64              `json:"epoch"`
	CleartextSignature string             `json:"signature"`
}

func NewOffer() Offer {
	return Offer{
		Epoch:          0,
		Candidates:     []Candidate{},
		Implementation: ImplementationTypeFull,
	}
}
