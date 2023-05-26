// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build tracer

// Package tracer uses Linux Kprobes to gather ephemeral keys from handshakes of local WireGuard interfaces
//
// Tested with Linux 5.15.0
package tracer

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
	"time"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
	"github.com/stv0g/cunicu/pkg/wg/tracer/kernel"
)

//go:generate make -C kernel config.h
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -tags tracer -cc clang -type handshake -target $GOARCH bpf kprobe_wg_index_hashtable_insert.c -- -I include

var bootTime time.Time

type HandshakeTracer struct {
	Handshakes chan *Handshake
	Errors     chan error

	reader *ringbuf.Reader
	kprobe link.Link
}

func NewHandshakeTracer() (*HandshakeTracer, error) {
	var err error

	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		return nil, fmt.Errorf("failed to remove memlock: %w", err)
	}

	ht := &HandshakeTracer{
		Errors:     make(chan error),
		Handshakes: make(chan *Handshake),
	}

	if err := ht.Check(); err != nil {
		return nil, err
	}

	// Load pre-compiled programs and maps into the kernel.
	objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		return nil, fmt.Errorf("failed loading objects: %w", err)
	}
	defer objs.Close()

	// Open a Kprobe at the entry point of the kernel function and attach the
	// pre-compiled program. Each time the kernel function enters, the program
	// will emit an event containing pid and command of the execved task.
	if ht.kprobe, err = link.Kprobe("wg_index_hashtable_insert", objs.KprobeWgIndexHashtableInsert, nil); err != nil {
		return nil, fmt.Errorf("failed opening kprobe: %w", err)
	}

	// Open a ringbuf reader from userspace RINGBUF map described in the
	// eBPF C program.
	ht.reader, err = ringbuf.NewReader(objs.Handshakes)
	if err != nil {
		return nil, fmt.Errorf("failed opening ringbuf reader: %w", err)
	}

	go ht.run()

	return ht, nil
}

func charsToString(is []int8) string {
	bs := []byte{}
	for i := 0; is[i] != 0; i++ {
		bs = append(bs, byte(is[i]))
	}
	return string(bs)
}

func (ht *HandshakeTracer) Check() error {
	uts := &syscall.Utsname{}
	if err := syscall.Uname(uts); err != nil {
		return fmt.Errorf("failed to get utsname: %w", err)
	}

	machine := charsToString(uts.Machine[:])
	release := charsToString(uts.Release[:])

	if machine != kernel.TargetMachine {
		return fmt.Errorf("machine mismatch: %s != %s", machine, kernel.TargetMachine)
	}

	if !strings.HasPrefix(release, kernel.TargetRelease) {
		return fmt.Errorf("release mismatch: %s != %s", release, kernel.TargetRelease)
	}

	return nil
}

func (ht *HandshakeTracer) Close() error {
	if err := ht.reader.Close(); err != nil {
		return fmt.Errorf("failed closing ringbuf reader: %w", err)
	}

	if err := ht.kprobe.Close(); err != nil {
		return fmt.Errorf("failed to close kprobe: %w", err)
	}

	return nil
}

func (ht *HandshakeTracer) run() {
	for {
		record, err := ht.reader.Read()
		if err != nil {
			if errors.Is(err, ringbuf.ErrClosed) {
				return
			}

			ht.Errors <- fmt.Errorf("failed to read from ringbuf: %w", err)
			continue
		}

		if hs, err := HandshakeFromBPFRecord(record); err != nil {
			ht.Errors <- fmt.Errorf("failed to parse handshake from record: %w", err)
		} else {
			ht.Handshakes <- hs
		}
	}
}
