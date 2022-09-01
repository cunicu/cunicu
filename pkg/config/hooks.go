package config

import (
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

type hookBase struct {
	Type string `yaml:"type"`
}

func hookDecodeHook(f, t reflect.Type, data any) (any, error) {
	if f.Kind() != reflect.Map || t.Name() != "HookSetting" {
		return data, nil
	}

	var base hookBase
	if err := mapstructure.Decode(data, &base); err != nil {
		return nil, err
	}

	var hook HookSetting
	switch base.Type {
	case "web":
		hook = &WebHookSetting{
			Method: "POST",
		}
	case "exec":
		hook = &ExecHookSetting{
			Stdin: true,
		}
	default:
		return nil, fmt.Errorf("unknown hook type: %s", base.Type)
	}

	return hook, decode(data, hook)
}
