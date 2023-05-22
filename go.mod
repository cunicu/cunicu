module github.com/stv0g/cunicu

go 1.20

require (
	github.com/cenkalti/backoff/v4 v4.2.1
	github.com/cilium/ebpf v0.10.0
	github.com/dchest/siphash v1.2.3
	github.com/fsnotify/fsnotify v1.6.0
	github.com/go-logr/logr v1.2.4
	github.com/go-logr/zapr v1.2.4
	github.com/google/nftables v0.1.0
	github.com/imdario/mergo v0.3.15
	github.com/knadh/koanf v1.5.0
	github.com/mdp/qrterminal/v3 v3.0.0
	github.com/miekg/dns v1.1.54
	github.com/mitchellh/mapstructure v1.5.0
	github.com/pion/ice/v2 v2.3.5
	github.com/pion/randutil v0.1.0
	github.com/pion/stun v0.6.0
	github.com/pion/zapion v0.0.0-20220425013423-f29ac63a7069
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/vishvananda/netlink v1.2.1-beta.2.0.20221214185949-378a404a26f0
	go.uber.org/zap v1.24.0
	golang.org/x/crypto v0.9.0
	golang.org/x/exp v0.0.0-20230519143937-03e91628a987
	golang.org/x/sync v0.2.0
	golang.org/x/sys v0.8.0
	golang.zx2c4.com/wireguard v0.0.0-20230223181233-21636207a675
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20230215201556-9c5414ab4bde
	google.golang.org/grpc v1.55.0
	google.golang.org/protobuf v1.30.0
	gopkg.in/ini.v1 v1.67.0
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/apimachinery v0.27.2
	k8s.io/client-go v0.27.2
	k8s.io/klog/v2 v2.100.1
	kernel.org/pub/linux/libs/security/libcap/cap v1.2.69
)

require (
	github.com/foxcpp/go-mockdns v1.0.0 // test-only
	github.com/gopacket/gopacket v1.1.0 // test-only
	github.com/onsi/ginkgo/v2 v2.9.5 // test-only
	github.com/onsi/gomega v1.27.7 // test-only
	github.com/stv0g/gont/v2 v2.3.0 // test-only
	sigs.k8s.io/controller-runtime v0.14.6 // test-only
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.10.2 // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/fxamacker/cbor/v2 v2.4.0 // indirect
	github.com/go-delve/delve v1.20.1 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/go-ping/ping v1.1.0 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/gnostic v0.6.9 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/go-dap v0.8.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20230510103437-eeec1cb781c3 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/josharian/native v1.1.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mdlayher/genetlink v1.3.2 // indirect
	github.com/mdlayher/netlink v1.7.2 // indirect
	github.com/mdlayher/socket v0.4.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pion/dtls/v2 v2.2.7 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/mdns v0.0.7 // indirect
	github.com/pion/transport/v2 v2.2.1 // indirect
	github.com/pion/turn/v2 v2.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.9.2 // indirect
	github.com/stretchr/testify v1.8.3 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/goleak v1.2.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/arch v0.3.0 // indirect
	golang.org/x/mod v0.10.0 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/oauth2 v0.8.0 // indirect
	golang.org/x/term v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.9.1 // indirect
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.27.2 // indirect
	k8s.io/apiextensions-apiserver v0.27.2 // indirect
	k8s.io/kube-openapi v0.0.0-20230501164219-8b0f38b5fd1f // indirect
	k8s.io/utils v0.0.0-20230505201702-9f6742963106 // indirect
	kernel.org/pub/linux/libs/security/libcap/psx v1.2.69 // indirect
	rsc.io/qr v0.2.0 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)
