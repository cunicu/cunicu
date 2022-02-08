# WICE - Wireguard Interactive Connectivity Establishment

[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/stv0g/wice/build?style=flat-square)](https://github.com/stv0g/wice/actions)
[![goreportcard](https://goreportcard.com/badge/github.com/stv0g/wice?style=flat-square)](https://goreportcard.com/report/riasc.eu/wice)
[![Codacy grade](https://img.shields.io/codacy/grade/4c4ecfff2f0d43948ded3d90f0bcf0cf?style=flat-square)](https://app.codacy.com/gh/stv0g/wice/)
[![Codacy coverage](https://img.shields.io/codacy/coverage/4c4ecfff2f0d43948ded3d90f0bcf0cf?style=flat-square)](https://app.codacy.com/gh/stv0g/wice/)
[![libraries.io](https://img.shields.io/librariesio/github/stv0g/wice?style=flat-square)](https://libraries.io/github/stv0g/wice)
[![License](https://img.shields.io/github/license/stv0g/wice?style=flat-square)](https://github.com/stv0g/wice/blob/master/LICENSE)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/stv0g/wice?style=flat-square)
[![Go Reference](https://pkg.go.dev/badge/github.com/stv0g/wice.svg)](https://pkg.go.dev/github.com/stv0g/wice)

<!-- [![DOI](https://zenodo.org/badge/413409974.svg)](https://zenodo.org/badge/latestdoi/413409974) -->

## 🚧 WICE is currently still in an Alpha state and not usable yet

WICE is a userspace daemon managing [Wireguard][wireguard] interfaces to establish peer-to-peer connections in harsh network environments.

It relies on the [awesome](https://github.com/pion/awesome-pion) [pion/ice][pion-ice] package for the interactive connectivity establishment as well as bundles the Go userspace implementation of Wiguard in a single binary for environments in which Wireguard kernel support has not landed yet.

## Getting started

To use WICE follow these steps on each host:

1.  Install WICE: `go install riasc.eu/wice/cmd/wice@latest`
2.  Configure your Wireguard interfaces using `wg`, `wg-quick` or [NetworkManager](https://blogs.gnome.org/thaller/2019/03/15/wireguard-in-networkmanager/)
3.  Start the WICE daemon by running: `sudo wice daemon`

Make sure that in step 2. you have created Wireguard keys and exchanged them by hand between the hosts.
WICE does not (yet) discover available peers. You are responsible to add the peers to the Wireguard interface by yourself.

After the WICE daemons have been started, they will attempt to discover valid endpoint addresses using the ICE protocol (e.g. contacting STUN servers).
These _ICE candidates_ are then exchanged via the signaling server and WICE will update the endpoint addresses of the Wireguard peers accordingly.
Once this has been done, the WICE logs should show a line `state=connected`.

## Documentation

Documentation of WICE can be found in the [`docs/`](./docs) directory.

## Authors

-   Steffen Vogel ([@stv0g](https://github.com/stv0g), Institute for Automation of Complex Power Systems, RWTH Aachen University)

## Funding acknowledment

![Flag of Europe](https://erigrid2.eu/wp-content/uploads/2020/03/europa_flag_low.jpg) The development of [WICE][wice] has been supported by the [ERIGrid 2.0][erigrid] project of the H2020 Programme under [Grant Agreement No. 870620](https://cordis.europa.eu/project/id/870620)

[wireguard]: https://wireguard.com

[pion-ice]: https://github.com/pion/ice

[wice]: https://github.com/stv0g/wice

[erigrid]: https://erigrid2.eu
