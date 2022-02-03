# Design

## Objectives

- Support [Trickle ICE][rfc8838]
- Support [ICE restart][rfc8445-ice-restart]
- Support [ICE-TCP][rfc6544]
- Encrypt exchanged ICE offers with Wireguard keys
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
  - [NetworkManager][network-manager]
  - [systemd-networkd][systemd-networkd]
  - [wg-quick][wg-quick]
  - [Kilo][kilo]
  - [drago][drago]
- Monitoring for new Wireguard interfaces and peers
  - Inotify for new UAPI sockets in /var/run/wireguard
  - Netlink subscription for link updates (patch is pending)

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
