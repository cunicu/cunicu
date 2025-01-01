// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"cunicu.li/cunicu/pkg/tty"
)

func Banner(color bool) string {
	nop := func(s string) string { return s }
	w, o, d := nop, nop, nop

	if color {
		w = func(s string) string {
			return tty.Mods(s, tty.Bold, tty.Color256(15))
		}

		o = func(s string) string {
			return tty.Mods(s, tty.Bold, tty.Color256(214))
		}

		d = func(s string) string {
			return tty.Mods(s, tty.Bold, tty.Color256(130))
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
