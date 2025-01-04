# SPDX-FileCopyrightText: 2023 Philipp Jungkamp <p.jungkamp@gmx.net>
# SPDX-License-Identifier: Apache-2.0
{
  mkShell,

  nix-update,
  goreleaser,
  svu,
}:
mkShell {
  packages = [
    nix-update
    goreleaser
    svu
  ];
}
