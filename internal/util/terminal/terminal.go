package terminal

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

const (
	FgBlack     = "\x1b[30m"
	FgRed       = "\x1b[31m"
	FgGreen     = "\x1b[32m"
	FgYellow    = "\x1b[33m"
	FgBlue      = "\x1b[34m"
	FgMagenta   = "\x1b[35m"
	FgCyan      = "\x1b[36m"
	FgWhite     = "\x1b[37m"
	FgDefault   = "\x1b[39m"
	BgBlack     = "\x1b[40m"
	BgRed       = "\x1b[41m"
	BgGreen     = "\x1b[42m"
	BgYellow    = "\x1b[43m"
	BgBlue      = "\x1b[44m"
	BgMagenta   = "\x1b[45m"
	BgCyan      = "\x1b[46m"
	BgWhite     = "\x1b[47m"
	BgDefault   = "\x1b[49m"
	Bold        = "\x1b[1m"
	NoBold      = "\x1b[22m"
	Underline   = "\x1b[4m"
	NoUnderline = "\x1b[24m"
	Reset       = "\x1b[0m"
)

var (
	reANSIEscape = regexp.MustCompile(`(?mi)\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])`)
)

func FprintfColored(wr io.Writer, color bool, f string, args ...interface{}) (int, error) {
	if !color {
		f = reANSIEscape.ReplaceAllString(f, "")
	}

	return fmt.Fprintf(wr, f, args...)
}

func Color(str string, mods ...string) string {
	return strings.Join(mods, "") + str + Reset
}

func PrintKeyValues(wr io.Writer, color bool, prefix string, kv map[string]interface{}) (int, error) {
	n := 0
	for k, v := range kv {
		if b, err := FprintfColored(wr, color, "%s"+Color("%s", Bold)+": %v\n", prefix, k, v); err != nil {
			return b, err
		} else {
			n += b
		}
	}

	return n, nil
}
