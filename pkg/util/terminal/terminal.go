package terminal

import (
	"fmt"
	"io"
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

func Color(str string, mods ...string) string {
	return strings.Join(mods, "") + str + Reset
}

func FprintKV(wr io.Writer, k string, v ...any) (int, error) {
	if len(v) == 0 {
		return fmt.Fprintf(wr, Color("%s", Bold)+":\n", k)
	} else if len(v) == 1 {
		return fmt.Fprintf(wr, Color("%s", Bold)+": %v\n", k, v[0])
	} else if len(v) > 1 {
		return fmt.Fprintf(wr, Color("%s", Bold)+": %v\n", k, v)
	} else {
		return 0, nil
	}
}
