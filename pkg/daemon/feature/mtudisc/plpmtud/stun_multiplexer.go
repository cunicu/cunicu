package plpmtud

import (
	"net"

	"github.com/pion/stun"
)

const (
	StunMethodFallthrough stun.Method = 0xffff
)

type StunHandler func(*stun.Message) error

type StunMultiplexer struct {
	*net.UDPConn

	methodHandlers map[stun.Method]StunHandler
	dataHandler    func([]byte)
}

func NewStunMultiplexer(c *net.UDPConn) (*StunMultiplexer, error) {
	sc := &StunMultiplexer{
		UDPConn: c,

		methodHandlers: map[stun.Method]StunHandler{},
	}

	return sc, nil
}

func (c *StunMultiplexer) RegisterStunHandler(m stun.Method, h StunHandler) {
	c.methodHandlers[m] = h
}

func (c *StunMultiplexer) ReadFromUDP(b []byte) (int, *net.UDPAddr, error) {
	for {
		n, addr, err := c.UDPConn.ReadFromUDP(b)
		if err != nil {
			return -1, nil, err
		}

		if stun.IsMessage(b) {
			msg := &stun.Message{
				Raw: b,
			}

			msg.Decode()

			if h, ok := c.methodHandlers[msg.Type.Method]; ok {
				h(msg)
			} else if h, ok := c.methodHandlers[StunMethodFallthrough]; ok {
				h(msg)
			}
		} else {
			return n, addr, nil
		}
	}
}

func (c *StunMultiplexer) WriteStunMessage(msg *stun.Message) error {
	msg.Encode()

	_, err := c.Write(msg.Raw)

	return err
}
