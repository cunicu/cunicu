package main

import (
	"os"

	t "github.com/stv0g/cunicu/pkg/util/terminal"
)

func Banner(color bool) string {
	nop := func(s string) string { return s }
	w, o, d := nop, nop, nop

	// Do not use colors during generation of docs
	isDocGen := len(os.Args) > 1 && os.Args[1] == "docs"

	if color && !isDocGen {
		w = func(s string) string {
			return t.Mods(s, t.Bold, t.Color(15))
		}

		o = func(s string) string {
			return t.Mods(s, t.Bold, t.Color(214))
		}

		d = func(s string) string {
			return t.Mods(s, t.Bold, t.Color(130))
		}
	}

	nl := "\n"
	sp := "     "

	return nl +
		sp + w(`  (\(\  `) + sp + o(`▟▀▀▙ █  █ █▀▀▙ ▀▀▀ ▟▀▀▙ █  ▙`) + sp + nl +
		sp + w(`  (-,-) `) + sp + o(`█    █  █ █  █ ▀█  █    █  █`) + sp + w(`(\_/)`) + nl +
		sp + w(`o_(")(")`) + sp + o(`▜▄▄▛ ▜▄▄▛ █  █ ▄█▄ ▜▄▄▛ ▜▄▄▛`) + sp + w(`(•_•)`) + nl +
		sp + w(`        `) + sp + d(`zero-conf • p2p • mesh • vpn`) + sp + w(`/> ❤️  WireGuard™`) + nl + nl
}
