package socket

import (
	"encoding/json"
	"io"
	"time"

	log "github.com/sirupsen/logrus"

	"net"
	"os"
)

type Server struct {
	listener net.Listener

	logger *log.Entry

	connections map[*Connection]*struct{}
}

type Connection struct {
	server *Server
	logger *log.Entry

	decoder *json.Decoder
	encoder *json.Encoder
}

func Listen(path string) (*Server, error) {
	// Remove old sockets
	if err := os.RemoveAll(path); err != nil {
		log.Fatal(err)
	}

	l, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}

	logger := log.WithField("logger", "socket")

	s := &Server{
		listener:    l,
		logger:      logger,
		connections: map[*Connection]*struct{}{},
	}

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.WithError(err).Error("Failed to accept client connection")
			}

			go s.HandleConn(conn)
		}
	}()

	return s, nil
}

func (s *Server) HandleConn(conn net.Conn) {
	c := &Connection{
		server:  s,
		decoder: json.NewDecoder(conn),
		encoder: json.NewEncoder(conn),
	}

	s.connections[c] = nil

	logger := s.logger.WithField("conn", conn.RemoteAddr().String())

	for {
		var req Request

		if err := c.decoder.Decode(&req); err == io.EOF {
			logger.Info("Connection closed")
			s.connections[c] = nil
			break
		} else if err != nil {
			log.WithError(err).Error("Failed to decode client request")
		} else {
			if err := c.HandleReq(&req); err != nil {
				log.WithError(err).Error("Failed to handle client request")
			}
		}
	}
}

func (c *Connection) HandleReq(req *Request) error {
	c.logger.Info("Handling request: %s", req)

	return nil
}

func (c *Connection) SendEvent(e *Event) error {
	return c.encoder.Encode(e)
}

func (s *Server) BroadcastEvent(e *Event) error {
	if e.Time.IsZero() {
		e.Time = time.Now()
	}

	for conn := range s.connections {
		conn.SendEvent(e)
	}

	return nil
}
