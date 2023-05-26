// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package net

import (
	"bytes"
	"encoding/binary"
	"net"

	"golang.org/x/exp/slices"
)

func CmpUDPAddr(a, b *net.UDPAddr) int {
	if a == nil && b == nil {
		return 0
	}
	if (a != nil && b == nil) || (a == nil && b != nil) {
		return 1
	}
	if !a.IP.Equal(b.IP) || a.Port != b.Port || a.Zone != b.Zone {
		return 1
	}
	return 0
}

func CmpNet(a, b net.IPNet) int {
	cmp := bytes.Compare(a.Mask, b.Mask)
	if cmp != 0 {
		return cmp
	}

	return bytes.Compare(a.IP, b.IP)
}

func ContainsNet(outer, inner *net.IPNet) bool {
	outerOnes, _ := outer.Mask.Size()
	innerOnes, _ := inner.Mask.Size()
	return outerOnes <= innerOnes && outer.Contains(inner.IP)
}

func OffsetIP(ip net.IP, off int) net.IP {
	oip := slices.Clone(ip)

	if isV6 := ip.To4() == nil; isV6 {
		num := binary.BigEndian.Uint64(ip[8:])
		binary.BigEndian.PutUint64(oip[8:], num+uint64(off))
	} else {
		num := binary.BigEndian.Uint32(ip[12:])
		binary.BigEndian.PutUint32(oip[12:], num+uint32(off))
	}

	return oip
}
