// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package device

import (
	"errors"
	"fmt"
	"net"
	"os"
	"sync"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"

	"github.com/stv0g/cunicu/pkg/link"
	"github.com/stv0g/cunicu/pkg/log"
	"github.com/stv0g/cunicu/pkg/wg"
)

//nolint:gochecknoglobals
var (
	userDevices     = map[string]*UserDevice{}
	userDevicesLock sync.Mutex
)

// Compile-time assertions
var _ Device = (*UserDevice)(nil)

type UserDevice struct {
	link.Link
	*device.Device

	apiListener net.Listener

	logger *log.Logger
}

func NewUserDevice(name string) (*UserDevice, error) {
	var err error

	logger := log.Global.Named("dev").With(
		zap.String("dev", name),
		zap.String("type", "user"),
	)

	wgLogger := logger.Named("wg").Sugar()
	wgDeviceLogger := &device.Logger{
		Verbosef: wgLogger.Debugf,
		Errorf:   wgLogger.Errorf,
	}

	dev := &UserDevice{
		logger: logger,
	}

	// Create TUN device
	tunDev, err := tun.CreateTUN(name, device.DefaultMTU)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUN device: %w", err)
	}

	// Fix TUN device name
	realName, err := tunDev.Name()
	if err == nil && realName != name {
		logger.Debug("using real tun device name", zap.String("real", realName))
		name = realName
	}

	// Create new device
	dev.Device = device.NewDevice(tunDev, wg.NewBind(logger), wgDeviceLogger)

	// TODO: Check that this is a TUN link
	if dev.Link, err = link.FindLink(name); err != nil {
		return nil, fmt.Errorf("failed to find kernel device: %w", err)
	}

	// Open UAPI socket
	if dev.apiListener, err = ListenUAPI(name); err != nil {
		return nil, fmt.Errorf("failed to listen on UAPI socket: %w", err)
	}

	// Handle UApi requests
	go dev.handleUserAPI()

	logger.Info("Started in-process wireguard-go interface")

	// Register user device
	userDevicesLock.Lock()
	userDevices[name] = dev
	userDevicesLock.Unlock()
	return dev, nil
}

func FindUserDevice(name string) (Device, error) {
	// Register user device
	userDevicesLock.Lock()
	defer userDevicesLock.Unlock()

	if dev, ok := userDevices[name]; ok {
		return dev, nil
	}

	return nil, os.ErrNotExist
}

func (d *UserDevice) Close() error {
	d.Device.Close()

	return d.apiListener.Close()
}

func (d *UserDevice) Bind() *wg.Bind {
	return d.Device.Bind().(*wg.Bind) //nolint:forcetypeassert
}

func (d *UserDevice) handleUserAPI() {
	for {
		conn, err := d.apiListener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}

			d.logger.Error("Failed to accept new user api connection", zap.Error(err))
			continue
		}

		d.logger.Debug("Handle new IPC connection", zap.Any("socket", conn.LocalAddr()))
		go d.IpcHandle(conn)
	}
}
