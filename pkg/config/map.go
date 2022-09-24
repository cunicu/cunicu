package config

import (
	"reflect"
	"strings"
)

var configPkgPath = reflect.TypeOf(Settings{}).PkgPath()

func Map(v any, tagName string) map[string]any {
	rv := reflect.ValueOf(v)

	return _map(rv, tagName).(map[string]any)
}

func _map(v reflect.Value, tagName string) any {
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	t := v.Type()

	// Types outside the config package will be taken as an interface
	if t.PkgPath() != configPkgPath && t.PkgPath() != "" {
		return v.Interface()
	}

	switch v.Kind() {
	case reflect.Struct:
		d := map[string]any{}

		for i := 0; i < v.NumField(); i++ {
			fv := v.Field(i)
			sf := t.Field(i)

			if fv.IsValid() && !fv.IsZero() {
				if tag, ok := sf.Tag.Lookup(tagName); ok {
					name := strings.Split(tag, ",")[0]
					n := _map(fv, tagName)
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

	case reflect.Map:
		if t.Key().Kind() == reflect.String {
			d := map[string]any{}

			for _, e := range v.MapKeys() {
				mv := v.MapIndex(e)

				d[e.String()] = _map(mv, tagName)
			}

			return d
		} else {
			return v.Interface()
		}

	default:
		return v.Interface()
	}
}
