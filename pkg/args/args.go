package args

import (
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	pice "riasc.eu/wice/internal/ice"
	"riasc.eu/wice/pkg/proxy"
	"riasc.eu/wice/pkg/signaling"

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

var (
	defaultICEUrls = []*ice.URL{
		{
			Scheme:   ice.SchemeTypeSTUN,
			Host:     "stun.l.google.com",
			Port:     19302,
			Username: "",
			Password: "",
			Proto:    ice.ProtoTypeUDP,
		},
	}
)

type logLevel struct {
	log.Level
}

func (l *logLevel) Set(value string) error {
	if m, err := log.ParseLevel(value); err != nil {
		return err
	} else {
		l.Level = m
		return nil
	}
}

type backendURLList []*url.URL

func (i *backendURLList) String() string {
	s := []string{}
	for _, u := range *i {
		s = append(s, u.String())
	}

	return strings.Join(s, ",")
}

func (i *backendURLList) Set(value string) error {

	// Allow the user to specify just the backend type as a valid url.
	// E.g. "p2p" instead of "p2p:"
	if !strings.Contains(value, ":") {
		value += ":"
	}

	uri, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("invalid backend URI: %w", err)
	}

	*i = append(*i, uri)

	return nil
}

type iceURLList []*ice.URL

func (i *iceURLList) String() string {
	s := []string{}
	for _, u := range *i {
		s = append(s, u.String())
	}

	return strings.Join(s, ",")
}

func (i *iceURLList) Set(value string) error {
	iceUrl, err := ice.ParseURL(value)
	if err != nil {
		return fmt.Errorf("failed to parse ICE url %s: %w", value, err)
	}

	*i = append(*i, iceUrl)

	return nil
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ",")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type Args struct {
	Backends        []*url.URL
	User            bool
	ProxyType       proxy.Type
	ConfigSync      bool
	ConfigPath      string
	WatchInterval   time.Duration
	RestartInterval time.Duration

	InterfaceRegex    *regexp.Regexp
	IceInterfaceRegex *regexp.Regexp
	AgentConfig       ice.AgentConfig

	Socket     string
	SocketWait bool

	Interfaces []string
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
	for name, plugin := range signaling.Backends {
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

	fmt.Fprintf(wr, "  User: %s\n", strconv.FormatBool(a.User))
	fmt.Fprintf(wr, "  ProxyType: %s\n", a.ProxyType.String())

	fmt.Fprintf(wr, "  Signalling Backends:\n")
	for _, b := range a.Backends {
		fmt.Fprintf(wr, "    %s\n", b)
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
	var err error
	var iceURLs, iceCandidateTypes, iceNetworkTypes, iceNat1to1IPs arrayFlags
	var backendURLs backendURLList
	var logLevel logLevel = logLevel{log.InfoLevel}

	flags := flag.NewFlagSet(progname, flag.ContinueOnError)

	flags.Usage = showUsage

	flags.Var(&logLevel, "log-level", "log level (one of \"panic\", \"fatal\", \"error\", \"warn\", \"info\", \"debug\", \"trace\")")
	flags.Var(&backendURLs, "backend", "backend type / URL")
	user := flags.Bool("user", false, "start userspace Wireguard daemon")
	proxyType := flags.String("proxy", "auto", "proxy type to use")
	interfaceFilter := flags.String("interface-filter", ".*", "regex for filtering Wireguard interfaces (e.g. \"wg-.*\")")
	configSync := flags.Bool("config-sync", false, "sync Wireguard interface with configuration file (see \"wg synconf\"")
	configPath := flags.String("config-path", "/etc/wireguard", "base path to search for Wireguard configuration files")
	watchInterval := flags.Duration("watch-interval", time.Second, "interval at which we are polling the kernel for updates on the Wireguard interfaces")

	// ice.AgentConfig fields
	flags.Var(&iceURLs, "url", "STUN and/or TURN server address  (**)")
	flags.Var(&iceCandidateTypes, "ice-candidate-type", "usable candidate types (**, one of \"host\", \"srflx\", \"prflx\", \"relay\")")
	flags.Var(&iceNetworkTypes, "ice-network-type", "usable network types (**, select from \"udp4\", \"udp6\", \"tcp4\", \"tcp6\")")
	flags.Var(&iceNat1to1IPs, "ice-nat-1to1-ip", "list of IP addresses which will be added as local server reflexive candidates (**)")

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
	socketWait := flags.Bool("socket-wait", false, "wait until first client connected to control socket before continuing start")

	if err := flags.Parse(argv); err != nil {
		return nil, fmt.Errorf("failed to parse args: %w", err)
	}

	log.WithField("level", logLevel.Level).Info("Setting debug level")
	log.SetLevel(logLevel.Level)

	args := &Args{
		User:      *user,
		Backends:  backendURLs,
		ProxyType: proxy.ProxyTypeFromString(*proxyType),
		// Discover:        *discover,
		ConfigSync:      *configSync,
		ConfigPath:      *configPath,
		WatchInterval:   *watchInterval,
		RestartInterval: *iceRestartInterval,
		Socket:          *socket,
		SocketWait:      *socketWait,
		Interfaces:      flag.Args(),
		AgentConfig: ice.AgentConfig{
			PortMin:            uint16(*icePortMin),
			PortMax:            uint16(*icePortMax),
			Lite:               *iceLite,
			InsecureSkipVerify: *iceInsecureSkipVerify,
			Urls:               []*ice.URL{},
		},
	}

	// Find best proxy method
	if args.ProxyType == proxy.TypeAuto {
		args.ProxyType = proxy.AutoProxy()
	}

	// Check proxy type
	if args.ProxyType == proxy.TypeInvalid {
		return nil, fmt.Errorf("invalid proxy type: %s", *proxyType)
	}

	// Compile interface regex
	args.InterfaceRegex, err = regexp.Compile(*interfaceFilter)
	if err != nil {
		return nil, fmt.Errorf("invalid interface filter: %w", err)
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

	// Add default backend
	if len(args.Backends) == 0 {
		args.Backends = append(args.Backends, &url.URL{
			Scheme: "p2p",
		})
	}

	// Add default STUN/TURN servers
	if len(args.AgentConfig.Urls) == 0 {
		args.AgentConfig.Urls = defaultICEUrls
	} else {
		// Set ICE credentials
		for _, u := range args.AgentConfig.Urls {
			if *iceUsername != "" {
				u.Username = *iceUsername
			}

			if *icePassword != "" {
				u.Password = *icePassword
			}
		}
	}

	args.AgentConfig.LoggerFactory = &pice.LoggerFactory{}

	return args, nil
}
