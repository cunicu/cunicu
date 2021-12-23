package intf

import (
	"fmt"

	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"riasc.eu/wice/pkg/args"
	nl "riasc.eu/wice/pkg/netlink"
	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/socket"
)

type KernelInterface struct {
	BaseInterface

	created bool

	link netlink.Link
}

func (i *KernelInterface) Close() error {

	if err := i.BaseInterface.Close(); err != nil {
		return err
	}

	if i.created {

		if err := i.Delete(); err != nil {
			return err
		}
	}

	return nil
}

func (i *KernelInterface) Delete() error {
	i.logger.Debug("Deleting kernel device")

	l := &nl.Wireguard{
		LinkAttrs: netlink.NewLinkAttrs(),
	}
	l.LinkAttrs.Name = i.Name()

	if err := netlink.LinkDel(l); err != nil {
		return fmt.Errorf("failed to delete Wireguard device: %w", err)
	}

	return nil
}

func (i *KernelInterface) SetMTU(mtu int) error {
	i.logger.Debug("Set link MTU")
	return netlink.LinkSetMTU(i.link, mtu)
}

func (i *KernelInterface) SetUp() error {
	i.logger.Debug("Set link up")
	return netlink.LinkSetUp(i.link)
}

func (i *KernelInterface) SetDown(mtu int) error {
	i.logger.Debug("Set link down")
	return netlink.LinkSetDown(i.link)
}

func CreateKernelInterface(name string, client *wgctrl.Client, backend signaling.Backend, server *socket.Server, args *args.Args) (Interface, error) {
	zap.L().Debug("Creating new kernel interface", zap.String("intf", name))

	l := &nl.Wireguard{
		LinkAttrs: netlink.NewLinkAttrs(),
	}
	l.LinkAttrs.Name = name

	if err := netlink.LinkAdd(l); err != nil {
		return nil, fmt.Errorf("failed to create Wireguard interface: %w", err)
	}

	link, err := netlink.LinkByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get link details: %w", err)
	}

	// Connect to UAPI
	wgDev, err := client.Device(name)
	if err != nil {
		return nil, err
	}

	baseDev, err := NewInterface(wgDev, client, backend, server, args)
	if err != nil {
		return nil, err
	}

	i := &KernelInterface{
		BaseInterface: baseDev,
		created:       true,
		link:          link,
	}

	if err = i.SetUp(); err != nil {
		return nil, fmt.Errorf("failed to bring link %s up: %w", name, err)
	}

	return i, nil
}
