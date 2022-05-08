package config

import (
	"net/url"
	"regexp"
	"strings"
)

type Regexp struct {
	regexp.Regexp
}

type BackendURL struct {
	url.URL
}

func (r *Regexp) UnmarshalText(text []byte) error {
	if re, err := regexp.Compile(string(text)); err != nil {
		return err
	} else {
		r.Regexp = *re
	}

	return nil
}

func (r *Regexp) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (u *BackendURL) UnmarshalText(text []byte) error {
	str := string(text)
	if !strings.Contains(str, ":") {
		str += ":"
	}

	if up, err := url.Parse(str); err != nil {
		return err
	} else {
		u.URL = *up
	}

	return nil
}

func (u BackendURL) MarshalText() ([]byte, error) {
	s := u.String()
	if s[len(s)-1] == ':' {
		s = s[:len(s)-1]
	}

	return []byte(s), nil
}
