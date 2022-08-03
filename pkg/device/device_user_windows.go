package device

import (
	"fmt"
	"net"

	"golang.zx2c4.com/wireguard/ipc"
)

func ListenUAPI(name string) (listener net.Listener, err error) {
	if listener, err = ipc.UAPIListen(name); err != nil {
		return nil, fmt.Errorf("failed to listen on UAPI socket: %w", err)
	}

	return listener, nil
}
