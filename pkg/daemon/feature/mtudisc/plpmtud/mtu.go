package plpmtud

import "golang.org/x/exp/slices"

const (
	OverheadWireGuard = 32 // https://lists.zx2c4.com/pipermail/wireguard/2017-December/002201.html
	OverheadPPPoE     = 8
	OverheadL2TP      = 12
	OverheadVXLAN     = 8
	OverheadGRE       = 4
	OverheadIPv4      = 20
	OverheadIPv6      = 40
	OverheadUDP       = 8
	Overhead802_3     = 14 // Ethernet header
	Overhead802_1q    = 4  // VLAN tag

	MtuIPv6Minimal    = 1280
	Mtu802_3Standard  = 1500
	Mtu802_3Jumbo     = 9000
	Mtu802_11         = 2304
	MtuLoopbackDarwin = 1 << 14
	MtuLoopbackLinux  = 1 << 16
)

var (
	WellKnownMTUs = CalculateWellKnownMTUs()
)

func CalculateWellKnownMTUs() []uint {
	mtus := []uint{}

	// Minimal IPv6 MTU can not be further encapsulated
	mtus = append(mtus, MtuIPv6Minimal)

	// There is usually no encapsulation over loopback interfaces..
	mtus = append(mtus, MtuLoopbackDarwin, MtuLoopbackLinux)

	// PPPoE usually only over standard Ethernet MTU
	mtus = append(mtus, Mtu802_3Standard-OverheadPPPoE)

	for _, base := range []uint{Mtu802_3Standard, Mtu802_3Jumbo} {
		mtus = append(mtus, base)
		mtus = append(mtus, base-20)

		for _, proto := range []uint{OverheadIPv4, OverheadIPv6} {
			base := base - proto

			// Encapsulation
			mtus = append(mtus, base-OverheadGRE)
			mtus = append(mtus, base-OverheadUDP-OverheadWireGuard)
			mtus = append(mtus, base-OverheadUDP-OverheadVXLAN-Overhead802_3)
		}
	}

	slices.Sort(mtus)
	slices.Compact(mtus)

	return mtus
}
