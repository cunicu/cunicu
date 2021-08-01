package util_test

import (
	"net"
	"testing"

	"riasc.eu/wice/internal/util"
)

func TestCmpEndpointEqual(t *testing.T) {
	a := net.UDPAddr{
		IP:   net.ParseIP("1.1.1.1"),
		Port: 1,
	}

	if util.CmpEndpoint(&a, &a) != 0 {
		t.Fail()
	}
}

func TestCmpEndpointUnequal(t *testing.T) {
	a := net.UDPAddr{
		IP:   net.ParseIP("1.1.1.1"),
		Port: 1,
	}

	b := net.UDPAddr{
		IP:   net.ParseIP("2.2.2.2"),
		Port: 1,
	}

	if util.CmpEndpoint(&a, &b) == 0 {
		t.Fail()
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	r, err := util.GenerateRandomBytes(16)
	if err != nil {
		t.Fail()
	}

	if len(r) != 16 {
		t.Fail()
	}
}

func TestCmpNetEqual(t *testing.T) {
	_, a, err := net.ParseCIDR("1.1.1.1/0")
	if err != nil {
		t.Fail()
	}

	if util.CmpNet(a, a) != 0 {
		t.Fail()
	}
}

func TestCmpNetUnequal(t *testing.T) {
	_, a, err := net.ParseCIDR("1.1.1.1/0")
	if err != nil {
		t.Fail()
	}

	_, b, err := net.ParseCIDR("1.1.1.1/1")
	if err != nil {
		t.Fail()
	}

	if util.CmpNet(a, b) == 0 {
		t.Fail()
	}
}
