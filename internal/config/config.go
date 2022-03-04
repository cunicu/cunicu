package config

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"go.uber.org/zap"
	"riasc.eu/wice/pkg/proxy"

	"github.com/pion/ice/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Copied from pion/ice/agent_config.go
const (
	// defaultCheckInterval is the interval at which the agent performs candidate checks in the connecting phase
	defaultCheckInterval = 200 * time.Millisecond

	// keepaliveInterval used to keep candidates alive
	defaultKeepaliveInterval = 2 * time.Second

	// defaultDisconnectedTimeout is the default time till an Agent transitions disconnected
	defaultDisconnectedTimeout = 5 * time.Second

	// defaultRestartInterval is the default time an Agent waits before it attempts an ICE restart
	defaultRestartTimeout = 5 * time.Second

	// defaultFailedTimeout is the default time till an Agent transitions to failed after disconnected
	defaultFailedTimeout = 25 * time.Second

	// max binding request before considering a pair failed
	defaultMaxBindingRequests = 7

	DefaultSocketPath = "/var/run/wice.sock"
)

var (
	defaultICEUrls = []*ice.URL{
		{
			Scheme: ice.SchemeTypeSTUN,
			Host:   "stun.l.google.com",
			Port:   19302,
			Proto:  ice.ProtoTypeUDP,
		},
	}

	defaultBackendURLs = []*url.URL{
		{
			Scheme: "p2p",
		},
	}
)

type Config struct {
	File string

	Community string

	Socket     string
	SocketWait bool

	Backends       backendURLList
	ProxyType      proxyType
	WatchInterval  time.Duration
	RestartTimeout time.Duration

	WireguardInterfaces      []string
	WireguardInterfaceFilter regex
	WireguardConfigSync      bool
	WireguardConfigPath      string
	WireguardUserspace       *bool

	// for ice.AgentConfig
	iceInterfaceFilter regex

	iceURLs iceURLList

	iceNat1to1IPs []net.IP

	iceInsecureSkipVerify bool

	iceCandidateTypes candidateTypeList
	iceNetworkTypes   networkTypeList

	iceDisconnectedTimeout time.Duration
	iceFailedTimeout       time.Duration
	iceKeepaliveInterval   time.Duration
	iceCheckInterval       time.Duration

	iceUsername string
	icePassword string

	icePortMin uint16
	icePortMax uint16

	iceMdns bool
	iceLite bool

	iceMaxBindingRequests uint16

	flags  *pflag.FlagSet
	viper  *viper.Viper
	logger *zap.Logger
}

func Parse(args ...string) (*Config, error) {
	f := pflag.NewFlagSet("", pflag.ContinueOnError)
	c := NewConfig(f)

	return c, c.flags.Parse(args)
}

func NewConfig(flags *pflag.FlagSet) *Config {
	matchAll, _ := regexp.Compile(".*")

	c := &Config{
		Backends:                 backendURLList{},
		iceCandidateTypes:        candidateTypeList{},
		iceInterfaceFilter:       regex{matchAll},
		iceNat1to1IPs:            []net.IP{},
		iceNetworkTypes:          networkTypeList{},
		iceURLs:                  iceURLList{},
		WireguardInterfaceFilter: regex{matchAll},
		WireguardInterfaces:      []string{},
		ProxyType:                proxyType{proxy.TypeAuto},

		flags:  flags,
		viper:  viper.New(),
		logger: zap.L().Named("config"),
	}

	flags.StringVarP(&c.Community, "community", "x", "", "Community passphrase for discovering other peers")
	flags.StringVarP(&c.File, "config", "c", "", "Path of configuration file")
	flags.VarP(&c.Backends, "backend", "b", "backend types / URLs")
	flags.VarP(&c.ProxyType, "proxy", "p", "proxy type to use")
	flags.VarP(&c.WireguardInterfaceFilter, "interface-filter", "f", "regex for filtering Wireguard interfaces (e.g. \"wg-.*\")")
	flags.DurationVarP(&c.WatchInterval, "watch-interval", "i", time.Second, "interval at which we are polling the kernel for updates on the Wireguard interfaces")

	c.WireguardUserspace = flags.BoolP("wg-userspace", "u", false, "start userspace Wireguard daemon")
	flags.BoolVarP(&c.WireguardConfigSync, "wg-config-sync", "S", false, "sync Wireguard interface with configuration file (see \"wg synconf\")")
	flags.StringVarP(&c.WireguardConfigPath, "wg-config-path", "w", "/etc/wireguard", "base path to search for Wireguard configuration files")

	// ice.AgentConfig fields
	flags.VarP(&c.iceURLs, "url", "a", "STUN and/or TURN server addresses")
	flags.Var(&c.iceCandidateTypes, "ice-candidate-type", "usable candidate types (select from \"host\", \"srflx\", \"prflx\", \"relay\")")
	flags.Var(&c.iceNetworkTypes, "ice-network-type", "usable network types (select from \"udp4\", \"udp6\", \"tcp4\", \"tcp6\")")
	flags.IPSliceVar(&c.iceNat1to1IPs, "ice-nat-1to1-ip", []net.IP{}, "IP addresses which will be added as local server reflexive candidates")

	flags.Uint16Var(&c.icePortMin, "ice-port-min", 0, "minimum port for allocation policy (range: 0-65535)")
	flags.Uint16Var(&c.icePortMax, "ice-port-max", 0, "maximum port for allocation policy (range: 0-65535)")
	flags.BoolVarP(&c.iceLite, "ice-lite", "L", false, "lite agents do not perform connectivity check and only provide host candidates")
	flags.BoolVarP(&c.iceMdns, "ice-mdns", "m", false, "enable local Multicast DNS discovery")
	flags.Uint16Var(&c.iceMaxBindingRequests, "ice-max-binding-requests", defaultMaxBindingRequests, "maximum number of binding request before considering a pair failed")
	flags.BoolVarP(&c.iceInsecureSkipVerify, "ice-insecure-skip-verify", "k", false, "skip verification of TLS certificates for secure STUN/TURN servers")
	flags.Var(&c.iceInterfaceFilter, "ice-interface-filter", "regex for filtering local interfaces for ICE candidate gathering (e.g. \"eth[0-9]+\")")
	flags.DurationVar(&c.iceDisconnectedTimeout, "ice-disconnected-timout", defaultDisconnectedTimeout, "time till an Agent transitions disconnected")
	flags.DurationVar(&c.iceFailedTimeout, "ice-failed-timeout", defaultFailedTimeout, "time until an Agent transitions to failed after disconnected")
	flags.DurationVar(&c.iceKeepaliveInterval, "ice-keepalive-interval", defaultKeepaliveInterval, "interval netween STUN keepalives")
	flags.DurationVar(&c.iceCheckInterval, "ice-check-interval", defaultCheckInterval, "interval at which the agent performs candidate checks in the connecting phase")
	flags.DurationVar(&c.RestartTimeout, "ice-restart-timeout", defaultRestartTimeout, "time to wait before ICE restart")
	flags.StringVarP(&c.iceUsername, "ice-user", "U", "", "username for STUN/TURN credentials")
	flags.StringVarP(&c.icePassword, "ice-pass", "P", "", "password for STUN/TURN credentials")

	flags.StringVarP(&c.Socket, "socket", "s", DefaultSocketPath, "Unix control and monitoring socket")
	flags.BoolVar(&c.SocketWait, "socket-wait", false, "wait until first client connected to control socket before continuing start")

	c.viper.BindPFlags(flags)

	return c
}

func (c *Config) Setup(args []string) {
	c.WireguardInterfaces = args

	// Find best proxy method
	if c.ProxyType.ProxyType == proxy.TypeAuto {
		c.ProxyType.ProxyType = proxy.AutoProxy()
	}

	// Add default backend
	if len(c.Backends) == 0 {
		c.Backends = defaultBackendURLs
	}

	if c.File != "" {
		// Use config file from the flag.
		viper.SetConfigFile(c.File)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			c.logger.Warn("Failed to determine home directory", zap.Error(err))
		} else {
			viper.AddConfigPath(filepath.Join(home, ".config", "wice"))
		}

		viper.AddConfigPath("/etc/wice")

		viper.SetConfigType("ini")
		viper.SetConfigName("wicerc")
	}

	c.viper.SetEnvPrefix("wice")
	c.viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		c.logger.Debug("Using config file", zap.String("file", viper.ConfigFileUsed()))
	}
}

func (c *Config) AgentConfig() (*ice.AgentConfig, error) {
	cfg := &ice.AgentConfig{
		InsecureSkipVerify: c.iceInsecureSkipVerify,
		NetworkTypes:       c.iceNetworkTypes,
		CandidateTypes:     c.iceCandidateTypes,
		Urls:               c.iceURLs,
		Lite:               c.iceLite,
		PortMin:            c.icePortMin,
		PortMax:            c.icePortMax,
	}

	cfg.InterfaceFilter = func(name string) bool {
		return c.iceInterfaceFilter.Match([]byte(name))
	}

	// Add default STUN/TURN servers
	if len(cfg.Urls) == 0 {
		cfg.Urls = defaultICEUrls
	} else {
		// Set ICE credentials
		for _, u := range cfg.Urls {
			if c.iceUsername != "" {
				u.Username = c.iceUsername
			}

			if c.icePassword != "" {
				u.Password = c.icePassword
			}
		}
	}

	if c.iceMaxBindingRequests > 0 {
		cfg.MaxBindingRequests = &c.iceMaxBindingRequests
	}

	if c.iceMdns {
		cfg.MulticastDNSMode = ice.MulticastDNSModeQueryAndGather
	}

	if c.iceDisconnectedTimeout > 0 {
		cfg.DisconnectedTimeout = &c.iceDisconnectedTimeout
	}

	if c.iceFailedTimeout > 0 {
		cfg.FailedTimeout = &c.iceFailedTimeout
	}

	if c.iceKeepaliveInterval > 0 {
		cfg.KeepaliveInterval = &c.iceKeepaliveInterval
	}

	if c.iceCheckInterval > 0 {
		cfg.CheckInterval = &c.iceCheckInterval
	}

	if len(c.iceNat1to1IPs) > 0 {
		cfg.NAT1To1IPCandidateType = ice.CandidateTypeServerReflexive

		cfg.NAT1To1IPs = []string{}
		for _, i := range c.iceNat1to1IPs {
			cfg.NAT1To1IPs = append(cfg.NAT1To1IPs, i.String())
		}
	}

	// Default network types
	if len(c.iceNetworkTypes) == 0 {
		cfg.NetworkTypes = append(cfg.NetworkTypes,
			ice.NetworkTypeUDP4,
			ice.NetworkTypeUDP6,
		)
	}

	return cfg, nil
}

func (c *Config) Dump(wr io.Writer) {
	cfg, _ := c.AgentConfig()

	fmt.Fprintln(wr, "Options:")
	fmt.Fprintf(wr, "  config file: %s\n", c.File)
	fmt.Fprintf(wr, "  community: %s\n", c.Community)
	fmt.Fprintf(wr, "  control socket: %s\n", c.Socket)
	fmt.Fprintf(wr, "  wait for control socket: %s\n", strconv.FormatBool(c.SocketWait))
	fmt.Fprintf(wr, "  userspace: %s\n", strconv.FormatBool(*c.WireguardUserspace))
	fmt.Fprintf(wr, "  config sync: %s\n", strconv.FormatBool(c.WireguardConfigSync))
	fmt.Fprintln(wr, "  urls:")
	for _, u := range cfg.Urls {
		fmt.Fprintf(wr, "    %s\n", u.String())
	}

	fmt.Fprintf(wr, "  interface filter: %s\n", c.WireguardInterfaceFilter)
	fmt.Fprintf(wr, "  interface filter ice: %s\n", c.iceInterfaceFilter)
	fmt.Fprintln(wr, "  interfaces:")
	for _, d := range c.WireguardInterfaces {
		fmt.Fprintf(wr, "    %s\n", d)
	}

	fmt.Fprintf(wr, "  restart timeout: %s\n", c.RestartTimeout)
	fmt.Fprintf(wr, "  watch interval: %s\n", c.WatchInterval)
	fmt.Fprintf(wr, "  proxy type: %s\n", c.ProxyType.String())

	fmt.Fprintf(wr, "  signaling backends:\n")
	for _, b := range c.Backends {
		fmt.Fprintf(wr, "    %s\n", b)
	}
}
