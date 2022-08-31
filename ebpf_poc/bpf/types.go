package bpf

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go@master -type state bpf kern/main.c

import (
	"github.com/cilium/ebpf"
	// #include "kern/types.h"
	"C"
)
import (
	"errors"
	"fmt"
	"net"
)

type Objects struct {
	Programs bpfPrograms
	Maps     Maps
}

func (o *Objects) Close() error {
	if err := o.Maps.Close(); err != nil {
		return err
	}

	if err := o.Programs.Close(); err != nil {
		return err
	}

	return nil
}

type Maps struct {
	EgressMap   MapState
	IngressMap  MapState
	SettingsMap MapSettings
}

func (m *Maps) Close() error {
	if err := m.EgressMap.Close(); err != nil {
		return err
	}

	if err := m.IngressMap.Close(); err != nil {
		return err
	}

	if err := m.SettingsMap.Close(); err != nil {
		return err
	}

	return nil
}

type MapSettings struct {
	*ebpf.Map
}

func (m *MapSettings) EnableDebug() error {
	return m.Put(uint32(C.SETTING_DEBUG), uint32(1))
}

type MapState struct {
	*ebpf.Map
}

type MapStateEntry = bpfState

func (m *MapState) AddEntry(addr *net.UDPAddr, me *MapStateEntry) error {
	inner, err := m.getOrCreateInnerMap(addr)
	if err != nil {
		return err
	}

	return inner.Update(addr.IP.To4(), me, ebpf.UpdateNoExist)
}

func (m *MapState) GetEntry(addr *net.UDPAddr) (*MapStateEntry, error) {
	inner, err := m.getInnerMap(addr)
	if err != nil {
		return nil, err
	}

	var me MapStateEntry
	return &me, inner.Lookup(addr.IP.To4(), &me)
}

func (m *MapState) DeleteEntry(addr *net.UDPAddr) error {
	inner, err := m.getInnerMap(addr)
	if err != nil {
		return err
	}

	return inner.Delete(addr.IP.To4())
}

func (m *MapState) getOrCreateInnerMap(addr *net.UDPAddr) (*ebpf.Map, error) {
	inner, err := m.getInnerMap(addr)
	if errors.Is(err, ebpf.ErrKeyNotExist) {
		if inner, err = ebpf.NewMap(&stateMapInnerSpec); err != nil {
			return nil, fmt.Errorf("failed to create new inner map: %w", err)
		}

		up := uint16(addr.Port)
		if err := m.Put(&up, inner); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return inner, nil
}

func (m *MapState) getInnerMap(addr *net.UDPAddr) (*ebpf.Map, error) {
	up := uint16(addr.Port)

	var inner *ebpf.Map
	if err := m.Lookup(&up, &inner); err != nil {
		return nil, err
	}

	return inner, nil
}
