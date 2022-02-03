package util

import (
	"fmt"
	"strings"
	"time"

	t "riasc.eu/wice/internal/util/terminal"
)

const (
	day  = 24 * time.Hour
	year = 365 * day
)

func PrettyDuration(left time.Duration, color bool) string {
	out := []string{}
	comps := []struct {
		name    string
		divisor time.Duration
	}{
		{"year", year},
		{"day", day},
		{"hour", time.Hour},
		{"minute", time.Minute},
		{"second", time.Second},
	}

	for _, comp := range comps {
		num := left / comp.divisor

		if num > 0 {
			left -= num * comp.divisor

			unit := comp.name
			if num > 1 {
				unit += "s" // plural s
			}

			if color {
				out = append(out, fmt.Sprintf("%d "+t.Color("%s", t.FgCyan), num, unit))
			} else {
				out = append(out, fmt.Sprintf("%d %s", num, unit))
			}
		}
	}

	return strings.Join(out, ", ")
}

func Ago(ts time.Time, colored bool) string {
	d := time.Since(ts)

	if d < time.Second {
		return "Now"
	}

	return PrettyDuration(d, colored) + " ago"
}

func Every(d time.Duration, color bool) string {
	return "every " + PrettyDuration(d, color)
}

func PrettyBytes(b int64, color bool) string {
	if b < 1024 {
		if color {
			return fmt.Sprintf("%d "+t.Color("B", t.FgCyan), b)
		}
		return fmt.Sprintf("%d B", b)
	}

	var suffices = []rune{'K', 'M', 'G', 'T', 'P', 'E', 'Z', 'Y'}
	var f float32
	var i int

	for i, f = 0, float32(b); i < len(suffices) && f > 1024; i, f = i+1, f/1024 {
	}

	if color {
		return fmt.Sprintf("%.2f "+t.Color("%ciB", t.FgCyan), f, suffices[i-1])
	}

	return fmt.Sprintf("%.2f %ciB", f, suffices[i-1])
}
