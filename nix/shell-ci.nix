# SPDX-FileCopyrightText: 2023 Philipp Jungkamp <p.jungkamp@gmx.net>
# SPDX-License-Identifier: Apache-2.0
{ pkgs, ... }:
pkgs.mkShell {
  packages = with pkgs; [
    nix-update
    goreleaser
    svu
  ];
}
