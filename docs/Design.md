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
- Monitoring for new Wireguard interfaces and peers
  - Inotify for new UAPI sockets in /var/run/wireguard
  - Netlink subscription for link updates
