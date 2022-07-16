//go:build !windows

/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2017-2021 WireGuard LLC. All Rights Reserved.
 */

package core

import (
	"fmt"
	"net"

	"go.uber.org/zap"
	"riasc.eu/wice/internal/config"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/ipc"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/wgctrl"
)

type UserDevice struct {
	BaseInterface

	tun        tun.Device
	log        *device.Logger
	userDevice *device.Device
	userAPI    net.Listener
}

func newLogger(log *zap.Logger) *device.Logger {
	logger := log.Named("wireguard").Sugar()

	return &device.Logger{
		Verbosef: logger.Debugf,
		Errorf:   logger.Errorf,
	}
}

func (i *UserDevice) Close() error {
	if err := i.userAPI.Close(); err != nil {
		return fmt.Errorf("failed to close user API: %w", err)
	}

	i.userDevice.Close()

	if err := i.BaseInterface.Close(); err != nil {
		return fmt.Errorf("failed to close interface: %w", err)
	}

	return nil
}

func (i *UserDevice) handleUserAPI() {
	for {
		conn, err := i.userAPI.Accept()
		if err != nil {
			i.logger.Warn("Failed to accept UAPI connection", zap.Error(err))
			return
		}

		go i.userDevice.IpcHandle(conn)
	}
}

func CreateUserInterface(name string, client *wgctrl.Client, backend signaling.Backend, events chan *pb.Event, cfg *config.Config) (Interface, error) {
	var err error
	logger := zap.L().With(
		zap.String("intf", name),
		zap.String("type", "user"),
	)

	dev := &UserDevice{
		log: newLogger(logger),
	}

	logger.Debug("Starting in-process wireguard-go interface")

	// Create TUN device
	dev.tun, err = tun.CreateTUN(name, device.DefaultMTU)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUN device: %w", err)
	}

	// Fix interface name
	realName, err := dev.tun.Name()
	if err == nil && realName != name {
		name = realName
	}

	// Open UAPI file (or use supplied fd)
	fileUAPI, err := ipc.UAPIOpen(name)
	if err != nil {
		return nil, fmt.Errorf("UAPI listen error: %w", err)
	}

	var bind conn.Bind
	if bind == nil {
		bind = conn.NewDefaultBind()
	}

	// Create new device
	dev.userDevice = device.NewDevice(dev.tun, bind, dev.log)

	logger.Debug("Device started")

	// Open UApi socket
	dev.userAPI, err = ipc.UAPIListen(name, fileUAPI)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on UAPI socket: %w", err)
	}

	// Handle UApi requests
	go dev.handleUserAPI()
	logger.Debug("UAPI listener started for interface")

	// Connect to UAPI
	wgDev, err := client.Device(name)
	if err != nil {
		return nil, err
	}

	dev.BaseInterface, err = NewInterface(wgDev, client, backend, events, cfg)
	if err != nil {
		return nil, err
	}

	return dev, nil
}
