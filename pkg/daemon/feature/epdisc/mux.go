package epdisc

import (
	"fmt"
	"math/rand"
	"net"
	"path/filepath"

	"github.com/pion/ice/v2"
	"github.com/pion/zapion"
	"github.com/stv0g/cunicu/pkg/config"
	icex "github.com/stv0g/cunicu/pkg/ice"
	"github.com/stv0g/cunicu/pkg/link"
	"github.com/stv0g/cunicu/pkg/log"
	netx "github.com/stv0g/cunicu/pkg/net"
	"github.com/stv0g/cunicu/pkg/wg"
	"go.uber.org/zap"
)

func (i *Interface) setupUDPMux() error {
	var err error

	i.muxPort = wg.DefaultPort + rand.Intn(config.EphemeralPortMax-wg.DefaultPort+1) //nolint:gosec

	listen := func(ip net.IP) (net.PacketConn, error) {
		udpConn, err := net.ListenUDP("udp", &net.UDPAddr{
			IP:   ip,
			Port: i.muxPort,
		})
		if err != nil {
			return nil, err
		}

		logger := i.logger.Named("mux_conn")
		filteredConn := netx.NewFilteredConn(udpConn, logger)

		i.muxConns = append(i.muxConns, filteredConn)

		var stunLogger *zap.Logger
		if i.logger.Core().Enabled(zap.DebugLevel) {
			stunLogger = i.logger.Named("stun_conn").WithOptions(log.WithVerbose(5))
		}

		stunConn := filteredConn.AddPacketReadHandlerConn(&netx.STUNPacketHandler{
			Logger: stunLogger,
		})

		return stunConn, nil
	}

	ifFilter := func(name string) bool {
		if match, err := filepath.Match(i.Settings.ICE.InterfaceFilter, name); err != nil {
			return false
		} else if !match {
			return false
		}

		// Do not use our own WireGuard interfaces
		if i.Daemon.InterfaceByName(name) != nil {
			return false
		}

		// TODO: Check why we cant use Daemon.InterfaceByName()
		if lnk, err := link.FindLink(name); err != nil {
			return false
		} else if lnk.Type() == "wireguard" {
			return false
		}

		return true
	}

	i.mux, err = icex.NewMultiUDPMuxWithListen(listen, ifFilter, nil, i.Settings.ICE.NetworkTypes, false)
	if err != nil {
		return fmt.Errorf("failed to create multi UDP mux: %w", err)
	}

	i.logger.Debug("Created UDP mux for host candidates", zap.Int("port", i.muxPort))

	return nil
}

func (i *Interface) setupUniversalUDPMux() error {
	udpConn, err := net.ListenUDP("udp", nil)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	logger := i.logger.Named("mux_conn")
	filteredConn := netx.NewFilteredConn(udpConn, logger)

	i.muxConns = append(i.muxConns, filteredConn)

	lf := zapion.ZapFactory{
		BaseLogger: zap.L().Named("ice"),
	}

	var stunLogger *zap.Logger
	if i.logger.Core().Enabled(zap.DebugLevel) {
		stunLogger = i.logger.Named("stun_conn")
	}

	stunConn := filteredConn.AddPacketReadHandlerConn(&netx.STUNPacketHandler{
		Logger: stunLogger,
	})

	i.muxSrflx = ice.NewUniversalUDPMuxDefault(ice.UniversalUDPMuxParams{
		UDPConn: stunConn,
		Logger:  lf.NewLogger("udpmux-universal"),
	})

	lAddr := udpConn.LocalAddr().(*net.UDPAddr) //nolint:forcetypeassert

	i.muxSrflxPort = lAddr.Port

	i.logger.Debug("Created UDP mux for server reflexive candidates", zap.Int("port", i.muxSrflxPort))

	return nil
}
