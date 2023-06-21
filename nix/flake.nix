# SPDX-FileCopyrightText: 2023 Philipp Jungkamp <p.jungkamp@gmx.net>
# SPDX-License-Identifier: Apache-2.0
{
  description = "cunīcu is a user-space daemon managing WireGuard® interfaces to establish a mesh of peer-to-peer VPN connections in harsh network environments.";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  outputs = {
    self,
    nixpkgs,
  }: let
    inherit (nixpkgs) lib;
    forSystems = lib.genAttrs ["x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin"];
    pkgsFor = system: nixpkgs.legacyPackages.${system};
    packagesWith = pkgs: {
      cunicu = pkgs.callPackage ./cunicu.nix {
        src = ./..;
      };
    };
  in {
    packages = forSystems (system: packagesWith (pkgsFor system) // {default = self.packages.${system}.cunicu;});
    formatter = forSystems (system: (pkgsFor system).alejandra);
    overlays.default = final: prev: packagesWith final;
  };
}
