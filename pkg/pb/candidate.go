package pb

import (
	"fmt"

	"github.com/pion/ice/v2"
)

func NewCandidate(ic ice.Candidate) *Candidate {
	c := &Candidate{
		Type:        Candidate_Type(ic.Type()),
		Foundation:  ic.Foundation(),
		Component:   int32(ic.Component()),
		NetworkType: Candidate_NetworkType(ic.NetworkType()),
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

	nw := ice.NetworkType(c.NetworkType)

	var ic ice.Candidate
	switch c.Type {
	case Candidate_TYPE_HOST:
		ic, err = ice.NewCandidateHost(&ice.CandidateHostConfig{
			Network:    nw.String(),
			Address:    c.Address,
			Port:       int(c.Port),
			Component:  uint16(c.Component),
			Priority:   uint32(c.Priority),
			Foundation: c.Foundation,
			TCPType:    ice.TCPType(c.TcpType),
		})
	case Candidate_TYPE_SERVER_REFLEXIVE:
		ic, err = ice.NewCandidateServerReflexive(&ice.CandidateServerReflexiveConfig{
			Network:    nw.String(),
			Address:    c.Address,
			Port:       int(c.Port),
			Component:  uint16(c.Component),
			Priority:   uint32(c.Priority),
			Foundation: c.Foundation,
			RelAddr:    relAddr,
			RelPort:    relPort,
		})
	case Candidate_TYPE_PEER_REFLEXIVE:
		ic, err = ice.NewCandidatePeerReflexive(&ice.CandidatePeerReflexiveConfig{
			Network:    nw.String(),
			Address:    c.Address,
			Port:       int(c.Port),
			Component:  uint16(c.Component),
			Priority:   uint32(c.Priority),
			Foundation: c.Foundation,
			RelAddr:    relAddr,
			RelPort:    relPort,
		})

	case Candidate_TYPE_RELAY:
		ic, err = ice.NewCandidateRelay(&ice.CandidateRelayConfig{
			Network:    nw.String(),
			Address:    c.Address,
			Port:       int(c.Port),
			Component:  uint16(c.Component),
			Priority:   uint32(c.Priority),
			Foundation: c.Foundation,
			RelAddr:    relAddr,
			RelPort:    relPort,
		})

	default:
		err = fmt.Errorf("unknown candidate type: %s", c.Type)
	}

	return ic, err
}
