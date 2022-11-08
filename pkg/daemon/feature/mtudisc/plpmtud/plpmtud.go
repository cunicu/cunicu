// Package plpmtud implements Packetization Layer Path MTU Discovery for Datagram Transports
// according to RFC8899 for detecting the maximum packet size (MPS).
package plpmtud

import (
	"time"

	"github.com/pion/stun"
)

const (
	DefaultMaxProbes = 3
	DefaultMinPLPMTU = 1280

	ProbeTimerTimeout = 15 * time.Second
	RaiseTimerTimeout = 600 * time.Second

	StunMethodProbe  stun.Method = 0x00D
	StunMethodReport stun.Method = 0x00E

	StunAttrPadding             stun.AttrType = 0x0026
	StunAttrIdentifier          stun.AttrType = 0x0031 // TODO: use final IANA allocation
	StunAttrProbePMTUDSupported stun.AttrType = 0x8031 // TODO: use final IANA allocation
)

type state byte

const (
	StateUnspecified    state = iota
	StateDisabled             = iota
	StateBase                 = iota
	StateError                = iota
	StateSearching            = iota
	StateSearchComplete       = iota
)

type eventType byte

const (
	EventAcknowledgementReceived eventType = iota
	EventConnected                         = iota
	EventDisconnected                      = iota
	EventNewLinkMTU                        = iota
	EventProbeTimerExpired                 = iota
	EventRaiseTimerExpired                 = iota
	EventPacketTooBigReceived              = iota
	EventBlackHoleDetected                 = iota
)

type event struct {
	Type eventType
	MTU  uint
}

// Discoverer implements the core state-machine of the
// Datagram Packet Layer Path MTU Discovery (DPLPMTUD) protocol
//
// DPLPMTUD aims at detected the largest supported MTU on a give path.
//
// See: https://datatracker.ietf.org/doc/html/rfc8899
type Discoverer struct {
	State state

	probedSize uint
	probeCount uint

	MaxProbes  uint
	MinPLPMTU  uint
	MaxPLPMTU  uint
	BasePLPMTU uint

	prober           Prober
	acknowledgedSize uint

	events chan event

	probeTimer *time.Timer
	raiseTimer *time.Timer
}

func NewDiscoverer(maxMTU uint, prober Prober) *Discoverer {
	d := &Discoverer{
		State: StateDisabled,

		prober: prober,

		probedSize: 0,
		probeCount: 0,

		MaxProbes: DefaultMaxProbes,
		MinPLPMTU: DefaultMinPLPMTU,
		MaxPLPMTU: maxMTU,

		events: make(chan event),
	}

	prober.RegisterDiscoverer(d)

	go d.run()

	return d
}

func (d *Discoverer) run() {
	for evt := range d.events {
		d.State = d.nextState(evt)
	}
}

func (d *Discoverer) nextState(e event) state {
	evt := e.Type
	mtu := e.MTU

	if evt == EventDisconnected {
		return StateDisabled
	}

	next := StateUnspecified

	switch d.State {
	case StateDisabled:
		if evt == EventConnected {
			next = StateBase
		}

	case StateBase:
		switch evt {
		case EventAcknowledgementReceived:
			d.ackedProbe()
			next = StateSearching

		case EventProbeTimerExpired:
			if d.probeCount == d.MaxProbes {
				next = StateError
			} else if d.probeCount < d.MaxProbes {
				next = StateBase
			}

		case EventPacketTooBigReceived:
			if mtu < d.BasePLPMTU {
				next = StateError
			}
		}

	case StateSearching:
		switch evt {
		case EventAcknowledgementReceived:
			d.ackedProbe()

			if d.acknowledgedSize == d.MaxPLPMTU {
				next = StateSearchComplete
			} else {
				next = StateSearching
			}

		case EventProbeTimerExpired:
			d.probeCount++

			if err := d.sendProbeRequest(d.BasePLPMTU); err != nil {
				next = StateDisabled
			} else {
				next = StateSearching
			}

		case EventPacketTooBigReceived:
			if mtu == d.acknowledgedSize {
				next = StateSearchComplete
			}

		case EventBlackHoleDetected:
			next = StateBase
		}

	case StateSearchComplete:
		switch evt {
		case EventRaiseTimerExpired:
			next = StateSearching

		case EventBlackHoleDetected:
			next = StateBase
		}

	case StateError:
	}

	switch next {
	case StateUnspecified:

	case StateSearching:
		if err := d.sendProbeRequest(d.nextProbeSize()); err != nil {
			next = StateDisabled
		}

	case StateBase:
		d.probeCount = 0
		if err := d.sendProbeRequest(d.BasePLPMTU); err != nil {
			next = StateDisabled
		}

	case StateSearchComplete:
		d.raiseTimer = time.AfterFunc(RaiseTimerTimeout, func() {
			d.events <- event{EventRaiseTimerExpired, 0}
		})
	}

	return next
}

func (d *Discoverer) ackedProbe() {
	d.probeTimer.Stop()
	d.probeCount = 0
	d.acknowledgedSize = d.probedSize
}

func (d *Discoverer) nextProbeSize() uint {
	if incr := (d.MaxPLPMTU - d.probedSize) / 2; incr > 0 {
		return d.probedSize + incr
	} else {
		return d.MaxPLPMTU
	}
}

func (d *Discoverer) sendProbeRequest(mtu uint) error {
	d.probedSize = mtu

	if err := d.prober.SendProbeRequest(mtu); err != nil {
		return err
	}

	d.probeTimer = time.AfterFunc(ProbeTimerTimeout, func() {
		d.events <- event{EventProbeTimerExpired, 0}
	})

	d.probeCount++

	return nil
}

func (d *Discoverer) sendProbeResponse(mtu uint) error {
	if err := d.prober.SendProbeResponse(mtu); err != nil {
		return err
	}

	return nil
}

func (d *Discoverer) OnNewLinkMTU(mtu uint) {
	d.events <- event{EventNewLinkMTU, mtu}
}

func (d *Discoverer) OnProbeRequest(mtu uint) {
	d.sendProbeResponse(mtu)
}

func (d *Discoverer) OnProbeResponse(mtu uint) {
	d.events <- event{EventAcknowledgementReceived, mtu}
}

func (d *Discoverer) OnConnectionLost() {
	d.events <- event{EventDisconnected, 0}
}

func (d *Discoverer) OnConnectionEstablished() {
	d.events <- event{EventConnected, 0}
}

func (d *Discoverer) OnPacketTooBig(mtu uint) {
	d.events <- event{EventPacketTooBigReceived, mtu}
}
