package epdisc

import (
	"fmt"
	"io"
	"strings"

	"github.com/pion/ice/v2"
	"github.com/pion/randutil"
	"golang.org/x/exp/slices"

	"github.com/stv0g/cunicu/pkg/util"
	t "github.com/stv0g/cunicu/pkg/util/terminal"

	icex "github.com/stv0g/cunicu/pkg/ice"
)

const (
	lenUFrag = 16
	lenPwd   = 32
)

func NewCredentials() Credentials {
	ufrag, err := randutil.GenerateCryptoRandomString(lenUFrag, util.RunesAlpha)
	if err != nil {
		panic(err)
	}

	pwd, err := randutil.GenerateCryptoRandomString(lenPwd, util.RunesAlpha)
	if err != nil {
		panic(err)
	}

	return Credentials{
		Ufrag: ufrag,
		Pwd:   pwd,
	}
}

func (i *Interface) Dump(wr io.Writer, verbosity int) error {
	if verbosity > 4 {
		if _, err := t.FprintKV(wr, "nat type", i.NatType); err != nil {
			return err
		}

		if i.NatType == NATType_NAT_NFTABLES {
			if _, err := t.FprintKV(wr, "mux ports", fmt.Sprintf("%d, %d", i.MuxPort, i.MuxSrflxPort)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Peer) Dump(wr io.Writer, verbosity int) error {
	var v string

	if _, err := t.FprintKV(wr, "state", t.Mods(p.State.String(), t.Bold, p.State.Color())); err != nil {
		return err
	}

	if p.SelectedCandidatePair != nil {
		if _, err := t.FprintKV(wr, "candidate pair", p.SelectedCandidatePair.ToString()); err != nil {
			return err
		}
	}

	if verbosity > 4 {
		if _, err := t.FprintKV(wr, "proxy type", p.ProxyType); err != nil {
			return err
		}

		if _, err := t.FprintKV(wr, "latest state change", util.Ago(p.LastStateChangeTimestamp.Time())); err != nil {
			return err
		}

		if p.Restarts > 0 {
			if _, err := t.FprintKV(wr, "restarts", p.Restarts); err != nil {
				return err
			}
		}

		if verbosity > 5 && len(p.CandidatePairStats) > 0 {
			var cmap = map[string]int{}
			var cpsNom *CandidatePairStats

			if len(p.CandidatePairStats) > 0 {
				for _, cps := range p.CandidatePairStats {
					if cps.Nominated {
						cpsNom = cps
					}
				}
			}

			if _, err := t.FprintKV(wr, "\ncandidates"); err != nil {
				return err
			}

			wr := t.NewIndenter(wr, "  ")
			wri := t.NewIndenter(wr, "  ")

			if len(p.LocalCandidateStats) > 0 {
				slices.SortFunc(p.LocalCandidateStats, func(a, b *CandidateStats) bool { return a.Priority < b.Priority })

				if _, err := t.FprintKV(wr, "local"); err != nil {
					return err
				}

				for i, cs := range p.LocalCandidateStats {
					cmap[cs.Id] = i
					v = fmt.Sprintf("l%d", i)
					if isNominated := cs.Id == cpsNom.LocalCandidateId; isNominated {
						v = t.Mods(v, t.FgRed)
					}
					if _, err := t.FprintKV(wri, v, cs.ToString()); err != nil {
						return err
					}
				}
			}

			if len(p.RemoteCandidateStats) > 0 {
				slices.SortFunc(p.RemoteCandidateStats, func(a, b *CandidateStats) bool { return a.Priority < b.Priority })

				if _, err := t.FprintKV(wr, "\nremote"); err != nil {
					return err
				}

				for i, cs := range p.RemoteCandidateStats {
					cmap[cs.Id] = i
					v = fmt.Sprintf("r%d", i)
					if isNominated := cs.Id == cpsNom.RemoteCandidateId; isNominated {
						v = t.Mods(v, t.FgRed)
					}
					if _, err := t.FprintKV(wri, v, cs.ToString()); err != nil {
						return err
					}
				}
			}

			if len(p.CandidatePairStats) > 0 && verbosity > 6 {
				if _, err := t.FprintKV(wr, "\npairs"); err != nil {
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
						v = t.Mods(v, t.FgRed)
					}

					if cps.Nominated {
						flags = append(flags, "nominated")
					}

					if _, err := t.FprintKV(wri, v, fmt.Sprintf("l%d <-> r%d, %s",
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

func NewConnectionState(s icex.ConnectionState) ConnectionState {
	switch s {
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

	case icex.ConnectionStateCreating:
		return ConnectionState_CREATING
	case icex.ConnectionStateClosing:
		return ConnectionState_CLOSING
	case icex.ConnectionStateConnecting:
		return ConnectionState_CONNECTING
	case icex.ConnectionStateIdle:
		return ConnectionState_IDLE
	}

	return -1
}

func (s *ConnectionState) ConnectionState() icex.ConnectionState {
	switch *s {
	case ConnectionState_NEW:
		return ice.ConnectionStateNew
	case ConnectionState_CHECKING:
		return ice.ConnectionStateChecking
	case ConnectionState_CONNECTED:
		return ice.ConnectionStateConnected
	case ConnectionState_COMPLETED:
		return ice.ConnectionStateCompleted
	case ConnectionState_FAILED:
		return ice.ConnectionStateFailed
	case ConnectionState_DISCONNECTED:
		return ice.ConnectionStateDisconnected
	case ConnectionState_CLOSED:
		return ice.ConnectionStateClosed

	case ConnectionState_CREATING:
		return icex.ConnectionStateCreating
	case ConnectionState_CLOSING:
		return icex.ConnectionStateClosing
	case ConnectionState_CONNECTING:
		return icex.ConnectionStateConnecting
	case ConnectionState_IDLE:
		return icex.ConnectionStateIdle
	}

	return -1
}

func (s *ConnectionState) Color() string {
	switch *s {
	case ConnectionState_CHECKING:
		return t.FgYellow
	case ConnectionState_CONNECTED:
		return t.FgGreen
	case ConnectionState_FAILED,
		ConnectionState_DISCONNECTED:
		return t.FgRed
	case ConnectionState_NEW,
		ConnectionState_COMPLETED,
		ConnectionState_CLOSED,
		ConnectionState_CREATING,
		ConnectionState_CLOSING,
		ConnectionState_CONNECTING,
		ConnectionState_IDLE:
		return t.FgWhite
	}

	return t.FgDefault
}

func (s *ConnectionState) MarshalText() ([]byte, error) {
	return []byte(strings.ToLower(s.String())), nil
}
