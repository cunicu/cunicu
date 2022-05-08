package config

import (
	"errors"
	"fmt"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Settings
	*viper.Viper

	ConfigFiles []string
	Domain      string

	flags  *pflag.FlagSet
	logger *zap.Logger
}

func init() {
	mime.AddExtensionType(".yaml", "text/yaml")
	mime.AddExtensionType(".yaml", "text/x-yaml")
	mime.AddExtensionType(".toml", "application/toml")
	mime.AddExtensionType(".env", "text/x-ini")
	mime.AddExtensionType(".ini", "text/x-ini")
	mime.AddExtensionType(".props", "text/x-java-properties")
}

func ParseArgs(args ...string) (*Config, error) {
	c := NewConfig(nil)

	if err := c.flags.Parse(args); err != nil {
		return nil, err
	}

	if err := c.Setup(c.flags.Args()); err != nil {
		return nil, err
	}

	return c, nil
}

func NewConfig(flags *pflag.FlagSet) *Config {
	if flags == nil {
		flags = pflag.NewFlagSet("", pflag.ContinueOnError)
	}

	c := &Config{
		Viper: viper.New(),
		flags: flags,

		ConfigFiles: []string{},
	}

	// Defaults
	c.SetDefault("backends", []string{DefaultBackend})
	c.SetDefault("watch_interval", "1s")
	c.SetDefault("socket.path", DefaultSocketPath)
	c.SetDefault("socket.wait", false)
	c.SetDefault("wg.config.path", "/etc/wireguard")
	c.SetDefault("proxy.ebpf", true)
	c.SetDefault("proxy.nft", true)
	c.SetDefault("ice.check_interval", "200ms")
	c.SetDefault("ice.keepalive_interval", "2s")
	c.SetDefault("ice.disconnected_timeout", "5s")
	c.SetDefault("ice.restart_timeout", "5s")
	c.SetDefault("ice.failed_timeout", "5s")
	c.SetDefault("ice.max_binding_requests", 7)
	c.SetDefault("ice.urls", []string{DefaultURL})
	c.SetDefault("ice.port.min", 1<<15+1<<14) // Ephemeral port range for private and dynamic ports (see RFC6335)
	c.SetDefault("ice.port.max", 1<<16-1)

	flags.StringVarP(&c.Domain, "domain", "A", "", "A DNS `domain` name used for DNS auto-configuration")
	flags.StringSliceVarP(&c.ConfigFiles, "config", "c", []string{}, "A `filename`s of a file")

	flags.StringP("community", "x", "", "A community `passphrase` for discovering other peers")
	flags.StringSliceP("backend", "b", []string{}, "A signaling backend `URL`")
	flags.DurationP("watch-interval", "i", 0, "An interval at which we are periodically polling the kernel for updates on Wireguard interfaces")
	flags.BoolP("proxy-ebpf", "p", true, "Use eBPF filters to diverge STUN packets to wice")
	flags.BoolP("proxy-nft", "n", true, "Use NFTables to redirect STUN packets to wice")

	flags.StringP("wg-interface-filter", "f", ".*", "A `regex` for filtering Wireguard interfaces (e.g. \"wg-.*\")")
	flags.BoolP("wg-userspace", "u", false, "Start userspace Wireguard daemon")
	flags.BoolP("wg-config-sync", "S", false, "Synchronize Wireguard interface with configuration file (see \"wg synconf\")")
	flags.StringP("wg-config-path", "w", "", "The `directory` of Wireguard wg/wg-quick configuration files")

	// ice.AgentConfig fields
	flags.StringSliceP("url", "a", []string{}, "A STUN and/or TURN server `URL`")
	flags.StringP("username", "U", "", "The `username` for STUN/TURN credentials")
	flags.StringP("password", "P", "", "The `password` for STUN/TURN credentials")

	flags.StringSlice("ice-candidate-type", []string{}, "Usable `candidate-type`s (select from \"host\", \"srflx\", \"prflx\", \"relay\")")
	flags.StringSlice("ice-network-type", []string{}, "Usable `network-type`s (select from \"udp4\", \"udp6\", \"tcp4\", \"tcp6\")")
	flags.StringSlice("ice-nat-1to1-ip", []string{}, "An `IP` address which will be added as local server reflexive candidates")

	flags.Uint16("ice-port-min", 0, "Minimum `port` for allocation policy (range: 0-65535)")
	flags.Uint16("ice-port-max", 0, "Maximum `port` for allocation policy (range: 0-65535)")
	flags.BoolP("ice-lite", "L", false, "Lite agents do not perform connectivity check and only provide host candidates")
	flags.BoolP("ice-mdns", "m", false, "Enable local Multicast DNS discovery")
	flags.Uint16("ice-max-binding-requests", 0, "Maximum `number` of binding request before considering a pair failed")
	flags.BoolP("ice-insecure-skip-verify", "k", false, "Skip verification of TLS certificates for secure STUN/TURN servers")
	flags.String("ice-interface-filter", ".*", "A `regex` for filtering local interfaces for ICE candidate gathering (e.g. \"eth[0-9]+\")")
	flags.Duration("ice-disconnected-timeout", 0, "Time until an Agent transitions disconnected")
	flags.Duration("ice-failed-timeout", 0, "Time until an Agent transitions to failed after disconnected")
	flags.Duration("ice-keepalive-interval", 0, "Interval netween STUN keepalives")
	flags.Duration("ice-check-interval", 0, "Interval at which the agent performs candidate checks in the connecting phase")
	flags.Duration("ice-restart-timeout", 0, "Time to wait before ICE restart")

	flags.StringP("socket", "s", "", "The `path` of the unix socket used by other wice commands")
	flags.Bool("socket-wait", false, "Wait until first client connected to control socket before continuing start")

	flagMap := map[string]string{
		"community":                "community",
		"backend":                  "backends",
		"watch-interval":           "watch_interval",
		"wg-userspace":             "wg.userspace",
		"wg-interface-filter":      "wg.interface_filter",
		"wg-config-sync":           "wg.config.sync",
		"wg-config-path":           "wg.config.path",
		"url":                      "ice.urls",
		"username":                 "ice.username",
		"password":                 "ice.password",
		"ice-candidate-type":       "ice.candidate_types",
		"ice-network-type":         "ice.network_types",
		"ice-nat-1to1-ip":          "ice.nat_1to1_ips",
		"ice-port-min":             "ice.port.min",
		"ice-port-max":             "ice.port.max",
		"ice-lite":                 "ice.lite",
		"ice-mdns":                 "ice.mdns",
		"ice-max-binding-requests": "ice.max_binding_requests",
		"ice-insecure-skip-verify": "ice.insecure_skip_verify",
		"ice-interface-filter":     "ice.interface_filter",
		"ice-disconnected-timeout": "ice.disconnected_timeout",
		"ice-failed-timeout":       "ice.failed_timeout",
		"ice-keepalive-interval":   "ice.keepalive_interval",
		"ice-check-interval":       "ice.check_interval",
		"ice-restart-timeout":      "ice.restart_timeout",
		"socket":                   "socket.path",
		"socket-wait":              "socket.wait",
		"proxy-ebpf":               "proxy.ebpf",
		"proxy-nft":                "proxy.nft",
	}

	showAdvancedFlags := os.Getenv("WICE_ADVANCED_CLI") != ""
	advancedFlags := map[string]bool{
		"watch-interval":           true,
		"wg-config-sync":           true,
		"wg-config-path":           true,
		"ice-candidate-type":       true,
		"ice-network-type":         true,
		"ice-nat-1to1-ip":          true,
		"ice-port-min":             true,
		"ice-port-max":             true,
		"ice-lite":                 true,
		"ice-mdns":                 true,
		"ice-max-binding-requests": true,
		"ice-insecure-skip-verify": true,
		"ice-interface-filter":     true,
		"ice-disconnected-timeout": true,
		"ice-failed-timeout":       true,
		"ice-keepalive-interval":   true,
		"ice-check-interval":       true,
		"ice-restart-timeout":      true,
		"socket-wait":              true,
		"proxy-ebpf":               true,
		"proxy-nft":                true,
	}

	flags.VisitAll(func(flag *pflag.Flag) {
		if newName, ok := flagMap[flag.Name]; ok {
			c.BindPFlag(newName, flag)
		}

		if hide, ok := advancedFlags[flag.Name]; ok && hide && !showAdvancedFlags {
			flags.MarkHidden(flag.Name)
		}
	})

	return c
}

func (c *Config) Setup(args []string) error {
	// We cant to this in NewConfig since its called by init()
	// at which time the logging system is not initialized yet.
	c.logger = zap.L().Named("config")

	// First lookup settings via DNS
	if c.Domain != "" {
		if err := c.Lookup(c.Domain); err != nil {
			return fmt.Errorf("DNS autoconfiguration failed: %w", err)
		}
	}

	if len(c.ConfigFiles) > 0 {
		// Merge config files from the flags.
		for _, file := range c.ConfigFiles {
			if u, err := url.Parse(file); err == nil && u.Scheme != "" {
				if err := c.MergeRemoteConfig(u); err != nil {
					return fmt.Errorf("failed to load remote config: %w", err)
				}

				c.logger.Debug("Using remote configuration file", zap.Any("url", u))
			} else {
				c.SetConfigFile(file)
				if err := c.MergeInConfig(); err != nil {
					return fmt.Errorf("failed to merge configurations: %w", err)
				}

				c.logger.Debug("Using configuration file", zap.String("file", c.ConfigFileUsed()))
			}
		}
	} else {
		c.AddConfigPath("/etc")
		c.AddConfigPath(filepath.Join("$HOME", ".config"))
		c.AddConfigPath(".")
		c.SetConfigName("wice")

		if err := c.MergeInConfig(); err == nil {
			c.logger.Debug("Using configuration file", zap.String("file", c.ConfigFileUsed()))
		} else if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to merge configurations: %w", err)
		}
	}

	c.SetEnvPrefix("wice")
	c.AutomaticEnv()
	c.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := c.Load(); err != nil {
		return err
	}

	// We append the interfaces here because Config.Load() will overwrite them otherwise
	c.Wireguard.Interfaces = append(c.Wireguard.Interfaces, args...)

	return nil
}

func (c *Config) MergeRemoteConfig(url *url.URL) error {
	if url.Scheme != "https" {
		host, _, _ := net.SplitHostPort(url.Host)
		ip, err := net.ResolveIPAddr("ip", host)
		if err != nil || !ip.IP.IsLoopback() {
			return errors.New("remote configuration must by provided via HTTPS")
		}
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req := &http.Request{
		Method: "GET",
		URL:    url,
		Header: http.Header{},
	}
	// TODO: Add version info
	req.Header.Set("User-Agent", "wice")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %w", url, err)
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch: %s: %s", url, resp.Status)
	}

	contentType := resp.Header.Get("Content-type")
	fileExtension := filepath.Ext(url.Path)

	if contentType != "" {
		if types, err := mime.ExtensionsByType(contentType); err == nil && types != nil && len(types) > 0 {
			fileExtension = types[0][1:] // strip leading dot
		}
	}

	if fileExtension == "" {
		return fmt.Errorf("failed to load remote configuration file: failed to determine file-type by mime-type or filename suffix")
	}

	c.SetConfigType(fileExtension)

	return c.MergeConfig(resp.Body)
}

func decodeOption(cfg *mapstructure.DecoderConfig) {
	cfg.DecodeHook = mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		mapstructure.TextUnmarshallerHookFunc(),
	)

	cfg.ZeroFields = false
	cfg.TagName = "yaml"
}

func (c *Config) Load() error {
	if err := c.UnmarshalExact(&c.Settings, decodeOption); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return nil
}
