package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/pion/ice/v2"
	"go.uber.org/zap"
	"riasc.eu/wice/pkg/proxy"
)

type backendURLList []*url.URL

func (i *backendURLList) Type() string {
	return "stringSlice"
}

func (i *backendURLList) String() string {
	s := []string{}
	for _, u := range *i {
		s = append(s, u.String())
	}

	return strings.Join(s, ",")
}

func (i *backendURLList) Set(value string) error {
	for _, value := range strings.Split(value, ",") {
		// Allow the user to specify just the backend type as a valid url.
		// E.g. "p2p" instead of "p2p:"
		if !strings.Contains(value, ":") {
			value += ":"
		}

		uri, err := url.Parse(value)
		if err != nil {
			return fmt.Errorf("invalid backend URI: %w", err)
		}

		*i = append(*i, uri)
	}

	return nil
}

type iceURLList []*ice.URL

func (ul *iceURLList) Type() string {
	return "stringSlice"
}

func (ul *iceURLList) Set(value string) error {
	for _, value := range strings.Split(value, ",") {
		u, err := ice.ParseURL(value)
		if err != nil {
			return err
		}

		*ul = append(*ul, u)
	}

	return nil
}

func (ul *iceURLList) String() string {
	l := []string{}

	for _, u := range *ul {
		l = append(l, u.String())
	}

	return strings.Join(l, ",")
}

type candidateTypeList []ice.CandidateType

func (cl *candidateTypeList) Type() string {
	return "stringSlice"
}

func (cl *candidateTypeList) Set(value string) error {
	for _, value := range strings.Split(value, ",") {
		ct, err := candidateTypeFromString(value)
		if err != nil {
			return err
		}

		*cl = append(*cl, ct)
	}

	return nil
}

func (cl *candidateTypeList) String() string {
	l := []string{}

	for _, c := range *cl {
		l = append(l, c.String())
	}

	return strings.Join(l, ",")
}

type networkTypeList []ice.NetworkType

func (nl *networkTypeList) Type() string {
	return "stringSlice"
}

func (nl *networkTypeList) Set(value string) error {
	for _, value := range strings.Split(value, ",") {
		ct, err := networkTypeFromString(value)
		if err != nil {
			return err
		}

		*nl = append(*nl, ct)
	}

	return nil
}

func (nl *networkTypeList) String() string {
	l := []string{}

	for _, c := range *nl {
		l = append(l, c.String())
	}

	return strings.Join(l, ",")
}

type proxyType struct{ proxy.ProxyType }

func (pt *proxyType) Type() string {
	return "string"
}

func (pt *proxyType) Set(value string) error {
	var err error
	pt.ProxyType, err = proxy.ProxyTypeFromString(value)
	return err
}

type logLevel struct{ zap.AtomicLevel }

func (ll *logLevel) Type() string {
	return "string"
}

func (ll *logLevel) Set(value string) error {
	return ll.UnmarshalText([]byte(value))
}

type regex struct{ *regexp.Regexp }

func (re *regex) Type() string {
	return "string"
}

func (re *regex) Set(value string) error {
	r, err := regexp.Compile(value)
	if err != nil {
		return fmt.Errorf("failed to compile regex: %w", err)
	}

	re.Regexp = r

	return nil
}
