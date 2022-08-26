package pb

import (
	"fmt"
	"io"
	"strings"

	"github.com/pion/ice/v2"
	"github.com/pion/randutil"
	"golang.org/x/exp/slices"
	"riasc.eu/wice/pkg/util"
	t "riasc.eu/wice/pkg/util/terminal"
)

const (
	runesAlpha = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	lenUFrag = 16
	lenPwd   = 32
)

func NewCredentials() Credentials {
	ufrag, err := randutil.GenerateCryptoRandomString(lenUFrag, runesAlpha)
	if err != nil {
		panic(err)
	}

	pwd, err := randutil.GenerateCryptoRandomString(lenPwd, runesAlpha)
	if err != nil {
		panic(err)
	}

	return Credentials{
		Ufrag: ufrag,
		Pwd:   pwd,
	}
}

func (i *ICEInterface) Dump(wr io.Writer, verbosity int) error {
	if _, err := t.FprintKV(wr, "nat type", i.NatType); err != nil {
		return err
	}

	if i.NatType == NATType_NAT_NFTABLES {
		if _, err := t.FprintKV(wr, "mux ports", fmt.Sprintf("%d, %d", i.MuxPort, i.MuxSrflxPort)); err != nil {
			return err
		}
	}

	return nil
}

func (p *ICEPeer) Dump(wr io.Writer, verbosity int) error {
	var v string

	if _, err := t.FprintKV(wr, "state", t.Color(p.State.String(), t.Bold, p.State.Color())); err != nil {
		return err
	}

	if _, err := t.FprintKV(wr, "proxy type", p.ProxyType); err != nil {
		return err
	}

	if _, err := t.FprintKV(wr, "reachability", p.Reachability); err != nil {
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
					v = t.Color(v, t.FgRed)
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
					v = t.Color(v, t.FgRed)
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
					v = t.Color(v, t.FgRed)
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

	return nil
}
