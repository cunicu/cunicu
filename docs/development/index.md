---
sidebar_position: 199
# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
---

# Development

cunīcu is written almost completely in [Go](https://go.dev/) and heavily relies on awesome tooling and packages for Golang:

-   [GoReleaser](https://goreleaser.com/) for release automation
-   [Ginkgo](https://onsi.github.io/ginkgo) and [Gomega](https://onsi.github.io/gomega) for testing
-   [Pion](https://github.com/pion) for its ICE, STUN, TURN implementation
-   [Gont](https://github.com/stv0g/gont) for end-to-end testing in various network topologies

Furthermore use the following services to manage our development:

-   [GitHub](https://github.com/stv0g/cunicu) for source code management and CI pipelines
-   [Codecov](https://app.codecov.io/gh/stv0g/cunicu) for code coverage analysis

## Testing

We aim to maintain a test coverage above 80% of the lines of code.
Please make sure that your merge requests include tests which do not lower the coverage percentage.

cunīcu's code-base is tested using the Ginkgo/Gomega testing framework.
Unit tests can be found alongside the code in files with a `_test.go` suffix.
End-to-end (e2e) integration tests can be found in the `test/e2e` directory.

The e2e tests use Gont to construct virtual network environment using Linux's `net` namespaces and `veth` point-to-point links.
This allows us to test cunīcu in both simple and complex network topologies including, L2 switches, L3 routers, firewalls and NAT boxes.
