package core

import (
	"riasc.eu/wice/pkg/crypto"
)

// InterfaceList stores all Wireguard interfaces indexed by their unique ifindex
type InterfaceList map[string]*Interface

func (l *InterfaceList) Close() error {
	for _, intf := range *l {
		intf.Close()
	}

	return nil
}

func (l *InterfaceList) ByIndex(index int) *Interface {
	for _, i := range *l {
		if i.KernelDevice.Index() == index {
			return i
		}
	}

	return nil
}

func (l *InterfaceList) ByName(name string) *Interface {
	return (*l)[name]
}

func (l *InterfaceList) ByPublicKey(pk crypto.Key) *Interface {
	for _, i := range *l {
		if i.PublicKey() == pk {
			return i
		}
	}

	return nil
}
