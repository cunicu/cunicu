package config

import (
	"errors"
	"reflect"
	"strings"
)

type confmapProvider struct {
	value any
}

// DefaultsProvider creates a new koanf provider which is very
// similar to koanf's confmap provider but slightly adjusted to
// our needs.
func ConfMapProvider(v any) *confmapProvider {
	return &confmapProvider{
		value: v,
	}
}

func (p *confmapProvider) ReadBytes() ([]byte, error) {
	return nil, errors.New("this provider requires no parser")
}

func (p *confmapProvider) Read() (map[string]any, error) {
	return Map(p.value), nil
}

func Map(v any) map[string]any {
	rv := reflect.ValueOf(v)

	return _map(rv).(map[string]any)
}

func _map(v reflect.Value) any {
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	t := v.Type()

	if v.Kind() == reflect.Struct {
		d := map[string]any{}

		for i := 0; i < v.NumField(); i++ {
			fv := v.Field(i)
			sf := t.Field(i)

			if fv.IsValid() && !fv.IsZero() {
				if tag, ok := sf.Tag.Lookup("koanf"); ok {
					name := strings.Split(tag, ",")[0]
					n := _map(fv)
					if name != "" {
						d[name] = n
					} else if m, ok := n.(map[string]any); ok {
						for k, v := range m {
							d[k] = v
						}
					}
				}
			}
		}

		return d
	} else {
		return v.Interface()
	}
}
