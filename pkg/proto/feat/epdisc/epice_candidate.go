package epdisc

import (
	"fmt"
	"io"

	"github.com/pion/ice/v2"

	"riasc.eu/wice/pkg/proto"

	t "riasc.eu/wice/pkg/util/terminal"
)

func NewCandidate(ic ice.Candidate) *Candidate {
	c := &Candidate{
		Type:        CandidateType(ic.Type()),
		Foundation:  ic.Foundation(),
		Component:   int32(ic.Component()),
		NetworkType: NetworkType(ic.NetworkType()),
		Priority:    int32(ic.Priority()),
		Address:     ic.Address(),
		Port:        int32(ic.Port()),
		TcpType:     TCPType(ic.TCPType()),
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
	case CandidateType_CANDIDATE_TYPE_HOST:
		ic, err = ice.NewCandidateHost(&ice.CandidateHostConfig{
			Network:    nw.String(),
			Address:    c.Address,
			Port:       int(c.Port),
			Component:  uint16(c.Component),
			Priority:   uint32(c.Priority),
			Foundation: c.Foundation,
			TCPType:    ice.TCPType(c.TcpType),
		})
	case CandidateType_CANDIDATE_TYPE_SERVER_REFLEXIVE:
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
	case CandidateType_CANDIDATE_TYPE_PEER_REFLEXIVE:
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

	case CandidateType_CANDIDATE_TYPE_RELAY:
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

func NewCandidatePairStats(cps *ice.CandidatePairStats) *CandidatePairStats {
	p := &CandidatePairStats{
		LocalCandidateId:           cps.LocalCandidateID,
		RemoteCandidateId:          cps.RemoteCandidateID,
		State:                      CandidatePairState(cps.State),
		Nominated:                  cps.Nominated,
		PacketsSent:                cps.PacketsSent,
		PacketsReceived:            cps.PacketsReceived,
		BytesSent:                  cps.BytesSent,
		BytesReceived:              cps.BytesReceived,
		TotalRoundtripTime:         cps.TotalRoundTripTime,
		CurrenTroundtripTime:       cps.CurrentRoundTripTime,
		AvailableOutgoingBitrate:   cps.AvailableOutgoingBitrate,
		AvailableIncomingBitrate:   cps.AvailableIncomingBitrate,
		CircuitBreakerTriggerCount: cps.CircuitBreakerTriggerCount,
		RequestsReceived:           cps.RequestsReceived,
		RequestsSent:               cps.RequestsReceived,
		ResponsesReceived:          cps.ResponsesReceived,
		RetransmissionsReceived:    cps.RetransmissionsReceived,
		RetransmissionsSent:        cps.RetransmissionsSent,
		ConsentRequestsSent:        cps.ConsentRequestsSent,
	}

	if !cps.Timestamp.IsZero() {
		p.Timestamp = proto.Time(cps.Timestamp)
	}

	if !cps.LastPacketSentTimestamp.IsZero() {
		p.LastPacketSentTimestamp = proto.Time(cps.LastPacketSentTimestamp)
	}

	if !cps.LastPacketReceivedTimestamp.IsZero() {
		p.LastPacketReceivedTimestamp = proto.Time(cps.LastPacketReceivedTimestamp)
	}

	if !cps.FirstRequestTimestamp.IsZero() {
		p.FirstRequestTimestamp = proto.Time(cps.FirstRequestTimestamp)
	}

	if !cps.LastRequestTimestamp.IsZero() {
		p.LastRequestTimestamp = proto.Time(cps.LastRequestTimestamp)
	}

	if !cps.LastResponseTimestamp.IsZero() {
		p.LastResponseTimestamp = proto.Time(cps.LastResponseTimestamp)
	}

	if !cps.ConsentExpiredTimestamp.IsZero() {
		p.ConsentExpiredTimestamp = proto.Time(cps.ConsentExpiredTimestamp)
	}

	return p
}

func NewCandidateStats(cs *ice.CandidateStats) *CandidateStats {
	return &CandidateStats{
		Timestamp:     proto.Time(cs.Timestamp),
		Id:            cs.ID,
		NetworkType:   NetworkType(cs.NetworkType),
		Ip:            cs.IP,
		Port:          int32(cs.Port),
		CandidateType: CandidateType(cs.CandidateType),
		Priority:      cs.Priority,
		Url:           cs.URL,
		RelayProtocol: cs.RelayProtocol,
		Deleted:       cs.Deleted,
	}
}

func (cp *CandidatePair) ToString() string {
	return fmt.Sprintf("%s <-> %s", cp.Local.ToString(), cp.Remote.ToString())
}

func (c *Candidate) ToString() string {
	return fmt.Sprintf("%s[%s, %s:%d]", ice.CandidateType(c.Type), ice.NetworkType(c.NetworkType), c.Address, c.Port)
}

func (cs *CandidateStats) ToString() string {
	return fmt.Sprintf("%s[%s, %s:%d]", ice.CandidateType(cs.CandidateType), ice.NetworkType(cs.NetworkType), cs.Ip, cs.Port)
}

func (cs *CandidateStats) Dump(wr io.Writer) error {
	// wri := util.NewIndenter(wr, "  ")

	if _, err := fmt.Fprintf(wr, t.Color("candidate", t.Bold, t.FgMagenta)+": "+t.Color("%s", t.FgMagenta)+"\n", cs.ToString()); err != nil {
		return err
	}

	return nil
}
