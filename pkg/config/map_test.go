// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"time"

	"github.com/pion/ice/v2"

	"github.com/stv0g/cunicu/pkg/config"
	icex "github.com/stv0g/cunicu/pkg/ice"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = It("map", func() {
	s := config.Settings{
		WatchInterval: 5 * time.Second,
		Interfaces: map[string]config.InterfaceSettings{
			"wg0": {
				ICE: config.ICESettings{
					NetworkTypes: []icex.NetworkType{
						{NetworkType: ice.NetworkTypeTCP4},
						{NetworkType: ice.NetworkTypeTCP6},
					},
				},
				HostName: "test",
			},
		},
		DefaultInterfaceSettings: config.InterfaceSettings{
			HostName: "test2",
			Hooks: []config.HookSetting{
				&config.ExecHookSetting{
					BaseHookSetting: config.BaseHookSetting{
						Type: "exec",
					},
					Command: "dummy",
					Env: map[string]string{
						"ENV1": "value1",
					},
				},
			},
		},
	}

	m := config.Map(s, "koanf")

	Expect(m).To(Equal(map[string]any{
		"watch_interval": "5s",
		"interfaces": map[string]any{
			"wg0": map[string]any{
				"ice": map[string]any{
					"network_types": []any{"tcp4", "tcp6"},
				},
				"hostname": "test",
			},
		},
		"hooks": []any{
			s.DefaultInterfaceSettings.Hooks[0],
		},
		"hostname": "test2",
	}))
})
