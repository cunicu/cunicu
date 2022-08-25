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
	runesDigit = "0123456789"

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
	t.FprintKV(wr, "nat type", i.NatType)
	if i.NatType == NATType_NAT_NFTABLES {
		t.FprintKV(wr, "mux ports", fmt.Sprintf("%d, %d", i.MuxPort, i.MuxSrflxPort))
	}

	return nil
}

func (p *ICEPeer) Dump(wr io.Writer, verbosity int) error {
	var v string

	t.FprintKV(wr, "state", p.State)
	t.FprintKV(wr, "proxy type", p.ProxyType)
	t.FprintKV(wr, "reachability", p.Reachability)
	t.FprintKV(wr, "latest state change", util.Ago(p.LastStateChangeTimestamp.Time()))
	t.FprintKV(wr, "restarts", p.Restarts)

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

		t.FprintKV(wr, "\ncandidates")

		wr := util.NewIndenter(wr, "  ")
		wri := util.NewIndenter(wr, "  ")

		if len(p.LocalCandidateStats) > 0 {
			slices.SortFunc(p.LocalCandidateStats, func(a, b *CandidateStats) bool { return a.Priority < b.Priority })

			t.FprintKV(wr, "local")
			for i, cs := range p.LocalCandidateStats {
				cmap[cs.Id] = i
				v = fmt.Sprintf("l%d", i)
				if isNominated := cs.Id == cpsNom.LocalCandidateId; isNominated {
					v = t.Color(v, t.FgRed)
				}
				t.FprintKV(wri, v, cs.ToString())
			}
		}

		if len(p.RemoteCandidateStats) > 0 {
			slices.SortFunc(p.RemoteCandidateStats, func(a, b *CandidateStats) bool { return a.Priority < b.Priority })

			t.FprintKV(wr, "\nremote")
			for i, cs := range p.RemoteCandidateStats {
				cmap[cs.Id] = i
				v = fmt.Sprintf("r%d", i)
				if isNominated := cs.Id == cpsNom.RemoteCandidateId; isNominated {
					v = t.Color(v, t.FgRed)
				}
				t.FprintKV(wri, v, cs.ToString())
			}
		}

		if len(p.CandidatePairStats) > 0 && verbosity > 6 {
			t.FprintKV(wr, "\npairs")
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

				t.FprintKV(wri, v, fmt.Sprintf("l%d <-> r%d, %s",
					lci, rci,
					strings.Join(flags, ", "),
				))
			}
		}
	}

	return nil
}
