// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

module github.com/stv0g/cunicu

go 1.20

require (
	github.com/cenkalti/backoff/v4 v4.2.1
	github.com/cilium/ebpf v0.10.0
	github.com/dchest/siphash v1.2.3
	github.com/fsnotify/fsnotify v1.6.0
	github.com/google/nftables v0.1.0
	github.com/imdario/mergo v0.3.16
	github.com/knadh/koanf/maps v0.1.1
	github.com/knadh/koanf/parsers/yaml v0.1.0
	github.com/knadh/koanf/providers/confmap v0.1.0
	github.com/knadh/koanf/providers/env v0.1.0
	github.com/knadh/koanf/providers/file v0.1.0
	github.com/knadh/koanf/providers/posflag v0.1.0
	github.com/knadh/koanf/providers/rawbytes v0.1.0
	github.com/knadh/koanf/v2 v2.0.1
	github.com/mdp/qrterminal/v3 v3.1.1
	github.com/miekg/dns v1.1.55
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pion/ice/v2 v2.3.8
	github.com/pion/logging v0.2.2
	github.com/pion/randutil v0.1.0
	github.com/pion/stun v0.6.1
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/vishvananda/netlink v1.2.1-beta.2.0.20221214185949-378a404a26f0
	go.uber.org/zap v1.24.0
	golang.org/x/crypto v0.10.0
	golang.org/x/exp v0.0.0-20230626212559-97b1e661b5df
	golang.org/x/sync v0.3.0
	golang.org/x/sys v0.9.0
	golang.zx2c4.com/wireguard v0.0.0-20230627160250-1f25eac395d6
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20230429144221-925a1e7659e6
	google.golang.org/grpc v1.56.1
	google.golang.org/protobuf v1.30.0
	gopkg.in/ini.v1 v1.67.0
	gopkg.in/yaml.v3 v3.0.1
	kernel.org/pub/linux/libs/security/libcap/cap v1.2.69
)

require (
	github.com/foxcpp/go-mockdns v1.0.0 // test-only
	github.com/gopacket/gopacket v1.1.1-0.20230627101822-4472aeca4ce4 // test-only
	github.com/onsi/ginkgo/v2 v2.10.0 // test-only
	github.com/onsi/gomega v1.27.8 // test-only
	github.com/stv0g/gont/v2 v2.3.6 // test-only
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fxamacker/cbor/v2 v2.4.0 // indirect
	github.com/go-delve/delve v1.20.2 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-ping/ping v1.1.0 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/go-dap v0.8.0 // indirect
	github.com/google/pprof v0.0.0-20230510103437-eeec1cb781c3 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/native v1.1.0 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mdlayher/genetlink v1.3.2 // indirect
	github.com/mdlayher/netlink v1.7.2 // indirect
	github.com/mdlayher/socket v0.4.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/pion/dtls/v2 v2.2.7 // indirect
	github.com/pion/mdns v0.0.7 // indirect
	github.com/pion/transport/v2 v2.2.1 // indirect
	github.com/pion/turn/v2 v2.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.9.2 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/goleak v1.2.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/arch v0.3.0 // indirect
	golang.org/x/mod v0.11.0 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/text v0.10.0 // indirect
	golang.org/x/tools v0.9.3 // indirect
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230525234030-28d5490b6b19 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	kernel.org/pub/linux/libs/security/libcap/psx v1.2.69 // indirect
	rsc.io/qr v0.2.0 // indirect
)
