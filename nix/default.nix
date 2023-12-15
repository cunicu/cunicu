# SPDX-FileCopyrightText: 2023 Philipp Jungkamp <p.jungkamp@gmx.net>
# SPDX-License-Identifier: Apache-2.0
{pkgs ? import <nixpkgs> {}}:
pkgs.callPackage ./cunicu.nix {
  src = ./..;
}
