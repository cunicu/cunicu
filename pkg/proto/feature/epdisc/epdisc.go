// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package epdisc

import (
	"fmt"
	"io"
	"strings"

	"github.com/pion/ice/v2"
	"github.com/pion/randutil"
	"golang.org/x/exp/slices"

	"github.com/stv0g/cunicu/pkg/log"
	"github.com/stv0g/cunicu/pkg/tty"
)

const (
	lenUFrag = 16
	lenPwd   = 32
)

func NewConnectionState(cs ice.ConnectionState) ConnectionState {
	switch cs {
	case ice.ConnectionStateNew:
		return ConnectionState_NEW
	case ice.ConnectionStateChecking:
		return ConnectionState_CHECKING
	case ice.ConnectionStateConnected:
		return ConnectionState_CONNECTED
	case ice.ConnectionStateCompleted:
		return ConnectionState_COMPLETED
	case ice.ConnectionStateFailed:
		return ConnectionState_FAILED
	case ice.ConnectionStateDisconnected:
		return ConnectionState_DISCONNECTED
	case ice.ConnectionStateClosed:
		return ConnectionState_CLOSED
	default:
		panic("unknown connection state")
	}
}

func NewCredentials() *Credentials {
	ufrag, err := randutil.GenerateCryptoRandomString(lenUFrag, tty.RunesAlpha)
	if err != nil {
		panic(err)
	}

	pwd, err := randutil.GenerateCryptoRandomString(lenPwd, tty.RunesAlpha)
	if err != nil {
		panic(err)
	}

	return &Credentials{
		Ufrag: ufrag,
		Pwd:   pwd,
	}
}

func (i *Interface) Dump(wr io.Writer, level log.Level) error {
	if level.Verbosity() > 4 {
		if _, err := tty.FprintKV(wr, "nat type", i.NatType); err != nil {
			return err
		}

		if i.NatType == NATType_NFTABLES {
			if _, err := tty.FprintKV(wr, "mux ports", fmt.Sprintf("%d, %d", i.MuxPort, i.MuxSrflxPort)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Peer) Dump(wr io.Writer, level log.Level) error { //nolint:gocognit
	var v string

	if p.SelectedCandidatePair != nil {
		if _, err := tty.FprintKV(wr, "candidate pair", p.SelectedCandidatePair.ToString()); err != nil {
			return err
		}
	}

	if level.Verbosity() > 4 {
		if _, err := tty.FprintKV(wr, "proxy type", p.ProxyType); err != nil {
			return err
		}

		if p.LastStateChangeTimestamp != nil {
			if _, err := tty.FprintKV(wr, "latest state change", tty.Ago(p.LastStateChangeTimestamp.Time())); err != nil {
				return err
			}
		}

		if p.Restarts > 0 {
			if _, err := tty.FprintKV(wr, "restarts", p.Restarts); err != nil {
				return err
			}
		}

		if level.Verbosity() > 5 && len(p.CandidatePairStats) > 0 {
			cmap := map[string]int{}
			var cpsNom *CandidatePairStats

			if len(p.CandidatePairStats) > 0 {
				for _, cps := range p.CandidatePairStats {
					if cps.Nominated {
						cpsNom = cps
					}
				}
			}

			if _, err := tty.FprintKV(wr, "\ncandidates"); err != nil {
				return err
			}

			wr := tty.NewIndenter(wr, "  ")
			wri := tty.NewIndenter(wr, "  ")

			if len(p.LocalCandidateStats) > 0 {
				slices.SortFunc(p.LocalCandidateStats, func(a, b *CandidateStats) bool { return a.Priority < b.Priority })

				if _, err := tty.FprintKV(wr, "local"); err != nil {
					return err
				}

				for i, cs := range p.LocalCandidateStats {
					cmap[cs.Id] = i
					v = fmt.Sprintf("l%d", i)
					if isNominated := cs.Id == cpsNom.LocalCandidateId; isNominated {
						v = tty.Mods(v, tty.FgRed)
					}
					if _, err := tty.FprintKV(wri, v, cs.ToString()); err != nil {
						return err
					}
				}
			}

			if len(p.RemoteCandidateStats) > 0 {
				slices.SortFunc(p.RemoteCandidateStats, func(a, b *CandidateStats) bool { return a.Priority < b.Priority })

				if _, err := tty.FprintKV(wr, "\nremote"); err != nil {
					return err
				}

				for i, cs := range p.RemoteCandidateStats {
					cmap[cs.Id] = i
					v = fmt.Sprintf("r%d", i)
					if isNominated := cs.Id == cpsNom.RemoteCandidateId; isNominated {
						v = tty.Mods(v, tty.FgRed)
					}
					if _, err := tty.FprintKV(wri, v, cs.ToString()); err != nil {
						return err
					}
				}
			}

			if len(p.CandidatePairStats) > 0 && level.Verbosity() > 6 {
				if _, err := tty.FprintKV(wr, "\npairs"); err != nil {
					return err
				}

				for i, cps := range p.CandidatePairStats {
					lci := cmap[cps.LocalCandidateId]
					rci := cmap[cps.RemoteCandidateId]

					flags := []string{
						ice.CandidatePairState(cps.State).String(),
					}

					v = fmt.Sprintf("p%d", i)
					if cps.Nominated {
						v = tty.Mods(v, tty.FgRed)
					}

					if cps.Nominated {
						flags = append(flags, "nominated")
					}

					if _, err := tty.FprintKV(wri, v, fmt.Sprintf("l%d <-> r%d, %s",
						lci, rci,
						strings.Join(flags, ", "),
					)); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
