# SPDX-FileCopyrightText: 2023 Philipp Jungkamp <p.jungkamp@gmx.net>
# SPDX-License-Identifier: Apache-2.0
{
  mkShell,

  coturn,
  ginkgo,
  golangci-lint,
  gocov-merger,
  goreleaser,
  nix-update,
  svu,
}:
mkShell {
  packages = [
    coturn
    ginkgo
    gocov-merger
    golangci-lint
    goreleaser
    nix-update
    svu
  ];
}
