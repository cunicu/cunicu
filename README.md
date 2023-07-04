<!--
SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
SPDX-License-Identifier: Apache-2.0
-->

<div align="center" >
    <img style="width: 40%; margin: 4em 0" src="website/static/img/cunicu_logo.svg" alt="cunīcu logo" />

[![GitHub build](https://img.shields.io/github/actions/workflow/status/stv0g/cunicu/build.yaml?style=flat-square)](https://github.com/stv0g/cunicu/actions)
[![goreportcard](https://goreportcard.com/badge/github.com/stv0g/cunicu?style=flat-square)](https://goreportcard.com/report/github.com/stv0g/cunicu)
[![Codecov](https://img.shields.io/codecov/c/github/stv0g/cunicu?token=WWQ6SR16LA&style=flat-square)](https://app.codecov.io/gh/stv0g/cunicu)
[![License](https://img.shields.io/github/license/stv0g/cunicu?style=flat-square)](https://github.com/stv0g/cunicu/blob/master/LICENSE)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/stv0g/cunicu?style=flat-square)
[![Go Reference](https://pkg.go.dev/badge/github.com/stv0g/cunicu.svg)](https://pkg.go.dev/github.com/stv0g/cunicu)
</div>

## 🚧 cunīcu is currently still in an Alpha state and not usable yet

[cunīcu][cunicu] is a user-space daemon managing [WireGuard®][wireguard] interfaces to establish a mesh of peer-to-peer VPN connections in harsh network environments.

To achieve this, cunīcu utilizes a signaling layer to exchange peer information such as public encryption keys, hostname, advertised networks and reachability information to automate the configuration of the networking links.
From a user perspective, cunīcu alleviates the need of manual configuration such as exchange of public keys, IP addresses, endpoints, etc..
Hence, it adopts the design goals of the WireGuard project, to be simple and easy to use.

Thanks to [Interactive Connectivity Establishment (ICE)](https://en.wikipedia.org/wiki/Interactive_Connectivity_Establishment), cunīcu is capable to establish direct connections between peers which are located behind NAT firewalls such as home routers.
In situations where ICE fails, or direct UDP connectivity is not available, cunīcu falls back to using TURN relays to reroute traffic over an intermediate hop or encapsulate the WireGuard traffic via TURN-TCP.

It relies on the [awesome](https://github.com/pion/awesome-pion) [pion/ice][pion-ice] package for ICE as well as bundles the a Go user-space implementation of WireGuard in a single binary for systems in which WireGuard kernel support has not landed yet.

With these features, cunīcu can be used to quickly build multi-agent systems or connect field devices such as power grid monitoring infrastructure into a fully connected mesh.
Within the [ERIGrid 2.0 project](https://erigrid2.eu), cunīcu is used to interconnect smart grid laboratories for geographically distributed simulation of energy systems.

## Documentation

cunīcu's documentation can be found here: [cunicu.li/docs](https://cunicu.li/docs).

## Authors

-   Steffen Vogel ([@stv0g](https://github.com/stv0g))

## License

cunīcu is licensed under the [Apache 2.0](./LICENSE) license.

## Funding acknowledgement

The project has been initiated by Steffen Vogel while working at the [Institute for Automation of Complex Power Systems (ACS)](https://www.acs.eonerc.rwth-aachen.de) of [RWTH Aachen University](https://www.rwth-aachen.de).

<img alt="European Flag" src="website/static/img/flag_of_europe.svg" align="left" style="height: 4em; margin-right: 10px"/> The development of cunīcu has been supported by the [ERIGrid 2.0][erigrid] project of the H2020 Programme under [Grant Agreement No. 870620](https://cordis.europa.eu/project/id/870620)

## Trademark

_WireGuard_ and the _WireGuard_ logo are [registered trademarks](https://www.wireguard.com/trademark-policy/) of Jason A. Donenfeld.

[wireguard]: https://wireguard.com

[pion-ice]: https://github.com/pion/ice

[cunicu]: https://github.com/stv0g/cunicu

[erigrid]: https://erigrid2.eu
