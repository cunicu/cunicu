package device

import (
	"errors"
	"fmt"
	"net"

	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/ipc"
	"golang.zx2c4.com/wireguard/tun"
)

type UserDevice struct {
	KernelDevice

	userDevice *device.Device
	userAPI    net.Listener

	logger *zap.Logger
}

func NewUserDevice(name string) (KernelDevice, error) {
	var err error

	logger := zap.L().Named("interface").With(
		zap.String("intf", name),
		zap.String("type", "user"),
	)

	wgLogger := newWireGuardLogger()

	dev := &UserDevice{
		logger: logger,
	}

	logger.Debug("Starting in-process wireguard-go interface")

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

	// Open UAPI file
	fileUAPI, err := ipc.UAPIOpen(name)
	if err != nil {
		return nil, fmt.Errorf("UAPI listen error: %w", err)
	}

	// Create new device
	bind := conn.NewDefaultBind()
	dev.userDevice = device.NewDevice(tunDev, bind, wgLogger)

	logger.Debug("Device started")

	if dev.KernelDevice, err = FindDevice(name); err != nil {
		return nil, err
	}

	// Open UApi socket
	if dev.userAPI, err = ipc.UAPIListen(name, fileUAPI); err != nil {
		return nil, fmt.Errorf("failed to listen on UAPI socket: %w", err)
	}

	// Handle UApi requests
	go dev.handleUserAPI()

	logger.Debug("UAPI listener started for interface")

	return dev, nil
}

func newWireGuardLogger() *device.Logger {
	logger := zap.L().Named("wireguard").Sugar()

	return &device.Logger{
		Verbosef: logger.Debugf,
		Errorf:   logger.Errorf,
	}
}

func (i *UserDevice) Close() error {
	i.userDevice.Close()

	if err := i.userAPI.Close(); err != nil {
		return err
	}

	if err := i.KernelDevice.Close(); err != nil {
		return fmt.Errorf("failed to close kernel device: %w", err)
	}

	return nil
}

func (i *UserDevice) handleUserAPI() {
	for {
		if conn, err := i.userAPI.Accept(); err == nil {
			go i.userDevice.IpcHandle(conn)
		} else if errors.Is(err, unix.EALREADY) {
			i.logger.Warn("Failed to accept UAPI connection", zap.Error(err))
		}
	}
}
