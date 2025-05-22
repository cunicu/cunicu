# SPDX-FileCopyrightText: 2023 Philipp Jungkamp <p.jungkamp@gmx.net>
# SPDX-License-Identifier: Apache-2.0
{
  description = "cunīcu is a user-space daemon managing WireGuard® interfaces to establish a mesh of peer-to-peer VPN connections in harsh network environments.";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    nix-update = {
      url = "github:Mic92/nix-update";
      inputs = {
        nixpkgs.follows = "nixpkgs";
        flake-parts.follows = "flake-parts";
      };
    };
  };

  outputs =
    inputs@{ self, ... }:
    inputs.flake-parts.lib.mkFlake { inherit inputs; } {
      flake = {
        nixosModules = {
          default = self.nixosModules.cunicu;
          cunicu = import ./nix/module.nix;
        };

        overlays = {
          default = final: prev: {
            cunicu = final.callPackage ./nix/cunicu.nix { src = self; };
            cunicu-website = final.callPackage ./nix/website.nix { };
            cunicu-scripts = final.callPackage ./nix/scripts.nix { };
            gocov-merger = final.callPackage ./nix/gocov-merger.nix { };
          };
        };
      };

      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];

      perSystem =
        {
          system,
          ...
        }:
        let
          pkgs = import inputs.nixpkgs {
            inherit system;
            overlays = [ self.overlays.default ];
          };
          lib = inputs.nixpkgs.lib;
        in
        {
          formatter = pkgs.nixfmt-rfc-style;

          devShells =
            let
              inherit (inputs.nix-update.packages.${system}) nix-update;
            in
            {
              default = pkgs.callPackage ./nix/shell.nix { inherit nix-update; };
              ci = pkgs.callPackage ./nix/shell-ci.nix { inherit nix-update; };
            };

          packages = {
            inherit (pkgs) cunicu gocov-merger;

            default = pkgs.cunicu;
            website = pkgs.cunicu-website;
            scripts = pkgs.cunicu-scripts;
            nixosTest = import ./nix/test.nix { inherit self pkgs lib; };
          };
        };
    };
}
