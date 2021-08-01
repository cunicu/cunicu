package intf

import "golang.zx2c4.com/wireguard/wgctrl/wgtypes"

type Devices []*wgtypes.Device

func (devs *Devices) GetByName(name string) *wgtypes.Device {
	for _, dev := range *devs {
		if dev.Name == name {
			return dev
		}
	}

	return nil
}
