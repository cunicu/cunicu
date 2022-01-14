package pb

import (
	"fmt"
	"strings"

	"github.com/pion/ice/v2"
)

func NewCandidate(ic ice.Candidate) *Candidate {
	c := &Candidate{
		Type:        Candidate_Type(ic.Type()),
		Foundation:  ic.Foundation(),
		Component:   int32(ic.Component()),
		NetworkType: NewNetworkType(ic.NetworkType()),
		Priority:    int32(ic.Priority()),
		Address:     ic.Address(),
		Port:        int32(ic.Port()),
		TcpType:     Candidate_TCPType(ic.TCPType()),
	}

	if r := c.RelatedAddress; r != nil {
		c.RelatedAddress = &RelatedAddress{
			Address: r.Address,
			Port:    r.Port,
		}
	}

	return c
}

func (c *Candidate) ICECandidate() (ice.Candidate, error) {
	var err error

	var relAddr = ""
	var relPort = 0
	if c.RelatedAddress != nil {
		relAddr = c.RelatedAddress.Address
		relPort = int(c.RelatedAddress.Port)
	}

	var ic ice.Candidate
	switch c.Type {
	case Candidate_TYPE_HOST:
		ic, err = ice.NewCandidateHost(&ice.CandidateHostConfig{
			CandidateID: "",
			Network:     strings.ToLower(c.NetworkType.String()),
			Address:     c.Address,
			Port:        int(c.Port),
			Component:   uint16(c.Component),
			Priority:    uint32(c.Priority),
			Foundation:  c.Foundation,
			TCPType:     ice.TCPType(c.TcpType),
		})
	case Candidate_TYPE_SERVER_REFLEXIVE:
		ic, err = ice.NewCandidateServerReflexive(&ice.CandidateServerReflexiveConfig{
			CandidateID: "",
			Network:     strings.ToLower(c.NetworkType.String()),
			Address:     c.Address,
			Port:        int(c.Port),
			Component:   uint16(c.Component),
			Priority:    uint32(c.Priority),
			Foundation:  c.Foundation,
			RelAddr:     relAddr,
			RelPort:     relPort,
		})
	case Candidate_TYPE_PEER_REFLEXIVE:
		ic, err = ice.NewCandidatePeerReflexive(&ice.CandidatePeerReflexiveConfig{
			CandidateID: "",
			Network:     strings.ToLower(c.NetworkType.String()),
			Address:     c.Address,
			Port:        int(c.Port),
			Component:   uint16(c.Component),
			Priority:    uint32(c.Priority),
			Foundation:  c.Foundation,
			RelAddr:     relAddr,
			RelPort:     relPort,
		})

	case Candidate_TYPE_RELAY:
		ic, err = ice.NewCandidateRelay(&ice.CandidateRelayConfig{
			CandidateID: "",
			Network:     strings.ToLower(c.NetworkType.String()),
			Address:     c.Address,
			Port:        int(c.Port),
			Component:   uint16(c.Component),
			Priority:    uint32(c.Priority),
			Foundation:  c.Foundation,
			RelAddr:     relAddr,
			RelPort:     relPort,
		})

	default:
		err = fmt.Errorf("unknown candidate type: %s", c.Type)
	}

	return ic, err
}

func NewNetworkType(nt ice.NetworkType) Candidate_NetworkType {
	switch nt {
	case ice.NetworkTypeUDP4:
		return Candidate_UDP4
	case ice.NetworkTypeUDP6:
		return Candidate_UDP6
	case ice.NetworkTypeTCP4:
		return Candidate_TCP4
	case ice.NetworkTypeTCP6:
		return Candidate_TCP6
	}

	return -1
}

func (nt *Candidate_NetworkType) NetworkType() ice.NetworkType {
	switch *nt {
	case Candidate_UDP4:
		return ice.NetworkTypeUDP4
	case Candidate_UDP6:
		return ice.NetworkTypeUDP6
	case Candidate_TCP4:
		return ice.NetworkTypeTCP4
	case Candidate_TCP6:
		return ice.NetworkTypeTCP6
	}

	return -1
}
