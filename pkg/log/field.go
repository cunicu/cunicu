// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"reflect"
	"strings"

	"golang.org/x/exp/slices"
)

func getFieldName(field reflect.StructField) (string, bool) {
	if value := getFirstTagValue(field.Tag, "json", "toml", "mapstructure", "koanf"); value != "" {
		parts := strings.Split(value, ",")
		return parts[0], slices.Contains(parts, "omitempty")
	}

	return strings.ToLower(field.Name), true
}

func getFirstTagValue(tag reflect.StructTag, searchTags ...string) string {
	for _, searchTag := range searchTags {
		if value, ok := tag.Lookup(searchTag); ok {
			return value
		}
	}

	return ""
}
