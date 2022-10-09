module github.com/stv0g/cunicu

go 1.19

require (
	github.com/cilium/ebpf v0.9.3
	github.com/dchest/siphash v1.2.3
	github.com/fsnotify/fsnotify v1.5.4
	github.com/go-logr/logr v1.2.3
	github.com/go-logr/zapr v1.2.3
	github.com/google/nftables v0.0.0-20221002140148-535f5eb8da79
	github.com/imdario/mergo v0.3.13
	github.com/jpillora/backoff v1.0.0
	github.com/knadh/koanf v1.4.3
	github.com/mdp/qrterminal/v3 v3.0.0
	github.com/miekg/dns v1.1.50
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pion/ice/v2 v2.2.10
	github.com/pion/logging v0.2.2
	github.com/pion/randutil v0.1.0
	github.com/pion/stun v0.3.5
	github.com/spf13/cobra v1.5.0
	github.com/spf13/pflag v1.0.5
	github.com/vishvananda/netlink v1.2.1-beta.2
	go.uber.org/atomic v1.10.0
	go.uber.org/zap v1.23.0
	golang.org/x/crypto v0.0.0-20221005025214-4161e89ecf1b
	golang.org/x/exp v0.0.0-20221006183845-316c7553db56
	golang.org/x/sync v0.0.0-20220929204114-8fcdb60fdcc0
	golang.org/x/sys v0.0.0-20221006211917-84dc82d7e875
	golang.zx2c4.com/wireguard v0.0.0-20220920152132-bb719d3a6e2c
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20220916014741-473347a5e6e3
	google.golang.org/grpc v1.50.0
	google.golang.org/protobuf v1.28.1
	gopkg.in/ini.v1 v1.67.0
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/apimachinery v0.25.2
	k8s.io/client-go v0.25.2
	k8s.io/klog/v2 v2.80.1
	kernel.org/pub/linux/libs/security/libcap/cap v1.2.66
)

require (
	github.com/foxcpp/go-mockdns v1.0.0 // test-only
	github.com/gopacket/gopacket v0.0.0-20221006103438-9e6d99b9b443 // test-only
	github.com/onsi/ginkgo/v2 v2.2.0 // test-only
	github.com/onsi/gomega v1.21.1 // test-only
	github.com/stv0g/gont v1.6.3 // test-only
	sigs.k8s.io/controller-runtime v0.13.0 // test-only
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.9.0 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/frankban/quicktest v1.14.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/go-ping/ping v1.1.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/gnostic v0.6.9 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/josharian/native v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mdlayher/genetlink v1.2.0 // indirect
	github.com/mdlayher/netlink v1.6.2 // indirect
	github.com/mdlayher/socket v0.2.3 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pion/dtls/v2 v2.1.5 // indirect
	github.com/pion/mdns v0.0.5 // indirect
	github.com/pion/transport v0.13.1 // indirect
	github.com/pion/turn/v2 v2.0.8 // indirect
	github.com/pion/udp v0.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.13.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/vishvananda/netns v0.0.0-20220913150850-18c4f4234207 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/net v0.0.0-20221004154528-8021a29435af // indirect
	golang.org/x/oauth2 v0.0.0-20221006150949-b44042a4b9c1 // indirect
	golang.org/x/term v0.0.0-20220919170432-7a66f970e087 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20220922220347-f3bd1da661af // indirect
	golang.org/x/tools v0.1.12 // indirect
	golang.zx2c4.com/wintun v0.0.0-20211104114900-415007cec224 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220930163606-c98284e70a91 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.25.2 // indirect
	k8s.io/apiextensions-apiserver v0.25.2 // indirect
	k8s.io/kube-openapi v0.0.0-20220928191237-829ce0c27909 // indirect
	k8s.io/utils v0.0.0-20220922133306-665eaaec4324 // indirect
	kernel.org/pub/linux/libs/security/libcap/psx v1.2.66 // indirect
	rsc.io/qr v0.2.0 // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

// Workaround until https://github.com/pion/ice/pull/483 is merged
replace github.com/pion/ice/v2 => github.com/pion/ice/v2 v2.2.11-0.20221009084925-46432b4dc499
