# Design

## Objectives

- Support [Trickle ICE]
- Support ICE restart
- Support [ICE-TCP]
- Sign and verify ICE offers with Wireguard keys (via [XEdDSA] signature scheme for Curve25519 key pairs)
- Seamless switch between ICE candidates and relays
- Zero configuration
  - Eleviate users of exchaging endpoint IPs & ports
- Enables direct communication of Wireguard peers behind NAT / UDP-blocking firewalls
- Single-binary, zero dependency installation
  - Bundled ICE agent & Wireguard userspace daemon
  - Portablilty
- Support for user and kernel-space Wireguard implementations
- Zero performance impact
  - Kernel-side filtering / redirection of Wireguard traffic
  - Fallback to userspace proxying only if no Kernel features are available 
- Minimized attack surface
  - Drop privileges after inital configuration
- Compatible with existing Wireguard configuration utilities like:
  - [NetworkManager]
  - [systemd-networkd]
  - [wg-quick]
  - [kilo]
  - [drago]
- Monitoring for new Wireguard interfaces and peers
  - Inotify for new UAPI sockets in /var/run/wireguard
  - Netlink subscription for link updates (patch is pending)

[kilo]: https://kilo.squat.ai
[drago]: https://seashell.github.io/drago/
[NetworkManager]: https://github.com/max-moser/network-manager-wireguard
[systemd-networkd]: https://www.freedesktop.org/software/systemd/man/systemd.netdev.html#%5BWireGuard%5D%20Section%20Options
[wg-quick]: https://manpages.debian.org/unstable/wireguard-tools/wg-quick.8.en.html

[ICE-TCP]: https://datatracker.ietf.org/doc/html/rfc6544
[Trickle ICE]: https://datatracker.ietf.org/doc/html/rfc8838
[XEdDSA]: https://signal.org/docs/specifications/xeddsa/
[ICE]: https://datatracker.ietf.org/doc/html/rfc8445
[ICE-PAC]: https://datatracker.ietf.org/doc/html/rfc8863
[ICE-SDP]: https://datatracker.ietf.org/doc/html/rfc8839
[TURN-TCP]: https://datatracker.ietf.org/doc/html/rfc6062
[TURN-STUN]: https://datatracker.ietf.org/doc/html/rfc8656
[STUN]: https://datatracker.ietf.org/doc/html/rfc8489
[SDP]: https://datatracker.ietf.org/doc/html/rfc8866
[SDP-Offer-Answer]: https://datatracker.ietf.org/doc/html/rfc3264
