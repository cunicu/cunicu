// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
)

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
