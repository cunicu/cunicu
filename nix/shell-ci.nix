# SPDX-FileCopyrightText: 2023 Philipp Jungkamp <p.jungkamp@gmx.net>
# SPDX-License-Identifier: Apache-2.0
{
  mkShell,

  golangci-lint,
  goreleaser,
  nix-update,
  svu,
}:
mkShell {
  packages = [
    golangci-lint
    goreleaser
    nix-update
    svu
  ];
}
