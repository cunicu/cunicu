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
	be "riasc.eu/wice/pkg/backend"
	"riasc.eu/wice/pkg/intf"

	_ "riasc.eu/wice/pkg/backend/http"
	_ "riasc.eu/wice/pkg/backend/k8s"
	_ "riasc.eu/wice/pkg/backend/p2p"
)

func setupLogging() {
	klogger := log.StandardLogger()
	klogr := logrusr.NewLogger(klogger)

	klog.SetLogger(klogr.WithName("k8s"))

	log.SetFormatter(&log.TextFormatter{
		ForceColors:  true,
		DisableQuote: true,
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

	// Create backend
	backend, err := be.NewBackend(args.Backend, args.BackendOptions)
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

	interfaces.CreateFromArgs(client, backend, args)

	events, errors, err := intf.WatchWireguardInterfaces()
	if err != nil {
		log.WithError(err).Error("Failed to watch interfaces")
		return
	}

	log.Debug("Starting initial interface sync")
	interfaces.SyncAll(client, backend, args)

	ticker := time.NewTicker(args.WatchInterval)

out:
	for {
		select {
		// We still a need periodic sync we can not (yet) monitor Wireguard interfaces
		// for changes via a netlink socket (patch is pending)
		case <-ticker.C:
			log.Trace("Starting periodic interface sync")
			interfaces.SyncAll(client, backend, args)

		case event := <-events:
			log.Trace("Received interface event: %s", event)
			interfaces.SyncAll(client, backend, args)

		case err := <-errors:
			log.WithError(err).Error("Failed to watch for interface changes")

		case sig := <-signals:
			log.WithField("signal", sig).Debug("Received signal")
			switch sig {
			case syscall.SIGUSR1:
				interfaces.SyncAll(client, backend, args)
			default:
				break out
			}
		}
	}
}
