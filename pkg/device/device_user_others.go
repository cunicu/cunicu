//go:build !windows

package device

import (
	"fmt"
	"net"

	"golang.zx2c4.com/wireguard/ipc"
)

func ListenUAPI(name string) (net.Listener, error) {
	file, err := ipc.UAPIOpen(name)
	if err != nil {
		return nil, fmt.Errorf("UAPI listen error: %w", err)
	}

	var listener net.Listener
	if listener, err = ipc.UAPIListen(name, file); err != nil {
		return nil, fmt.Errorf("failed to listen on UAPI socket: %w", err)
	}

	return listener, nil
}
