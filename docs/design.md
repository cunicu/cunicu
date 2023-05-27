---
sidebar_position: 20
# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
---

# Design

## Architecture

![](/img/architecture.svg)

## Objectives

-   Encrypt all signaling messages

-   Plug-able signaling backends:
    -   GRPC
    -   Kubernetes API-server
    -   WebSocket

-   Support [Trickle ICE][rfc8838]

-   Support [ICE restart][rfc8445-ice-restart]

-   Support [ICE-TCP][rfc6544]

-   Encrypt exchanged ICE offers with WireGuard keys

-   Seamless switch between ICE candidates and relays

-   Zero configuration
    -   Alleviate users of exchanging endpoint IPs & ports

-   Enables direct communication of WireGuard peers behind NAT / UDP-blocking firewalls

-   Single-binary, zero dependency installation
    -   Bundled ICE agent & [WireGuard user-space daemon][wireguard-go]
    -   Portability

-   Support for user and kernel-space WireGuard implementations

-   Zero performance impact
    -   Kernel-side filtering / redirection of WireGuard traffic
    -   Fallback to user-space proxying only if no Kernel features are available 

-   Minimized attack surface
    -   Drop privileges after initial configuration

-   Compatible with existing WireGuard configuration utilities like:
    -   [NetworkManager][network-manager]
    -   [systemd-networkd][systemd-networkd]
    -   [wg-quick][wg-quick]
    -   [Kilo][kilo]
    -   [drago][drago]

-   Monitoring for new WireGuard interfaces and peers
    -   Inotify for new UAPI sockets in /var/run/wireguard
    -   Netlink subscription for link updates (patch is pending)

## Related RFCs

-   [RFC6544][rfc6544] TCP Candidates with Interactive Connectivity Establishment (ICE)
-   [RFC8838][rfc8838] Trickle ICE: Incremental Provisioning of Candidates for the Interactive Connectivity Establishment (ICE) Protocol
-   [RFC8445][rfc8445] Interactive Connectivity Establishment (ICE): A Protocol for Network Address Translator (NAT) Traversal
-   [RFC8863][rfc8863] Interactive Connectivity Establishment Patiently Awaiting Connectivity (ICE PAC)
-   [RFC8839][rfc8839] Session Description Protocol (SDP) Offer/Answer Procedures for Interactive Connectivity Establishment (ICE)
-   [RFC6062][rfc6062] Traversal Using Relays around NAT (TURN) Extensions for TCP Allocations
-   [RFC8656][rfc8656] Traversal Using Relays around NAT (TURN): Relay Extensions to Session Traversal Utilities for NAT (STUN)
-   [RFC8489][rfc8489] Session Traversal Utilities for NAT (STUN)
-   [RFC8866][rfc8866] SDP: Session Description Protocol
-   [RFC3264][rfc3264] An Offer/Answer Model with the Session Description Protocol (SDP)
-   [RFC7064][rfc7064] URI Scheme for the Session Traversal Utilities for NAT (STUN) Protocol
-   [RFC7065][rfc7065] Traversal Using Relays around NAT (TURN) Uniform Resource Identifiers

[wireguard-go]: https://git.zx2c4.com/wireguard-go

[kilo]: https://kilo.squat.ai

[drago]: https://seashell.github.io/drago/

[network-manager]: https://github.com/max-moser/network-manager-wireguard

[systemd-networkd]: https://www.freedesktop.org/software/systemd/man/systemd.netdev.html#%5BWireGuard%5D%20Section%20Options

[wg-quick]: https://manpages.debian.org/unstable/wireguard-tools/wg-quick.8.en.html

[rfc6544]: https://datatracker.ietf.org/doc/html/rfc6544

[rfc8838]: https://datatracker.ietf.org/doc/html/rfc8838

[rfc8445-ice-restart]: https://datatracker.ietf.org/doc/html/rfc8445#section-2.4

[rfc8445]: https://datatracker.ietf.org/doc/html/rfc8445

[rfc8863]: https://datatracker.ietf.org/doc/html/rfc8863

[rfc8839]: https://datatracker.ietf.org/doc/html/rfc8839

[rfc6062]: https://datatracker.ietf.org/doc/html/rfc6062

[rfc8656]: https://datatracker.ietf.org/doc/html/rfc8656

[rfc8489]: https://datatracker.ietf.org/doc/html/rfc8489

[rfc8866]: https://datatracker.ietf.org/doc/html/rfc8866

[rfc3264]: https://datatracker.ietf.org/doc/html/rfc3264

[rfc7064]: https://datatracker.ietf.org/doc/html/rfc7064

[rfc7065]: https://datatracker.ietf.org/doc/html/rfc7065
