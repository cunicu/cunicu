---
title: MTU Discovery
---

# MTU Discovery

The MTU discovery feature determines correct Maximum Transfer Unit (MTU) sizes for WireGuard tunnel interfaces.
Currently, it only takes into account local per-link and per-route MTUs configured in the system.
To do so, it follows the same algorithm as `wg-quick` for determining the MTU:

> The MTU is automatically determined from the endpoint addresses or the system default route, which is usually a sane choice.

If a WireGuard interface is connected to multiple peers, the smallest of the per-peer MTUs is used for the interface.

## Future developments

We plan to also utilize Path MTU Discovery (PMTUD) to considers not only local MTUs but also restrictions of the MTU sizes along the the full path.

We plan to look into utilizing per-route MTUs, to allow using different MTUs for different peers in case a WireGuard interface is connected to more than a single peer.