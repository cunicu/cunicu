// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package tty

import (
	"io"
	"regexp"
)

var stripANSI = regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))")

type ansiStripper struct {
	io.Writer
}

func NewANSIStripper(wr io.Writer) io.Writer {
	return &ansiStripper{
		Writer: wr,
	}
}

func (a *ansiStripper) Write(p []byte) (int, error) {
	line := stripANSI.ReplaceAll(p, []byte{})
	return a.Writer.Write(line)
}

type WriteSyncer interface {
	io.Writer
	Sync() error
}

type ansiStripperSynced struct {
	WriteSyncer
}

func NewANSIStripperSynced(wr WriteSyncer) WriteSyncer {
	return &ansiStripperSynced{
		WriteSyncer: wr,
	}
}

func (a *ansiStripperSynced) Write(p []byte) (int, error) {
	line := stripANSI.ReplaceAll(p, []byte{})
	return a.WriteSyncer.Write(line)
}

func StripANSI(s string) string {
	return stripANSI.ReplaceAllString(s, "")
}
