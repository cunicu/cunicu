// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package link

import (
	"net"

	"go.uber.org/zap"
)

func (d *BSDLink) AddRoute(dst net.IPNet, gw net.IP, table int) error {
	d.logger.Debug("Add route",
		zap.String("dst", dst.String()),
		zap.String("gw", gw.String()))

	if table != 0 {
		return errNotSupported
	}

	args := []string{"route", "add", "-" + addressFamily(dst), "-net", dst.String()}
	if gw == nil {
		args = append(args, "-interface", d.Name())
	} else {
		args = append(args, gw.String())
	}

	if _, err := run(args...); err != nil {
		return err
	}

	return nil
}

func (d *BSDLink) DeleteRoute(dst net.IPNet, table int) error {
	d.logger.Debug("Delete route",
		zap.String("dst", dst.String()))

	if table != 0 {
		return errNotSupported
	}

	if _, err := run("route", "delete", "-net", dst.String(), "-interface", d.Name()); err != nil {
		return err
	}

	return nil
}
