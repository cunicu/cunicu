//go:build linux

package mtudisc

import (
	"errors"
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"go.uber.org/zap"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"golang.org/x/sys/unix"
)

const (
	ICMPv4CodeFragmentationRequired uint8 = 4
)

func (p *Peer) DiscoverPathMTU() (int, error) {
	var err error

	isV4 := p.Endpoint.IP.To4() != nil

	if p.fd >= 0 {
		if err := syscall.Close(p.fd); err != nil {
			return -1, fmt.Errorf("failed to close connection: %w", err)
		}
	}

	// TODO: Should we use the local candidate from a
	// selected candidate pair to fix our local address here?
	p.fd, err = syscall.Socket(syscall.AF_INET6, syscall.SOCK_DGRAM, 0)
	if err != nil {
		return -1, fmt.Errorf("failed to create connection: %w", err)
	}

	var saddr syscall.Sockaddr
	if isV4 {
		saddr = &syscall.SockaddrInet4{
			Addr: *(*[4]byte)(p.Endpoint.IP.To4()),
			Port: p.Endpoint.Port,
		}
	} else {
		saddr = &syscall.SockaddrInet6{
			Addr: *(*[16]byte)(p.Endpoint.IP.To16()),
			Port: p.Endpoint.Port,
		}
	}

	if err = syscall.Connect(p.fd, saddr); err != nil {
		return -1, fmt.Errorf("failed to connect: %w", err)
	}

	// Enable path discovery
	if isV4 {
		if err = syscall.SetsockoptInt(p.fd, syscall.IPPROTO_IP, syscall.IP_MTU_DISCOVER, syscall.IP_PMTUDISC_DO); err != nil {
			return -1, fmt.Errorf("failed to enable path MTU discovery: %w", err)
		}
	} else {
		if err = syscall.SetsockoptInt(p.fd, syscall.IPPROTO_IPV6, syscall.IPV6_MTU_DISCOVER, syscall.IPV6_PMTUDISC_DO); err != nil {
			return -1, fmt.Errorf("failed to enable path MTU discovery: %w", err)
		}
	}

	// Keep watching for changes Path MTU
	go p.monitorPathMTU(p.fd)

	// Get initial estimate of MTU
	var mtu int
	if isV4 {
		mtu, err = syscall.GetsockoptInt(p.fd, syscall.SOL_IP, syscall.IP_MTU)
		if err != nil {
			return -1, fmt.Errorf("failed to get path MTU: %w", err)
		}
	} else {
		mtu, err = syscall.GetsockoptInt(p.fd, syscall.SOL_IPV6, syscall.IPV6_MTU)
		if err != nil {
			return -1, fmt.Errorf("failed to get path MTU: %w", err)
		}
	}

	p.logger.Debug("Using initial path MTU", zap.Int("mtu", mtu))

	return mtu, nil
}

func (p *Peer) monitorPathMTU(fd int) {
	if isV4 := p.Endpoint.IP.To4() != nil; isV4 {
		if err := syscall.SetsockoptInt(fd, syscall.SOL_IP, syscall.IP_RECVERR, 1); err != nil {
			p.logger.Error("Failed to enable watch for path MTU changes", zap.Error(err))
		}
	} else {
		if err := syscall.SetsockoptInt(fd, syscall.SOL_IPV6, syscall.IPV6_RECVERR, 1); err != nil {
			p.logger.Error("Failed to enable watch for path MTU changes", zap.Error(err))
		}
	}

	// TODO: configurable recv size?
	buf := make([]byte, 1500)
	oob := make([]byte, 1500)

	for {
		_, oobn, _, _, err := syscall.Recvmsg(fd, buf, oob, syscall.MSG_ERRQUEUE)
		if err != nil || oobn <= 0 {
			if errors.Is(err, syscall.EAGAIN) {
				time.Sleep(1 * time.Second)
				continue
			}

			p.logger.Error("Failed recvmsg", zap.Error(err), zap.Int("n", oobn))

			if errors.Is(err, syscall.EBADF) {
				break
			} else {
				continue
			}
		}

		cmsgs, err := syscall.ParseSocketControlMessage(oob)
		if err != nil {
			p.logger.Error("Failed to parse cmsg", zap.Error(err))
		}

		for _, cmsg := range cmsgs {
			se := (*unix.SockExtendedErr)(unsafe.Pointer(&cmsg.Data[0]))

			if cmsg.Header.Type != syscall.IP_RECVERR {
				continue
			}

			switch cmsg.Header.Level {
			case syscall.SOL_IP:
				if se.Origin != unix.SO_EE_ORIGIN_ICMP ||
					se.Type != uint8(ipv4.ICMPTypeDestinationUnreachable) ||
					se.Code != ICMPv4CodeFragmentationRequired {
					continue
				}
			case syscall.SOL_IPV6:
				if se.Origin != unix.SO_EE_ORIGIN_ICMP6 ||
					se.Type != uint8(ipv6.ICMPTypePacketTooBig) ||
					se.Code != 0 {
					continue
				}
			}

			p.MTU = int(se.Info)
		}

		i := p.Interface
		if err := i.UpdateMTU(); err != nil {
			i.logger.Error("Failed to update MTU", zap.Error(err))
		}
	}
}
