// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package tty

import (
	"bytes"
	"io"
)

// NewIndenter returns an io.Writer that prefixes the lines written to it with
// indent and then writes them to w. The writer returns the number of bytes
// written to the underlying Writer.
func NewIndenter(w io.Writer, indent string) io.Writer {
	if len(indent) == 0 {
		return w
	}

	return &indenter{
		w:      w,
		prefix: []byte(indent),
	}
}

type indenter struct {
	w       io.Writer
	prefix  []byte
	partial bool // true if next line's indent already written
}

// Write implements io.Writer.
func (w *indenter) Write(buf []byte) (int, error) {
	newline := []byte{'\n'}

	if len(buf) == 0 {
		return 0, nil
	}

	lines := bytes.SplitAfter(buf, newline)
	if len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}

	n := 0
	write := func(b []byte) error {
		m, err := w.w.Write(b)
		if err != nil {
			return err
		}

		n += m
		return nil
	}

	if !w.partial {
		if err := write(w.prefix); err != nil {
			return -1, err
		}
	}

	for i, line := range lines {
		if err := write(line); err != nil {
			return -1, err
		}

		if isLast := i+1 == len(lines); !isLast {
			if err := write(w.prefix); err != nil {
				return -1, err
			}
		}
	}

	w.partial = !bytes.HasSuffix(lines[len(lines)-1], newline)

	return n, nil
}
