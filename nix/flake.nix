# SPDX-FileCopyrightText: 2023 Philipp Jungkamp <p.jungkamp@gmx.net>
# SPDX-License-Identifier: Apache-2.0
{
  description = "cunīcu is a user-space daemon managing WireGuard® interfaces to establish a mesh of peer-to-peer VPN connections in harsh network environments.";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs = inputs @ {
    self,
    nixpkgs,
    flake-parts,
  }:
    flake-parts.lib.mkFlake {inherit inputs;} {
      flake = {
        # Put your original flake attributes here.
      };
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];
      perSystem = {
        config,
        pkgs,
        self',
        ...
      }: {
        formatter = pkgs.alejandra;
        devShells.default = import ./dev.nix {
          inherit self pkgs self';
        };
        packages = rec {
          cunicu = import ./default.nix {
            inherit pkgs;
          };
          default = cunicu;
        };
      };
    };
}
