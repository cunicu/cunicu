package proxy

import (
	"fmt"
	"net"

	"github.com/pion/ice/v2"
	"go.uber.org/zap"
	icex "riasc.eu/wice/internal/ice"
)

func CreateUDPMuxSrflx(listenPort int) (ice.UniversalUDPMux, error) {
	addr := net.UDPAddr{
		IP:   net.IPv4zero,
		Port: listenPort,
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create filtered UDP connection: %w", err)
	}

	lf := &icex.LoggerFactory{
		Base: zap.L(),
	}

	mux := ice.NewUniversalUDPMuxDefault(ice.UniversalUDPMuxParams{
		UDPConn: conn,
		Logger:  lf.NewLogger("udpmux"),
	})

	return mux, nil
}
