// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package epdisc

import (
	"fmt"
	"io"

	"github.com/pion/ice/v2"

	"github.com/stv0g/cunicu/pkg/proto"
	"github.com/stv0g/cunicu/pkg/tty"
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

	if r := ic.RelatedAddress(); r != nil {
		c.RelatedAddress = &RelatedAddress{
			Address: r.Address,
			Port:    int32(r.Port),
		}
	}

	if rc, ok := ic.(*ice.CandidateRelay); ok {
		c.RelayProtocol = NewProtocol(rc.RelayProtocol())
	}

	return c
}

func (c *Candidate) ICECandidate() (ice.Candidate, error) {
	var err error

	relAddr := ""
	relPort := 0
	if c.RelatedAddress != nil {
		relAddr = c.RelatedAddress.Address
		relPort = int(c.RelatedAddress.Port)
	}

	nw := ice.NetworkType(c.NetworkType)

	var ic ice.Candidate
	switch c.Type {
	case CandidateType_HOST:
		ic, err = ice.NewCandidateHost(&ice.CandidateHostConfig{
			Network:    nw.String(),
			Address:    c.Address,
			Port:       int(c.Port),
			Component:  uint16(c.Component),
			Priority:   uint32(c.Priority),
			Foundation: c.Foundation,
			TCPType:    ice.TCPType(c.TcpType),
		})
	case CandidateType_SERVER_REFLEXIVE:
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
	case CandidateType_PEER_REFLEXIVE:
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

	case CandidateType_RELAY:
		ic, err = ice.NewCandidateRelay(&ice.CandidateRelayConfig{
			Network:       nw.String(),
			Address:       c.Address,
			Port:          int(c.Port),
			Component:     uint16(c.Component),
			Priority:      uint32(c.Priority),
			Foundation:    c.Foundation,
			RelAddr:       relAddr,
			RelPort:       relPort,
			RelayProtocol: c.RelayProtocol.ToString(),
		})

	default:
		err = fmt.Errorf("%w: %s", ice.ErrUnknownCandidateTyp, c.Type)
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
		CurrentRoundtripTime:       cps.CurrentRoundTripTime,
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
		RelayProtocol: NewProtocol(cs.RelayProtocol),
		Deleted:       cs.Deleted,
	}
}

func (cp *CandidatePair) ToString() string {
	return fmt.Sprintf("%s <-> %s", cp.Local.ToString(), cp.Remote.ToString())
}

func (c *Candidate) ToString() string {
	var addr string
	switch c.NetworkType {
	case NetworkType_UDP6, NetworkType_TCP6:
		addr = fmt.Sprintf("[%s]", c.Address)
	case NetworkType_UDP4, NetworkType_TCP4:
		addr = c.Address
	case NetworkType_UNSPECIFIED_NETWORK_TYPE:
	}

	var nt string
	if c.Type == CandidateType_RELAY && c.RelayProtocol != RelayProtocol_UNSPECIFIED_RELAY_PROTOCOL {
		nt = fmt.Sprintf("%s->%s",
			c.RelayProtocol.ToString(),
			ice.NetworkType(c.NetworkType),
		)
	} else {
		nt = ice.NetworkType(c.NetworkType).String()
	}

	return fmt.Sprintf("%s[%s, %s:%d]", ice.CandidateType(c.Type), nt, addr, c.Port)
}

func (cs *CandidateStats) ToString() string {
	var addr string
	switch cs.NetworkType {
	case NetworkType_UDP6, NetworkType_TCP6:
		addr = fmt.Sprintf("[%s]", cs.Ip)
	case NetworkType_UDP4, NetworkType_TCP4:
		addr = cs.Ip
	case NetworkType_UNSPECIFIED_NETWORK_TYPE:
	}

	var nt string
	if cs.CandidateType == CandidateType_RELAY && cs.RelayProtocol != RelayProtocol_UNSPECIFIED_RELAY_PROTOCOL {
		nt = fmt.Sprintf("%s->%s",
			cs.RelayProtocol.ToString(),
			ice.NetworkType(cs.NetworkType),
		)
	} else {
		nt = ice.NetworkType(cs.NetworkType).String()
	}

	return fmt.Sprintf("%s[%s, %s:%d]", ice.CandidateType(cs.CandidateType), nt, addr, cs.Port)
}

func (cs *CandidateStats) Dump(wr io.Writer) error {
	if _, err := fmt.Fprintf(wr, tty.Mods("candidate", tty.Bold, tty.FgMagenta)+": "+tty.Mods("%s", tty.FgMagenta)+"\n", cs.ToString()); err != nil {
		return err
	}

	return nil
}
