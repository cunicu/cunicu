<p align="center" >
    <img style="width: 40%; margin: 4em 0" src="docs/images/cunicu_logo.svg" alt="cunīcu logo" />
</p>

[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/stv0g/cunicu/build?style=flat-square)](https://github.com/stv0g/cunicu/actions)
[![goreportcard](https://goreportcard.com/badge/github.com/stv0g/cunicu?style=flat-square)](https://goreportcard.com/report/github.com/stv0g/cunicu)
[![Codacy grade](https://img.shields.io/codacy/grade/4c4ecfff2f0d43948ded3d90f0bcf0cf?style=flat-square)](https://app.codacy.com/gh/stv0g/cunicu/)
[![Codecov](https://img.shields.io/codecov/c/github/stv0g/cunicu?token=WWQ6SR16LA&style=flat-square)](https://app.codecov.io/gh/stv0g/cunicu)
[![License](https://img.shields.io/github/license/stv0g/cunicu?style=flat-square)](https://github.com/stv0g/cunicu/blob/master/LICENSE)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/stv0g/cunicu?style=flat-square)
[![Go Reference](https://pkg.go.dev/badge/github.com/stv0g/cunicu.svg)](https://pkg.go.dev/github.com/stv0g/cunicu)

<!-- [![DOI](https://zenodo.org/badge/413409974.svg)](https://zenodo.org/badge/latestdoi/413409974) -->

## 🚧 cunīcu is currently still in an Alpha state and not usable yet

[cunīcu][cunicu] is a user-space daemon managing [WireGuard®][wireguard] interfaces to establish peer-to-peer connections in harsh network environments.

It relies on the [awesome](https://github.com/pion/awesome-pion) [pion/ice][pion-ice] package for the interactive connectivity establishment as well as bundles the Go user-space implementation of WireGuard in a single binary for environments in which WireGuard kernel support has not landed yet.

## Getting started

To use cunīcu follow these steps on each host:

1.  [Install cunīcu](docs/Installation.md)
2.  Configure your WireGuard interfaces using `wg`, `wg-quick` or [NetworkManager](https://blogs.gnome.org/thaller/2019/03/15/wireguard-in-networkmanager/)
3.  Start the cunīcu daemon by running: `sudo cunicu daemon`

Make sure that in step 2. you have created WireGuard keys and exchanged them by hand between the hosts.
cunīcu does not (yet) discover available peers. You are responsible to add the peers to the WireGuard interface by yourself.

After the cunīcu daemons have been started, they will attempt to discover valid endpoint addresses using the ICE protocol (e.g. contacting STUN servers).
These _ICE candidates_ are then exchanged via the signaling server and cunīcu will update the endpoint addresses of the WireGuard peers accordingly.
Once this has been done, the cunīcu logs should show a line `state=connected`.

## Documentation

Documentation of cunīcu can be found in the [`docs/`](./docs) directory.

## Authors

-   Steffen Vogel ([@stv0g](https://github.com/stv0g), Institute for Automation of Complex Power Systems, RWTH Aachen University)

## Join us

Please feel free to [join our Slack channel](https://join.slack.com/t/gophers/shared_invite/zt-1447h1rgj-s9W5BcyRzBxUwNAZJUKmaQ) `#cunicu` in the [Gophers workspace](https://gophers.slack.com/) and say 👋.

## Name

The project name _cunīcu_ [kʊˈniːkʊ] is derived from the [latin noun cunīculus](https://en.wiktionary.org/wiki/cuniculus#Latin) which means rabbit, a rabbit burrow or underground tunnel. We have choosen it as a name for this project as _cunīcu_ builds tunnels between otherwise hard to reach network locations.
It has been changed from the former name _wice_ in order to broaden the scope of the project and avoid any potential trademark violations. 

## License

cunīcu is licensed under the [Apache 2.0](./LICENSE) license.

Copyright 2022 Institute for Automation of Complex Power Systems, RWTH Aachen University

## Funding acknowledgement

<img alt="European Flag" src="./docs/images/flag_of_europe.svg" align="left" style="height: 4em; margin-right: 10px"/> The development of cunīcu has been supported by the [ERIGrid 2.0][erigrid] project of the H2020 Programme under [Grant Agreement No. 870620](https://cordis.europa.eu/project/id/870620)

[wireguard]: https://wireguard.com

[pion-ice]: https://github.com/pion/ice

[cunicu]: https://github.com/stv0g/cunicu

[erigrid]: https://erigrid2.eu

## Trademark

"WireGuard" and the "WireGuard" logo are [registered trademarks](https://www.wireguard.com/trademark-policy/) of Jason A. Donenfeld.