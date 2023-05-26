// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

var errNoMappingNode = errors.New("no mapping node")

type stringMapSlice []string

func (keys *stringMapSlice) UnmarshalYAML(v *yaml.Node) error {
	if v.Kind != yaml.MappingNode {
		return fmt.Errorf("%w, has %v", errNoMappingNode, v.Kind)
	}

	*keys = make([]string, len(v.Content)/2)
	for i := 0; i < len(v.Content); i += 2 {
		if err := v.Content[i].Decode(&(*keys)[i/2]); err != nil {
			return err
		}
	}

	return nil
}

func ExtractInterfaceOrder(buf []byte) ([]string, error) {
	var s struct {
		Interfaces stringMapSlice `yaml:"interfaces,omitempty"`
	}

	if err := yaml.Unmarshal(buf, &s); err != nil {
		return nil, err
	}

	return s.Interfaces, nil
}
