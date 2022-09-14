package config

import (
	"reflect"
	"strings"
)

const delim = "."

type ConfigChangedHandler interface {
	OnConfigChanged(path string, old, new any)
}

type Meta struct {
	Fields map[string]*Meta

	Type reflect.Type

	OnChangedHandlers []ConfigChangedHandler
}

func Metadata() *Meta {
	return metadata(reflect.TypeOf(Settings{}))
}

func metadata(typ reflect.Type) *Meta {
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	m := &Meta{
		Type: typ,
	}

	if typ.Kind() == reflect.Struct {
		m.Fields = map[string]*Meta{}

		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			if tag, ok := field.Tag.Lookup("koanf"); ok {
				name := strings.Split(tag, ",")[0]
				n := metadata(field.Type)
				if name != "" {
					m.Fields[name] = n
				} else {
					for k, v := range n.Fields {
						m.Fields[k] = v
					}
				}
			}
		}
	}

	return m
}

func (m *Meta) Keys() []string {
	keys := []string{}

	for k, v := range m.Fields {
		if v.Fields == nil {
			keys = append(keys, k)
		} else {
			for _, p := range v.Keys() {
				keys = append(keys, k+delim+p)
			}
		}
	}

	return keys
}

func (m *Meta) Lookup(path string) *Meta {
	parts := strings.Split(path, delim)
	return m.lookup(parts)
}

func (m *Meta) lookup(path []string) *Meta {
	if len(path) == 0 {
		return m
	} else {
		if m.Fields != nil {
			if n, ok := m.Fields[path[0]]; ok {
				return n.lookup(path[1:])
			}
		}
	}

	return nil
}
