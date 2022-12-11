package device

import (
	"net"

	"golang.zx2c4.com/wireguard/ipc"
)

func ListenUAPI(name string) (listener net.Listener, err error) {
	return ipc.UAPIListen(name)
}
