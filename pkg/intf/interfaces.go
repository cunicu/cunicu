package intf

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
	"riasc.eu/wice/pkg/args"
	"riasc.eu/wice/pkg/pb"
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
	logger := zap.L().Named("interfaces")

	devices, err := client.Devices()
	if err != nil {
		logger.Fatal("Failed to list Wireguard interfaces", zap.Error(err))
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
			logger.Info("Adding new interface", zap.String("intf", device.Name))

			i, err := NewInterface(device, client, backend, server, args)
			if err != nil {
				logger.Fatal("Failed to create new interface",
					zap.Error(err),
					zap.String("intf", device.Name),
				)
			}

			intf = &i

			*interfaces = append(*interfaces, &i)
		} else { // existing interface
			logger.Debug("Sync existing interface", zap.String("intf", device.Name))

			if err := intf.Sync(device); err != nil {
				logger.Fatal("Failed to sync interface",
					zap.Error(err),
					zap.String("intf", device.Name),
				)
			}
		}

		syncedInterfaces = append(syncedInterfaces, intf)
	}

	for _, intf := range *interfaces {
		i := syncedInterfaces.GetByName(intf.Name())
		if i == nil {
			logger.Info("Removing vanished interface", zap.String("intf", intf.Name()))

			if err := intf.Close(); err != nil {
				logger.Fatal("Failed to close interface", zap.Error(err))
			}

			server.BroadcastEvent(&pb.Event{
				Type:  "interface",
				State: "removed",
				Event: &pb.Event_Intf{
					Intf: &pb.InterfaceEvent{
						Interface: &pb.Interface{
							Name: i.Name(),
						},
					},
				},
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

	logger := zap.L().Named("interfaces")

	for _, i := range args.Interfaces {
		dev := devs.GetByName(i)
		if dev != nil {
			logger.Warn("Interface already exists. Skipping..", zap.Any("intf", i))
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

		if logger.Core().Enabled(zap.DebugLevel) {
			logger.Debug("Intialized interface:")
			intf.DumpConfig(&zapio.Writer{Log: logger})
		}

		*interfaces = append(*interfaces, intf)
	}

	return nil
}
