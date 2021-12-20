package args

import (
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	pice "riasc.eu/wice/internal/ice"
	"riasc.eu/wice/pkg/backend"
	"riasc.eu/wice/pkg/proxy"

	"github.com/pion/ice/v2"
)

// Copied from pion/ice/agent_config.go
const (
	// defaultCheckInterval is the interval at which the agent performs candidate checks in the connecting phase
	defaultCheckInterval = 200 * time.Millisecond

	// keepaliveInterval used to keep candidates alive
	defaultKeepaliveInterval = 2 * time.Second

	// defaultDisconnectedTimeout is the default time till an Agent transitions disconnected
	defaultDisconnectedTimeout = 5 * time.Second

	// defaultFailedTimeout is the default time till an Agent transitions to failed after disconnected
	defaultFailedTimeout = 25 * time.Second

	// max binding request before considering a pair failed
	defaultMaxBindingRequests = 7
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ",")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type Args struct {
	Backend        *url.URL
	BackendOptions map[string]string
	User           bool
	ProxyType      proxy.ProxyType
	// Discover        bool
	ConfigSync      bool
	ConfigPath      string
	WatchInterval   time.Duration
	RestartInterval time.Duration

	InterfaceRegex    *regexp.Regexp
	IceInterfaceRegex *regexp.Regexp
	AgentConfig       ice.AgentConfig

	Socket string

	Interfaces []string
}

var yesno = map[bool]string{
	true:  "yes",
	false: "no",
}

func showUsage() {
	fmt.Fprintf(os.Stderr, "usage: %s [OPTIONS] [IFACES ...]\n", os.Args[0])
	fmt.Println()
	fmt.Println("  IFACES  is a list of Wireguard interfaces")
	fmt.Println("          (defaults to all available Wireguard interfaces)")
	fmt.Println("")
	fmt.Println(("Available OPTIONS are:"))
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("  (**) These options can be specified multiple times")
	fmt.Println()
	fmt.Println("Available backends types are:")
	for name, plugin := range backend.Backends {
		fmt.Printf("  %-7s %s\n", name, plugin.Description)
	}
}

func (a *Args) DumpConfig(wr io.Writer) {
	fmt.Fprintln(wr, "Options:")
	fmt.Fprintln(wr, "  URLs:")
	for _, u := range a.AgentConfig.Urls {
		fmt.Fprintf(wr, "    %s\n", u.String())
	}

	fmt.Fprintln(wr, "  Interfaces:")
	for _, d := range a.Interfaces {
		fmt.Fprintf(wr, "    %s\n", d)
	}

	fmt.Fprintf(wr, "  User: %s\n", yesno[a.User])
	fmt.Fprintf(wr, "  ProxyType: %s\n", a.ProxyType.String())

	fmt.Fprintf(wr, "  Backend: %s\n", a.Backend.String())
	fmt.Fprintln(wr, "  Backend options:")
	for k := range a.BackendOptions {
		fmt.Fprintf(wr, "    %s=%s\n", k, a.BackendOptions[k])
	}
}

func candidateTypeFromString(t string) (ice.CandidateType, error) {
	switch t {
	case "host":
		return ice.CandidateTypeHost, nil
	case "srflx":
		return ice.CandidateTypeServerReflexive, nil
	case "prflx":
		return ice.CandidateTypePeerReflexive, nil
	case "relay":
		return ice.CandidateTypeRelay, nil
	default:
		return ice.CandidateTypeUnspecified, fmt.Errorf("unknown candidate type: %s", t)
	}
}

func networkTypeFromString(t string) (ice.NetworkType, error) {
	switch t {
	case "udp4":
		return ice.NetworkTypeUDP4, nil
	case "udp6":
		return ice.NetworkTypeUDP6, nil
	case "tcp4":
		return ice.NetworkTypeTCP4, nil
	case "tcp6":
		return ice.NetworkTypeTCP6, nil
	default:
		return ice.NetworkTypeTCP4, fmt.Errorf("unknown network type: %s", t)
	}
}

func Parse(progname string, argv []string) (*Args, error) {
	var uri string
	var err error

	var iceURLs, iceCandidateTypes, iceNetworkTypes, iceNat1to1IPs arrayFlags

	flags := flag.NewFlagSet(progname, flag.ContinueOnError)

	flags.Usage = showUsage

	logLevel := flags.String("log-level", "info", "log level (one of \"panic\", \"fatal\", \"error\", \"warn\", \"info\", \"debug\", \"trace\")")
	// discover := flag.Bool("discover", false, "discover peers using the backend")
	backend := flags.String("backend", "http://localhost:8080", "backend URL")
	backendOpts := flags.String("backend-opts", "", "comma-separated list of additional backend options (e.g. \"key1=val1,key2-val2\")")
	user := flags.Bool("user", false, "start userspace Wireguard daemon")
	proxyType := flags.String("proxy", "auto", "proxy type to use")
	interfaceFilter := flags.String("interface-filter", ".*", "regex for filtering Wireguard interfaces (e.g. \"wg-.*\")")
	configSync := flags.Bool("config-sync", false, "sync Wireguard interface with configuration file (see \"wg synconf\"")
	configPath := flags.String("config-path", "/etc/wireguard", "base path to search for Wireguard configuration files")
	watchInterval := flags.Duration("watch-interval", 2*time.Second, "interval at which we are polling the kernel for updates on the Wireguard interfaces")

	// ice.AgentConfig fields
	flags.Var(&iceURLs, "url", "STUN and/or TURN server address  (**)")
	flags.Var(&iceCandidateTypes, "ice-candidate-type", "usable candidate types (**, one of \"host\", \"srflx\", \"prflx\", \"relay\")")
	flags.Var(&iceNetworkTypes, "ice-network-type", "usable network types (**, select from \"udp4\", \"udp6\", \"tcp4\", \"tcp6\")")
	flags.Var(&iceNat1to1IPs, "ice-nat-1to1-ips", "list of IP addresses which will be added as local server reflexive candidates (**)")

	icePortMin := flags.Uint("ice-port-min", 0, "minimum port for allocation policy (range: 0-65535)")
	icePortMax := flags.Uint("ice-port-max", 0, "maximum port for allocation policy (range: 0-65535)")
	iceLite := flags.Bool("ice-lite", false, "lite agents do not perform connectivity check and only provide host candidates")
	iceMdns := flags.Bool("ice-mdns", false, "enable local Multicast DNS discovery")
	iceMaxBindingRequests := flags.Int("ice-max-binding-requests", defaultMaxBindingRequests, "maximum number of binding request before considering a pair failed")
	iceInsecureSkipVerify := flags.Bool("ice-insecure-skip-verify", false, "skip verification of TLS certificates for secure STUN/TURN servers")
	iceInterfaceFilter := flags.String("ice-interface-filter", ".*", "regex for filtering local interfaces for ICE candidate gathering (e.g. \"eth[0-9]+\")")
	iceDisconnectedTimeout := flags.Duration("ice-disconnected-timout", defaultDisconnectedTimeout, "time till an Agent transitions disconnected")
	iceFailedTimeout := flags.Duration("ice-failed-timeout", defaultFailedTimeout, "time until an Agent transitions to failed after disconnected")
	iceKeepaliveInterval := flags.Duration("ice-keepalive-interval", defaultKeepaliveInterval, "interval netween STUN keepalives")
	iceCheckInterval := flags.Duration("ice-check-interval", defaultCheckInterval, "interval at which the agent performs candidate checks in the connecting phase")
	iceRestartInterval := flags.Duration("ice-restart-interval", defaultDisconnectedTimeout, "time to wait before ICE restart")
	iceUsername := flags.String("ice-user", "", "username for STUN/TURN credentials")
	icePassword := flags.String("ice-pass", "", "password for STUN/TURN credentials")
	// iceMaxBindingRequestTimeout := flag.Duration("ice-max-binding-request-timeout", maxBindingRequestTimeout, "wait time before binding requests can be deleted")

	socket := flags.String("socket", "/var/run/wice.sock", "Unix control and monitoring socket")

	flags.Parse(argv)

	args := &Args{
		User:           *user,
		ProxyType:      proxy.ProxyTypeFromString(*proxyType),
		BackendOptions: make(map[string]string),
		// Discover:        *discover,
		ConfigSync:      *configSync,
		ConfigPath:      *configPath,
		WatchInterval:   *watchInterval,
		RestartInterval: *iceRestartInterval,
		Socket:          *socket,
		Interfaces:      flag.Args(),
		AgentConfig: ice.AgentConfig{
			PortMin:            uint16(*icePortMin),
			PortMax:            uint16(*icePortMax),
			Lite:               *iceLite,
			InsecureSkipVerify: *iceInsecureSkipVerify,
		},
	}

	// Find best proxy method
	if args.ProxyType == proxy.ProxyTypeAuto {
		args.ProxyType = proxy.AutoProxy()
	}

	// Check proxy type
	if args.ProxyType == proxy.ProxyTypeInvalid {
		return nil, fmt.Errorf("invalid proxy type: %s", *proxyType)
	}

	// Compile interface regex
	args.InterfaceRegex, err = regexp.Compile(*interfaceFilter)
	if err != nil {
		return nil, fmt.Errorf("invalid interface filter: %w", err)
	}

	// Parse log level
	if lvl, err := log.ParseLevel(*logLevel); err != nil {
		return nil, fmt.Errorf("invalid log level: %s", *logLevel)
	} else {
		log.SetLevel(lvl)
	}

	// Parse backend URI
	if !strings.Contains(*backend, ":") {
		*backend += ":"
	}
	if args.Backend, err = url.Parse(*backend); err != nil {
		return nil, fmt.Errorf("invalid URI: %w", err)
	}

	// Parse additional backend options
	if *backendOpts != "" {
		opts := strings.Split(*backendOpts, ",")
		for _, opt := range opts {
			kv := strings.SplitN(opt, "=", 2)
			if len(kv) < 2 {
				return nil, fmt.Errorf("invalid backend option: %s", opt)
			}

			key := kv[0]
			value := kv[1]

			args.BackendOptions[key] = value
		}
	}

	if *iceMaxBindingRequests >= 0 {
		maxBindingReqs := uint16(*iceMaxBindingRequests)
		args.AgentConfig.MaxBindingRequests = &maxBindingReqs
	}
	if *iceMdns {
		args.AgentConfig.MulticastDNSMode = ice.MulticastDNSModeQueryAndGather
	}
	if *iceDisconnectedTimeout > 0 {
		args.AgentConfig.DisconnectedTimeout = iceDisconnectedTimeout
	}
	if *iceFailedTimeout > 0 {
		args.AgentConfig.FailedTimeout = iceFailedTimeout
	}
	if *iceKeepaliveInterval > 0 {
		args.AgentConfig.KeepaliveInterval = iceKeepaliveInterval
	}
	if *iceCheckInterval > 0 {
		args.AgentConfig.CheckInterval = iceCheckInterval
	}
	if len(iceNat1to1IPs) > 0 {
		args.AgentConfig.NAT1To1IPCandidateType = ice.CandidateTypeServerReflexive
		args.AgentConfig.NAT1To1IPs = iceNat1to1IPs
	}

	args.IceInterfaceRegex, err = regexp.Compile(*iceInterfaceFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to compile interface regex: %w", err)
	}

	// Parse candidate types
	for _, c := range iceCandidateTypes {
		candType, err := candidateTypeFromString(c)
		if err != nil {
			return nil, err
		}
		args.AgentConfig.CandidateTypes = append(args.AgentConfig.CandidateTypes, candType)
	}

	// Parse network types
	if len(iceNetworkTypes) == 0 {
		args.AgentConfig.NetworkTypes = []ice.NetworkType{
			ice.NetworkTypeUDP4,
			ice.NetworkTypeUDP6,
		}
	} else {
		for _, n := range iceNetworkTypes {
			netType, err := networkTypeFromString(n)
			if err != nil {
				return nil, err
			}
			args.AgentConfig.NetworkTypes = append(args.AgentConfig.NetworkTypes, netType)
		}
	}

	// Parse ICE urls
	for _, uri = range iceURLs {
		iceUrl, err := ice.ParseURL(uri)
		if err != nil {
			return nil, fmt.Errorf("failed to parse url %s: %w", uri, err)
		}

		if *iceUsername != "" {
			iceUrl.Username = *iceUsername
		}
		if *icePassword != "" {
			iceUrl.Password = *icePassword
		}

		args.AgentConfig.Urls = append(args.AgentConfig.Urls, iceUrl)
	}

	// Add default STUN server
	if len(args.AgentConfig.Urls) == 0 {
		url := &ice.URL{
			Scheme:   ice.SchemeTypeSTUN,
			Host:     "stun.l.google.com",
			Port:     19302,
			Username: "",
			Password: "",
			Proto:    ice.ProtoTypeUDP,
		}
		args.AgentConfig.Urls = append(args.AgentConfig.Urls, url)
	}

	args.AgentConfig.LoggerFactory = &pice.LoggerFactory{}

	return args, nil
}
