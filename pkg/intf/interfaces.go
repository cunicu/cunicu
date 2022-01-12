package intf

type InterfaceList []Interface

func (l *InterfaceList) Close() error {
	for _, intf := range *l {
		intf.Close()
	}

	return nil
}

func (l *InterfaceList) GetByName(name string) Interface {
	for _, intf := range *l {
		if intf.Name() == name {
			return intf
		}
	}

	return nil
}
