package main

import (
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bombsimon/logrusr"
	log "github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl"
	"k8s.io/klog/v2"

	"riasc.eu/wice/pkg/args"
	"riasc.eu/wice/pkg/intf"
	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/socket"

	_ "riasc.eu/wice/pkg/signaling/k8s"
	_ "riasc.eu/wice/pkg/signaling/p2p"
)

func setupLogging() {
	klogger := log.StandardLogger()
	klogr := logrusr.NewLogger(klogger)

	klog.SetLogger(klogr.WithName("k8s"))

	log.SetFormatter(&log.TextFormatter{
		// ForceColors:  true,
		// DisableQuote: true,
	})
}

func setupRand() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func setupSignals() chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	return ch
}

func main() {
	setupLogging()
	setupRand()
	signals := setupSignals()

	args, err := args.Parse(os.Args[0], os.Args[1:])
	if err != nil {
		log.WithError(err).Fatal("Failed to parse arguments")
	}
	if log.GetLevel() > log.DebugLevel {
		args.DumpConfig(os.Stdout)
	}

	// Create control socket server
	server, err := socket.Listen("unix", args.Socket, args.SocketWait)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize control socket")
	}

	// Create backend
	var backend signaling.Backend
	if len(args.Backends) == 1 {
		backend, err = signaling.NewBackend(args.Backends[0], server)
	} else {
		backend, err = signaling.NewMultiBackend(args.Backends, server)
	}
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize backend")
	}

	// Create Wireguard netlink socket
	client, err := wgctrl.New()
	if err != nil {
		log.Fatal(err)
	}

	// Create interfaces
	interfaces := &intf.Interfaces{}
	defer interfaces.CloseAll()

	interfaces.CreateFromArgs(client, backend, server, args)

	events := make(chan intf.InterfaceEvent, 16)
	errors := make(chan error, 16)

	if err := intf.WatchWireguardUserspaceInterfaces(events, errors); err != nil {
		log.WithError(err).Error("Failed to watch userspace interfaces")
		return
	}

	if err := intf.WatchWireguardKernelInterfaces(events, errors); err != nil {
		log.WithError(err).Error("Failed to watch kernel interfaces")
		return
	}

	log.Debug("Starting initial interface sync")
	interfaces.SyncAll(client, backend, server, args)

	ticker := time.NewTicker(args.WatchInterval)

out:
	for {
		select {
		// We still a need periodic sync we can not (yet) monitor Wireguard interfaces
		// for changes via a netlink socket (patch is pending)
		case <-ticker.C:
			log.Trace("Starting periodic interface sync")
			interfaces.SyncAll(client, backend, server, args)

			backend.Tick()

		case event := <-events:
			log.Trace("Received interface event: %s", event)
			interfaces.SyncAll(client, backend, server, args)

		case err := <-errors:
			log.WithError(err).Error("Failed to watch for interface changes")

		case sig := <-signals:
			log.WithField("signal", sig).Debug("Received signal")
			switch sig {
			case syscall.SIGUSR1:
				interfaces.SyncAll(client, backend, server, args)
			default:
				break out
			}
		}
	}
}
