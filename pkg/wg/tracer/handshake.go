// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build tracer

package tracer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/cilium/ebpf/ringbuf"
	"golang.org/x/sys/unix"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Handshake struct {
	bpfHandshake
}

func (hs *Handshake) Time() time.Time {
	if bootTime.IsZero() {
		var bt unix.Timespec

		if err := unix.ClockGettime(unix.CLOCK_BOOTTIME, &bt); err != nil {
			panic(err)
		}

		btt := time.Duration(bt.Sec*1e9 + bt.Sec)
		now := time.Now()

		bootTime = now.Add(-btt)
	}

	return bootTime.Add(time.Duration(hs.Ktime))
}

func (hs *Handshake) LocalStaticPrivateKey() wgtypes.Key {
	return hs.bpfHandshake.LocalStaticPrivateKey
}

func (hs *Handshake) LocalEphemeralPrivateKey() wgtypes.Key {
	return hs.bpfHandshake.LocalEphemeralPrivateKey
}

func (hs *Handshake) RemoteStaticPublicKey() wgtypes.Key {
	return hs.bpfHandshake.RemoteStaticPublicKey
}

func (hs *Handshake) PresharedKey() wgtypes.Key {
	return hs.bpfHandshake.PresharedKey
}

func HandshakeFromBPFRecord(record ringbuf.Record) (*Handshake, error) {
	hs := &Handshake{}

	// Parse the ringbuf event entry into a bpfHandshake structure.
	if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &hs.bpfHandshake); err != nil {
		return nil, fmt.Errorf("failed to parse handshake from ringbuf: %w", err)
	}

	return hs, nil
}

func (hs *Handshake) DumpKeyLog(wr io.Writer) error {
	if _, err := fmt.Fprintf(wr, "LOCAL_STATIC_PRIVATE_KEY=%s\n", hs.LocalStaticPrivateKey()); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(wr, "REMOTE_STATIC_PUBLIC_KEY=%s\n", hs.RemoteStaticPublicKey()); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(wr, "LOCAL_EPHEMERAL_PRIVATE_KEY=%s\n", hs.LocalEphemeralPrivateKey()); err != nil {
		return err
	}

	zeroKey := wgtypes.Key{}
	if hs.PresharedKey() != zeroKey {
		fmt.Fprintf(wr, "PRESHARED_KEY=%s\n", hs.PresharedKey())
	}

	return nil
}
