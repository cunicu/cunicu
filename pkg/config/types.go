// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type BackendURL struct {
	url.URL
}

func (u *BackendURL) UnmarshalText(text []byte) error {
	str := string(text)
	if !strings.Contains(str, ":") {
		str += ":"
	}

	up, err := url.Parse(str)
	if err != nil {
		return err
	}

	u.URL = *up

	return nil
}

func (u BackendURL) MarshalText() ([]byte, error) {
	s := u.String()
	if s[len(s)-1] == ':' {
		s = s[:len(s)-1]
	}

	return []byte(s), nil
}

type URL struct {
	url.URL
}

func (u *URL) UnmarshalText(text []byte) error {
	up, err := url.Parse(string(text))
	if err != nil {
		return err
	}

	u.URL = *up

	return nil
}

func (u URL) MarshalText() ([]byte, error) {
	s := u.String()
	return []byte(s), nil
}

type OutputFormat string

const (
	OutputFormatJSON   OutputFormat = "json"
	OutputFormatLogger OutputFormat = "logger"
	OutputFormatHuman  OutputFormat = "human"
)

//nolint:gochecknoglobals
var OutputFormats = []OutputFormat{
	OutputFormatJSON,
	OutputFormatLogger,
	OutputFormatHuman,
}

var errUnknownFormat = errors.New("unknown output format")

func (f *OutputFormat) UnmarshalText(text []byte) error {
	*f = OutputFormat(text)

	switch *f {
	case OutputFormatJSON, OutputFormatLogger, OutputFormatHuman:
		return nil
	}

	return fmt.Errorf("%w: %s", errUnknownFormat, string(text))
}

func (f OutputFormat) MarshalText() ([]byte, error) {
	return []byte(f), nil
}

func (f OutputFormat) String() string {
	return string(f)
}

func (f *OutputFormat) Set(str string) error {
	return f.UnmarshalText([]byte(str))
}

func (f *OutputFormat) Type() string {
	return "string"
}
