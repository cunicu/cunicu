// +build !windows

/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2017-2021 WireGuard LLC. All Rights Reserved.
 */

package intf

import (
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
	"riasc.eu/wice/pkg/args"
	"riasc.eu/wice/pkg/backend"

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

func newLogger(log *log.Entry) *device.Logger {
	logger := log.WithField("logger", "wireguard")

	return &device.Logger{
		Verbosef: logger.Debugf,
		Errorf:   logger.Errorf,
	}
}

func (i *UserDevice) Close() error {
	err := i.userAPI.Close()
	if err != nil {
		return err
	}

	i.userDevice.Close()

	err = i.BaseInterface.Close()
	if err != nil {
		return err
	}

	return nil
}

func (i *UserDevice) handleUserApi() {
	for {
		conn, err := i.userAPI.Accept()
		if err != nil {
			i.logger.WithError(err).Warn("Failed to accept UAPI connection")
			return
		}

		go i.userDevice.IpcHandle(conn)
	}
}

func CreateUserInterface(name string, client *wgctrl.Client, backend backend.Backend, args *args.Args) (Interface, error) {
	var err error
	logger := log.WithFields(log.Fields{
		"intf": name,
		"type": "user",
	})

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

	var bind conn.Bind = nil
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
	go dev.handleUserApi()
	logger.Debug("UAPI listener started for interface")

	// Connect to UAPI
	wgDev, err := client.Device(name)
	if err != nil {
		return nil, err
	}

	dev.BaseInterface, err = NewInterface(wgDev, client, backend, args)
	if err != nil {
		return nil, err
	}

	return dev, nil
}
