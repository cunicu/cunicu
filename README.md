# WICE - Wireguard Interactive Connectivity Establishment

[![Go Reference](https://pkg.go.dev/badge/github.com/stv0g/wice.svg)](https://pkg.go.dev/github.com/stv0g/wice)
![Snyk](https://img.shields.io/snyk/vulnerabilities/github/stv0g/wice)
[![Build](https://img.shields.io/github/checks-status/stv0g/wice/master)](https://github.com/stv0g/wice/actions)
[![Dependencies](https://img.shields.io/librariesio/release/stv0g/wice)](https://libraries.io/github/stv0g/wice)
[![GitHub](https://img.shields.io/github/license/stv0g/wice)](https://github.com/stv0g/wice/blob/master/LICENSE)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/stv0g/wice)

# 🚧 WICE is currently still in an Alpha state and not usable yet ⚠️

WICE is a userspace daemon managing Wireguard interfaces to establish peer-to-peer connections in harsh network environments.

It relies on the [awesome](https://github.com/pion/awesome-pion) [pion/ice] package for the interactive connectivity establishment as well as bundles the Go userspace implementation of Wiguard in a single binary for environments in which Wireguard kernel support has not landed yet.

## Getting started

To use WICE follow these steps on each host:

1. Install WICE: `go install riasc.eu/wice/cmd@latest`
2. Configure your Wireguard interfaces using `wg`, `wg-quick` or [NetworkManager](https://blogs.gnome.org/thaller/2019/03/15/wireguard-in-networkmanager/)
3. Start the WICE daemon by running: `sudo wice daemon`

Make sure that in step 2. you have created Wireguard keys and exchanged them by hand between the hosts.
WICE does not (yet) discover available peers. You are responsible to add the peers to the Wireguard interface by yourself.

After the WICE daemons have been started, they will attempt to discover valid endpoint addresses using the ICE protocol (e.g. contacting STUN servers).
These _ICE candidates_ are then exchanged via the signaling server and WICE will update the endpoint addresses of the Wireguard peers accordingly.
Once this has been done, the WICE logs should show a line `state=connected`.

## Documentation

Documentation of WICE can be found in the [`docs/`](./docs) directory.

## Authors

- Steffen Vogel ([@stv0g](https://github.com/stv0g), Institute for Automation of Complex Power Systems, RWTH Aachen University)

## Funding acknowledment

![Flag of Europe](https://erigrid2.eu/wp-content/uploads/2020/03/europa_flag_low.jpg) The development of [WICE] has been supported by the [ERIGrid 2.0] project of the H2020 Programme under [Grant Agreement No. 870620](https://cordis.europa.eu/project/id/870620)

[Wireguard]: https://wireguard.com
[wireguard-go]: https://git.zx2c4.com/wireguard-go
[pion/ice]: https://github.com/pion/ice
[ICE]: https://datatracker.ietf.org/doc/html/rfc8445
[ICE-PAC]: https://datatracker.ietf.org/doc/html/rfc8863
[ICE-TCP]: https://datatracker.ietf.org/doc/html/rfc6544
[Trickle ICE]: https://datatracker.ietf.org/doc/html/rfc8838
[ICE-SDP]: https://datatracker.ietf.org/doc/html/rfc8839
[TURN-TCP]: https://datatracker.ietf.org/doc/html/rfc6062
[TURN-STUN]: https://datatracker.ietf.org/doc/html/rfc8656
[STUN]: https://datatracker.ietf.org/doc/html/rfc8489
[SDP]: https://datatracker.ietf.org/doc/html/rfc8866
[SDP-Offer-Answer]: https://datatracker.ietf.org/doc/html/rfc3264
[WICE]: https://github.com/stv0g/wice
[ERIGrid 2.0]: https://erigrid2.eu
[NetworkManager]: https://github.com/max-moser/network-manager-wireguard
[systemd-networkd]: https://www.freedesktop.org/software/systemd/man/systemd.netdev.html#%5BWireGuard%5D%20Section%20Options
[wg-quick]: https://manpages.debian.org/unstable/wireguard-tools/wg-quick.8.en.html
[kilo]: https://kilo.squat.ai
[Nftables]: https://www.netfilter.org/projects/nftables/manpage.html
[XEdDSA]: https://signal.org/docs/specifications/xeddsa/

[Golang BPF]: https://riyazali.net/posts/berkeley-packet-filter-in-golang/
[Linux Raw Sockets]: https://squidarth.com/networking/systems/rc/2018/05/28/using-raw-sockets.html
