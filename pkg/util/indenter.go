package util

import (
	"bytes"
	"io"
)

// NewIndenter returns an io.Writer that prefixes the lines written to it with
// indent and then writes them to w. The writer returns the number of bytes
// written to the underlying Writer.
func NewIndenter(w io.Writer, indent string) io.Writer {
	if indent == "" {
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
	if len(buf) == 0 {
		return 0, nil
	}

	lines := bytes.SplitAfter(buf, []byte{'\n'})
	if len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}

	if !w.partial {
		lines = append([][]byte{{}}, lines...)
	}

	joined := bytes.Join(lines, w.prefix)
	w.partial = joined[len(joined)-1] != '\n'

	if n, err := w.w.Write(joined); err != nil {
		return actualWrittenSize(n, len(w.prefix), lines), err
	}

	return len(buf), nil
}

func actualWrittenSize(underlay, prefix int, lines [][]byte) int {
	actual := 0
	remain := underlay
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		addition := remain - prefix
		if addition <= 0 {
			return actual
		}

		if addition <= len(line) {
			return actual + addition
		}

		actual += len(line)
		remain -= prefix + len(line)
	}

	return actual
}
