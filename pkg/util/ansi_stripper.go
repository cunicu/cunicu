package util

import (
	"io"
	"regexp"
)

var stripANSI = regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))")

type ANSIStripper struct {
	io.Writer
}

func (a *ANSIStripper) Write(p []byte) (int, error) {
	line := stripANSI.ReplaceAll(p, []byte{})
	return a.Writer.Write(line)
}
