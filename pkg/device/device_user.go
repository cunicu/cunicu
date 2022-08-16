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
	"riasc.eu/wice/pkg/wg"
)

var (
	userDevices     = map[string]*UserDevice{}
	userDevicesLock sync.Mutex
)

type UserDevice struct {
	Device

	device *device.Device
	api    net.Listener
	Bind   *wg.UserBind

	logger *zap.Logger
}

func NewUserDevice(name string) (*UserDevice, error) {
	var err error

	logger := zap.L().Named("device").With(
		zap.String("dev", name),
		zap.String("type", "user"),
	)

	wgLogger := logger.Named("wg").Sugar()
	wgDeviceLogger := &device.Logger{
		Verbosef: wgLogger.Debugf,
		Errorf:   wgLogger.Errorf,
	}

	dev := &UserDevice{
		Bind:   wg.NewUserBind(),
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

	// Open UAPI socket
	if dev.api, err = ListenUAPI(name); err != nil {
		return nil, err
	}

	// Handle UApi requests
	go dev.handleUserAPI()

	// Create new device
	dev.device = device.NewDevice(tunDev, dev.Bind, wgDeviceLogger)

	if dev.Device, err = FindKernelDevice(name); err != nil {
		return nil, err
	}

	logger.Info("Started in-process wireguard-go interface")

	// Register user device
	userDevicesLock.Lock()
	defer userDevicesLock.Unlock()

	userDevices[name] = dev

	return dev, nil
}

func (i *UserDevice) Close() error {
	i.device.Close()

	if err := i.api.Close(); err != nil {
		return err
	}

	if err := i.Device.Close(); err != nil {
		return fmt.Errorf("failed to close kernel device: %w", err)
	}

	return nil
}

func (i *UserDevice) Delete() error {
	return nil
}

func (i *UserDevice) handleUserAPI() {
	for {
		conn, err := i.api.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			} else {
				i.logger.Error("Failed to accept new user api connection", zap.Error(err))
				continue
			}
		} else if i.device == nil {
			i.logger.Warn("Dropping user api connection as device is not ready yet")
			continue
		}

		i.logger.Debug("Handle new IPC connection", zap.String("socket", conn.LocalAddr().String()))
		go i.device.IpcHandle(conn)
	}
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
