package intf

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"riasc.eu/wice/pkg/args"
	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/socket"

	"golang.zx2c4.com/wireguard/wgctrl"
)

type Interfaces []Interface

func (interfaces *Interfaces) GetByName(name string) Interface {
	for _, intf := range *interfaces {
		if intf.Name() == name {
			return intf
		}
	}

	return nil
}

func (interfaces *Interfaces) CloseAll() {
	for _, intf := range *interfaces {
		intf.Close()
	}
}

func (interfaces *Interfaces) SyncAll(client *wgctrl.Client, backend signaling.Backend, server *socket.Server, args *args.Args) error {
	devices, err := client.Devices()
	if err != nil {
		log.WithError(err).Fatal("Failed to list Wireguard interfaces")
	}

	syncedInterfaces := Interfaces{}
	keepInterfaces := Interfaces{}

	for _, device := range devices {
		if !args.InterfaceRegex.Match([]byte(device.Name)) {
			continue // Skip interfaces which dont match the filter
		}

		// Find matching interface
		intf := interfaces.GetByName(device.Name)
		if intf == nil { // new interface
			log.WithField("intf", device.Name).Info("Adding new interface")

			i, err := NewInterface(device, client, backend, server, args)
			if err != nil {
				log.WithField("intf", device.Name).WithError(err).Fatalf("Failed to create new interface")
			}

			intf = &i

			*interfaces = append(*interfaces, &i)
		} else { // existing interface
			log.WithField("intf", intf.Name()).Trace("Sync existing interface")

			if err := intf.Sync(device); err != nil {
				log.WithError(err).WithField("intf", intf.Name()).Fatal("Failed to sync interface")
			}
		}

		syncedInterfaces = append(syncedInterfaces, intf)
	}

	for _, intf := range *interfaces {
		i := syncedInterfaces.GetByName(intf.Name())
		if i == nil {
			log.WithField("intf", intf.Name()).Info("Removing vanished interface")

			if err := intf.Close(); err != nil {
				log.WithError(err).Fatal("Failed to close interface")
			}

			server.BroadcastEvent(&socket.Event{
				Type:      "interface",
				State:     "removed",
				Interface: i.Name(),
			})
		} else {
			keepInterfaces = append(keepInterfaces, intf)
		}
	}

	*interfaces = keepInterfaces

	return nil
}

func (interfaces *Interfaces) CreateFromArgs(client *wgctrl.Client, backend signaling.Backend, server *socket.Server, args *args.Args) error {
	var devs Devices
	devs, err := client.Devices()
	if err != nil {
		return err
	}

	for _, i := range args.Interfaces {
		dev := devs.GetByName(i)
		if dev != nil {
			log.WithField("intf", i).Warn("Interface already exists. Skipping..")
			continue
		}

		var intf Interface
		if args.User {
			intf, err = CreateUserInterface(i, client, backend, server, args)
		} else {
			intf, err = CreateKernelInterface(i, client, backend, server, args)
		}
		if err != nil {
			return fmt.Errorf("failed to create Wireguard device: %w", err)
		}

		if log.GetLevel() >= log.DebugLevel {
			log.Debug("Intialized interface:")
			intf.DumpConfig(os.Stdout)
		}

		*interfaces = append(*interfaces, intf)
	}

	return nil
}
