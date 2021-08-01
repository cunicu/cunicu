package intf

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"riasc.eu/wice/pkg/args"
	"riasc.eu/wice/pkg/backend"
	nl "riasc.eu/wice/pkg/netlink"
)

type KernelInterface struct {
	BaseInterface

	created bool

	link netlink.Link
}

func (i *KernelInterface) Close() error {
	err := i.BaseInterface.Close()
	if err != nil {
		return err
	}

	if i.created {
		err := i.Delete()
		if err != nil {
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

	err := netlink.LinkDel(l)
	if err != nil {
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

func CreateKernelInterface(name string, client *wgctrl.Client, backend backend.Backend, args *args.Args) (Interface, error) {
	log.WithField("intf", name).Debug("Creating new kernel interface")

	l := &nl.Wireguard{
		LinkAttrs: netlink.NewLinkAttrs(),
	}
	l.LinkAttrs.Name = name
	err := netlink.LinkAdd(l)
	if err != nil {
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

	baseDev, err := NewInterface(wgDev, client, backend, args)
	if err != nil {
		return nil, err
	}

	i := &KernelInterface{
		BaseInterface: baseDev,
		created:       true,
		link:          link,
	}

	err = i.SetUp()
	if err != nil {
		return nil, fmt.Errorf("failed to bring link %s up: %w", name, err)
	}

	return i, nil
}
