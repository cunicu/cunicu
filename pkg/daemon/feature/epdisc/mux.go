// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package epdisc

import (
	"fmt"
	"math/rand"
	"net"
	"path/filepath"

	"github.com/pion/ice/v4"
	"go.uber.org/zap"

	"cunicu.li/cunicu/pkg/config"
	icex "cunicu.li/cunicu/pkg/ice"
	"cunicu.li/cunicu/pkg/link"
	"cunicu.li/cunicu/pkg/log"
	netx "cunicu.li/cunicu/pkg/net"
	"cunicu.li/cunicu/pkg/wg"
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

		stunConn := filteredConn.AddPacketReadHandlerConn(&netx.STUNPacketHandler{
			Logger: i.logger.Named("stun_conn"),
		})

		return stunConn, nil
	}

	ifFilter := func(name string) bool {
		if include, err := filepath.Match(i.Settings.ICE.InterfacesInclude, name); err != nil {
			return false
		} else if exclude, err := filepath.Match(i.Settings.ICE.InterfacesExclude, name); err != nil {
			return false
		} else if !include || exclude {
			return false
		}

		// Do not use our own WireGuard interfaces
		if i.Daemon.InterfaceByName(name) != nil {
			return false
		}

		// TODO: Check why we cant use Daemon.InterfaceByName()
		if lnk, err := link.FindLink(name); err != nil {
			return false
		} else if lnk.Type() == link.TypeWireGuard {
			return false
		}

		return true
	}

	i.mux, err = icex.NewMultiUDPMuxWithListen(listen, ifFilter, nil, i.Settings.ICE.NetworkTypes, false, i.logger)
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

	stunConn := filteredConn.AddPacketReadHandlerConn(&netx.STUNPacketHandler{
		Logger: i.logger.Named("stun_conn"),
	})

	i.muxSrflx = ice.NewUniversalUDPMuxDefault(ice.UniversalUDPMuxParams{
		UDPConn: stunConn,
		Logger:  log.NewPionLogger(i.logger, "ice.udpmux"),
	})

	lAddr := udpConn.LocalAddr().(*net.UDPAddr) //nolint:forcetypeassert

	i.muxSrflxPort = lAddr.Port

	i.logger.Debug("Created UDP mux for server reflexive candidates", zap.Int("port", i.muxSrflxPort))

	return nil
}
