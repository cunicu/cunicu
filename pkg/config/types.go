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

type OutputFormat int

const (
	OutputFormatJSON OutputFormat = iota
	OutputFormatLogger
	OutputFormatHuman
)

var (
	OutputFormatNames = []string{"json", "logger", "human"}
)

func (f *OutputFormat) UnmarshalText(text []byte) error {
	for i, of := range OutputFormatNames {
		if of == string(text) {
			*f = OutputFormat(i)
			return nil
		}
	}

	return fmt.Errorf("unknown output format: %s", string(text))
}

func (f OutputFormat) MarshalText() ([]byte, error) {
	return []byte(OutputFormatNames[int(f)]), nil
}

func (f OutputFormat) String() string {
	b, err := f.MarshalText()
	if err != nil {
		panic(fmt.Errorf("failed marshal: %w", err))
	}

	return string(b)
}

func (f OutputFormat) Set(str string) error {
	return f.UnmarshalText([]byte(str))
}

func (f OutputFormat) Type() string {
	return "string"
}

type Level struct {
	zapcore.Level
}

func (l *Level) Type() string {
	return "string"
}
