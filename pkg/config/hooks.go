// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
	"net"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

var errUnknownHookType = errors.New("unknown hook type")

func hookDecodeHook(f, t reflect.Type, data any) (any, error) {
	if f.Kind() != reflect.Map {
		return data, nil
	}

	if t.Name() != "HookSetting" {
		return data, nil
	}

	var base BaseHookSetting
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
		return nil, fmt.Errorf("%w: %s", errUnknownHookType, base.Type)
	}

	decoder, err := mapstructure.NewDecoder(DecoderConfig(hook))
	if err != nil {
		return nil, err
	}

	return hook, decoder.Decode(data)
}

// stringToIPAddrHook is a DecodeHookFunc that converts strings to net.IPAddr
func stringToIPAddrHook(
	f reflect.Type,
	t reflect.Type,
	data interface{},
) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}
	if t != reflect.TypeOf(net.IPAddr{}) {
		return data, nil
	}

	// Convert it by parsing
	ip, err := net.ResolveIPAddr("ip", data.(string))
	return *ip, err
}

// StringToIPNetHookFunc returns a DecodeHookFunc that converts
// strings to IPNetAddr
func stringToIPNetAddrHookFunc(
	f reflect.Type,
	t reflect.Type,
	data interface{},
) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}
	if t != reflect.TypeOf(net.IPNet{}) {
		return data, nil
	}

	// Convert it by parsing
	str, ok := data.(string)
	if !ok {
		panic("type assertion failed")
	}

	ip, net, err := net.ParseCIDR(str)
	if err != nil {
		return nil, err
	}

	net.IP = ip

	return net, nil
}
