package terminal

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mdp/qrterminal/v3"
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

func Color(b byte) string {
	return fmt.Sprintf("\x1b[38;5;%dm", b)
}

func Mods(str string, mods ...string) string {
	return strings.Join(mods, "") + str + Reset
}

func FprintKV(wr io.Writer, k string, v ...any) (int, error) {
	if len(v) == 0 {
		return fmt.Fprintf(wr, Mods("%s", Bold)+":\n", k)
	} else if len(v) == 1 {
		return fmt.Fprintf(wr, Mods("%s", Bold)+": %v\n", k, v[0])
	} else if len(v) > 1 {
		return fmt.Fprintf(wr, Mods("%s", Bold)+": %v\n", k, v)
	} else {
		return 0, nil
	}
}

func IsATTY(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		panic(fmt.Errorf("failed to stat stdout: %w", err))
	}

	return (fi.Mode() & os.ModeCharDevice) != 0
}

func QRCode(buf string) {
	wr := NewIndenter(os.Stdout, "  ")

	fmt.Println()
	fmt.Fprint(wr, Color(15))

	qrterminal.GenerateHalfBlock(buf, qrterminal.M, wr)

	fmt.Fprint(wr, Reset)
	fmt.Println()
}
