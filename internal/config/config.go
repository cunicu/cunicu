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

	Socket     string
	SocketWait bool

	Backends        backendURLList
	User            bool
	ProxyType       proxyType
	ConfigSync      bool
	ConfigPath      string
	WatchInterval   time.Duration
	RestartInterval time.Duration

	Interfaces []string

	InterfaceFilter    regex
	InterfaceFilterICE regex

	// for ice.AgentConfig
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

	cfg := &Config{
		Backends:           backendURLList{},
		iceCandidateTypes:  candidateTypeList{},
		InterfaceFilterICE: regex{matchAll},
		iceNat1to1IPs:      []net.IP{},
		iceNetworkTypes:    networkTypeList{},
		iceURLs:            iceURLList{},
		InterfaceFilter:    regex{matchAll},
		Interfaces:         []string{},
		ProxyType:          proxyType{proxy.TypeAuto},

		flags:  flags,
		viper:  viper.New(),
		logger: zap.L().Named("config"),
	}

	flags.StringVarP(&cfg.File, "config", "c", "", "Path of configuration file")
	flags.VarP(&cfg.Backends, "backend", "b", "backend types / URLs")
	flags.VarP(&cfg.ProxyType, "proxy", "p", "proxy type to use")
	flags.VarP(&cfg.InterfaceFilter, "interface-filter", "f", "regex for filtering Wireguard interfaces (e.g. \"wg-.*\")")
	flags.DurationVarP(&cfg.WatchInterval, "watch-interval", "i", time.Second, "interval at which we are polling the kernel for updates on the Wireguard interfaces")

	flags.BoolVarP(&cfg.User, "wg-user", "u", false, "start userspace Wireguard daemon")
	flags.BoolVarP(&cfg.ConfigSync, "wg-config-sync", "S", false, "sync Wireguard interface with configuration file (see \"wg synconf\")")
	flags.StringVarP(&cfg.ConfigPath, "wg-config-path", "w", "/etc/wireguard", "base path to search for Wireguard configuration files")

	// ice.AgentConfig fields
	flags.VarP(&cfg.iceURLs, "url", "a", "STUN and/or TURN server addresses")
	flags.Var(&cfg.iceCandidateTypes, "ice-candidate-type", "usable candidate types (select from \"host\", \"srflx\", \"prflx\", \"relay\")")
	flags.Var(&cfg.iceNetworkTypes, "ice-network-type", "usable network types (select from \"udp4\", \"udp6\", \"tcp4\", \"tcp6\")")
	flags.IPSliceVar(&cfg.iceNat1to1IPs, "ice-nat-1to1-ip", []net.IP{}, "IP addresses which will be added as local server reflexive candidates")

	flags.Uint16Var(&cfg.icePortMin, "ice-port-min", 0, "minimum port for allocation policy (range: 0-65535)")
	flags.Uint16Var(&cfg.icePortMax, "ice-port-max", 0, "maximum port for allocation policy (range: 0-65535)")
	flags.BoolVarP(&cfg.iceLite, "ice-lite", "l", false, "lite agents do not perform connectivity check and only provide host candidates")
	flags.BoolVarP(&cfg.iceMdns, "ice-mdns", "m", false, "enable local Multicast DNS discovery")
	flags.Uint16Var(&cfg.iceMaxBindingRequests, "ice-max-binding-requests", defaultMaxBindingRequests, "maximum number of binding request before considering a pair failed")
	flags.BoolVarP(&cfg.iceInsecureSkipVerify, "ice-insecure-skip-verify", "k", false, "skip verification of TLS certificates for secure STUN/TURN servers")
	flags.Var(&cfg.InterfaceFilterICE, "ice-interface-filter", "regex for filtering local interfaces for ICE candidate gathering (e.g. \"eth[0-9]+\")")
	flags.DurationVar(&cfg.iceDisconnectedTimeout, "ice-disconnected-timout", defaultDisconnectedTimeout, "time till an Agent transitions disconnected")
	flags.DurationVar(&cfg.iceFailedTimeout, "ice-failed-timeout", defaultFailedTimeout, "time until an Agent transitions to failed after disconnected")
	flags.DurationVar(&cfg.iceKeepaliveInterval, "ice-keepalive-interval", defaultKeepaliveInterval, "interval netween STUN keepalives")
	flags.DurationVar(&cfg.iceCheckInterval, "ice-check-interval", defaultCheckInterval, "interval at which the agent performs candidate checks in the connecting phase")
	flags.DurationVar(&cfg.RestartInterval, "ice-restart-interval", defaultDisconnectedTimeout, "time to wait before ICE restart")
	flags.StringVarP(&cfg.iceUsername, "ice-user", "U", "", "username for STUN/TURN credentials")
	flags.StringVarP(&cfg.icePassword, "ice-pass", "P", "", "password for STUN/TURN credentials")

	flags.StringVarP(&cfg.Socket, "socket", "s", DefaultSocketPath, "Unix control and monitoring socket")
	flags.BoolVar(&cfg.SocketWait, "socket-wait", false, "wait until first client connected to control socket before continuing start")

	cfg.viper.BindPFlags(flags)

	return cfg
}

func (c *Config) Setup(args []string) {
	c.Interfaces = args

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

func (a *Config) AgentConfig() (*ice.AgentConfig, error) {
	cfg := &ice.AgentConfig{
		InsecureSkipVerify: a.iceInsecureSkipVerify,
		NetworkTypes:       a.iceNetworkTypes,
		CandidateTypes:     a.iceCandidateTypes,
		Urls:               a.iceURLs,
	}

	// Add default STUN/TURN servers
	if len(cfg.Urls) == 0 {
		cfg.Urls = defaultICEUrls
	} else {
		// Set ICE credentials
		for _, u := range cfg.Urls {
			if a.iceUsername != "" {
				u.Username = a.iceUsername
			}

			if a.icePassword != "" {
				u.Password = a.icePassword
			}
		}
	}

	if a.iceMaxBindingRequests > 0 {
		cfg.MaxBindingRequests = &a.iceMaxBindingRequests
	}

	if a.iceMdns {
		cfg.MulticastDNSMode = ice.MulticastDNSModeQueryAndGather
	}

	if a.iceDisconnectedTimeout > 0 {
		cfg.DisconnectedTimeout = &a.iceDisconnectedTimeout
	}

	if a.iceFailedTimeout > 0 {
		cfg.FailedTimeout = &a.iceFailedTimeout
	}

	if a.iceKeepaliveInterval > 0 {
		cfg.KeepaliveInterval = &a.iceKeepaliveInterval
	}

	if a.iceCheckInterval > 0 {
		cfg.CheckInterval = &a.iceCheckInterval
	}

	if len(a.iceNat1to1IPs) > 0 {
		cfg.NAT1To1IPCandidateType = ice.CandidateTypeServerReflexive

		cfg.NAT1To1IPs = []string{}
		for _, i := range a.iceNat1to1IPs {
			cfg.NAT1To1IPs = append(cfg.NAT1To1IPs, i.String())
		}
	}

	// Default network types
	if len(a.iceNetworkTypes) == 0 {
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
	fmt.Fprintln(wr, "  URLs:")
	for _, u := range cfg.Urls {
		fmt.Fprintf(wr, "    %s\n", u.String())
	}

	fmt.Fprintln(wr, "  Interfaces:")
	for _, d := range c.Interfaces {
		fmt.Fprintf(wr, "    %s\n", d)
	}

	fmt.Fprintf(wr, "  User: %s\n", strconv.FormatBool(c.User))
	fmt.Fprintf(wr, "  ProxyType: %s\n", c.ProxyType.String())

	fmt.Fprintf(wr, "  Signaling Backends:\n")
	for _, b := range c.Backends {
		fmt.Fprintf(wr, "    %s\n", b)
	}
}
