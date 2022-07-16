package core

import (
	"riasc.eu/wice/internal/wg"
)

// Interface add/remove

type InterfaceHandlerList []InterfaceHandler
type InterfaceHandler interface {
	OnInterfaceAdded(i *Interface)
	OnInterfaceRemoved(i *Interface)
}

func (hl *InterfaceHandlerList) Register(h InterfaceHandler) {
	*hl = append(*hl, h)
}

func (hl *InterfaceHandlerList) InvokeAdded(i *Interface) {
	for _, h := range *hl {
		h.OnInterfaceAdded(i)
	}
}

func (hl *InterfaceHandlerList) InvokeRemoved(i *Interface) {
	for _, h := range *hl {
		h.OnInterfaceRemoved(i)
	}
}

// Interface modified

type InterfaceModifiedHandlerList []InterfaceModifiedHandler
type InterfaceModifiedHandler interface {
	OnInterfaceModified(i *Interface, old *wg.Device, m InterfaceModifier)
}

func (hl *InterfaceModifiedHandlerList) Register(h InterfaceModifiedHandler) {
	*hl = append(*hl, h)
}

func (hl *InterfaceModifiedHandlerList) Invoke(i *Interface, old *wg.Device, mod InterfaceModifier) {
	for _, h := range *hl {
		h.OnInterfaceModified(i, old, mod)
	}
}
