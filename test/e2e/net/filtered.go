package net

import (
	"net"

	g "github.com/stv0g/gont/pkg"
	gopt "github.com/stv0g/gont/pkg/options"
	gfopt "github.com/stv0g/gont/pkg/options/filters"
)

func Filtered(p *NetworkParams) (*Network, error) {
	// We are dropped packets between the ɯice nodes to force ICE using the relay
	_, hostNetV4, err := net.ParseCIDR("10.0.1.0/24")
	if err != nil {
		return nil, err
	}

	_, hostNetV6, err := net.ParseCIDR("fc::1:0/112")
	if err != nil {
		return nil, err
	}

	p.HostOptions = append(p.HostOptions,
		gopt.Filter(g.FilterInput, gfopt.Source(hostNetV4), gfopt.Drop),
		gopt.Filter(g.FilterInput, gfopt.Source(hostNetV6), gfopt.Drop),
	)

	n, err := Simple(p)
	if err != nil {
		return nil, err
	}

	return n, nil
}
