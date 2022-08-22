package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"go.uber.org/zap/zapcore"
)

type Regexp struct {
	regexp.Regexp
}

func (r *Regexp) UnmarshalText(text []byte) error {
	re, err := regexp.Compile(string(text))
	if err != nil {
		return err
	}

	r.Regexp = *re

	return nil
}

func (r *Regexp) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

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
	return []byte(u.String()), nil
}

type OutputFormat string

const (
	OutputFormatJSON   OutputFormat = "json"
	OutputFormatLogger OutputFormat = "logger"
	OutputFormatHuman  OutputFormat = "human"
)

var (
	OutputFormats = []OutputFormat{
		OutputFormatJSON,
		OutputFormatLogger,
		OutputFormatHuman,
	}
)

func (f *OutputFormat) UnmarshalText(text []byte) error {
	*f = OutputFormat(text)

	switch *f {
	case OutputFormatJSON:
	case OutputFormatLogger:
	case OutputFormatHuman:
		return nil
	}

	return fmt.Errorf("unknown output format: %s", string(text))
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

type Level struct {
	zapcore.Level
}

func (l *Level) Type() string {
	return "string"
}
