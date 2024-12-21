# SPDX-FileCopyrightText: 2023 Philipp Jungkamp <p.jungkamp@gmx.net>
# SPDX-License-Identifier: Apache-2.0
{
  description = "cunīcu is a user-space daemon managing WireGuard® interfaces to establish a mesh of peer-to-peer VPN connections in harsh network environments.";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs =
    inputs@{
      nixpkgs,
      self,
      flake-parts,
    }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      flake = {
        nixosModules = rec {
          default = cunicu;
          cunicu = import ./nix/module.nix;
        };

        overlays = {
          default = final: prev: { cunicu = import ./nix/default.nix { pkgs = final; }; };
        };
      };

      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];

      perSystem =
        {
          pkgs,
          self',
          system,
          ...
        }:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [ self.overlays.default ];
          };
        in
        {
          formatter = pkgs.nixfmt-rfc-style;

          devShells.default = import ./nix/dev.nix { inherit pkgs self'; };

          packages = {
            cunicu = import ./nix/default.nix { inherit pkgs; };
            default = self'.packages.cunicu;
          };
        };
    };
}
