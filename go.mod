// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

module cunicu.li/cunicu

go 1.23.0

toolchain go1.23.1

require (
	dario.cat/mergo v1.0.1
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/cilium/ebpf v0.16.0
	github.com/dchest/siphash v1.2.3
	github.com/fsnotify/fsnotify v1.7.0
	github.com/google/nftables v0.2.0
	github.com/knadh/koanf/maps v0.1.1
	github.com/knadh/koanf/parsers/yaml v0.1.0
	github.com/knadh/koanf/providers/confmap v0.1.0
	github.com/knadh/koanf/providers/env v1.0.0
	github.com/knadh/koanf/providers/file v1.1.2
	github.com/knadh/koanf/providers/posflag v0.1.0
	github.com/knadh/koanf/providers/rawbytes v0.1.0
	github.com/knadh/koanf/v2 v2.1.1
	github.com/mdp/qrterminal/v3 v3.2.0
	github.com/miekg/dns v1.1.62
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pion/ice/v3 v3.0.16
	github.com/pion/ice/v4 v4.0.1
	github.com/pion/logging v0.2.2
	github.com/pion/randutil v0.1.0
	github.com/pion/stun/v2 v2.0.0
	github.com/pion/stun/v3 v3.0.0
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/vishvananda/netlink v1.3.0
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.27.0
	golang.org/x/sync v0.8.0
	golang.org/x/sys v0.25.0
	golang.zx2c4.com/wireguard v0.0.0-20231211153847-12269c276173
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20230429144221-925a1e7659e6
	google.golang.org/grpc v1.67.0
	google.golang.org/protobuf v1.34.2
	gopkg.in/ini.v1 v1.67.0
	gopkg.in/yaml.v3 v3.0.1
	kernel.org/pub/linux/libs/security/libcap/cap v1.2.70
)

require (
	cunicu.li/gont/v2 v2.6.18 // test-only
	github.com/foxcpp/go-mockdns v1.1.0 // test-only
	github.com/gopacket/gopacket v1.3.0 // test-only
	github.com/onsi/ginkgo/v2 v2.20.2 // test-only
	github.com/onsi/gomega v1.34.2 // test-only
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/go-delve/delve v1.21.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-ping/ping v1.1.0 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.0.0-alpha.1 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/go-dap v0.10.0 // indirect
	github.com/google/pprof v0.0.0-20240827171923-fa2c70bbbfe5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/native v1.1.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mdlayher/genetlink v1.3.2 // indirect
	github.com/mdlayher/netlink v1.7.2 // indirect
	github.com/mdlayher/socket v0.5.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/pion/dtls/v2 v2.2.12 // indirect
	github.com/pion/mdns/v2 v2.0.7 // indirect
	github.com/pion/transport/v2 v2.2.4 // indirect
	github.com/pion/transport/v3 v3.0.7 // indirect
	github.com/pion/turn/v3 v3.0.3 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.9.2 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	github.com/wlynxg/anet v0.0.3 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/arch v0.3.0 // indirect
	golang.org/x/exp v0.0.0-20240909161429-701f63a606c0 // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/term v0.24.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	golang.org/x/tools v0.25.0 // indirect
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	kernel.org/pub/linux/libs/security/libcap/psx v1.2.70 // indirect
	rsc.io/qr v0.2.0 // indirect
)
