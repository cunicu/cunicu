package proxy

import (
	"net"

	"github.com/pion/ice/v2"
	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/log"
)

func CreateUDPMux() (ice.UDPMux, int, error) {
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return nil, 0, err
	}

	lAddr := conn.LocalAddr().(*net.UDPAddr)

	mux := ice.NewUDPMuxDefault(ice.UDPMuxParams{
		UDPConn: conn,
		Logger:  log.NewPionLoggerFactory(zap.L()).NewLogger("udpmux"),
	})

	return mux, lAddr.Port, nil
}

func CreateUniversalUDPMux() (ice.UniversalUDPMux, int, error) {
	// We do not need a filtered connection here as we anyway need to redirect
	// the non-STUN traffic via nftables

	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return nil, 0, err
	}

	lAddr := conn.LocalAddr().(*net.UDPAddr)

	mux := ice.NewUniversalUDPMuxDefault(ice.UniversalUDPMuxParams{
		UDPConn: conn,
		Logger:  log.NewPionLoggerFactory(zap.L()).NewLogger("udpmux-universal"),
	})

	return mux, lAddr.Port, nil
}
