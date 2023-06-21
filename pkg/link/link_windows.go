// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package link

import (
	"net"
	"strconv"

	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/log"
)

type WindowsLink struct {
	index int

	logger *log.Logger
}

func (d *WindowsLink) AddAddress(ip net.IPNet) error {
	d.logger.Debug("Add address", zap.String("addr", ip.String()))

	return errNotSupported
}

func (d *WindowsLink) AddRoute(dst net.IPNet, gw net.IP, _ int) error {
	d.logger.Debug("Add route",
		zap.String("dst", dst.String()),
		zap.String("gw", gw.String()))

	return errNotSupported
}

func (d *WindowsLink) DeleteAddress(ip net.IPNet) error {
	d.logger.Debug("Delete address", zap.String("addr", ip.String()))

	return errNotSupported
}

func (d *WindowsLink) DeleteRoute(dst net.IPNet, _ int) error {
	d.logger.Debug("Delete route",
		zap.String("dst", dst.String()))

	return errNotSupported
}

func (d *WindowsLink) Index() int {
	return -1
}

func (d *WindowsLink) Flags() net.Flags {
	i, err := net.InterfaceByIndex(d.index)
	if err != nil {
		panic(err)
	}

	return i.Flags
}

func (d *WindowsLink) Type() string {
	return "" // TODO: Is this supported?
}

func (d *WindowsLink) MTU() int {
	// MTU is a route attribute which we need to adjust for all routes added for the interface
	return -1
}

func (d *WindowsLink) SetMTU(mtu int) error {
	d.logger.Debug("Set link MTU", zap.Int("mtu", mtu))

	// MTU is a route attribute which we need to adjust for all routes added for the interface
	return errNotSupported
}

func (d *WindowsLink) SetUp() error {
	d.logger.Debug("Set link up")

	return errNotSupported
}

func (d *WindowsLink) SetDown() error {
	d.logger.Debug("Set link down")

	return errNotSupported
}

func (d *WindowsLink) Close() error {
	d.logger.Debug("Deleting kernel device")

	return nil
}

func DetectMTU(_ net.IP, _ int) (int, error) {
	// TODO: Thats just a guess
	return EthernetMTU, nil
}

func DetectDefaultMTU(_ int) (int, error) {
	// TODO: Thats just a guess
	return EthernetMTU, nil
}

func Table(str string) (int, error) {
	i, err := strconv.Atoi(str)
	if err != nil {
		return -1, err
	}

	return i, nil
}
