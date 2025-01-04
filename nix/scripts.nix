# SPDX-FileCopyrightText: 2025 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{ buildGoModule, cunicu }:
buildGoModule {
  pname = "scripts";
  inherit (cunicu) version;

  src = ../scripts;

  subPackages = [ "vanity_redirects" ];

  vendorHash = "sha256-AGepsQRIFVGTWRUQPnSuLEJb/Oxp2G+V0QxBOgb2L1U=";

  meta = {
    mainProgram = "vanity_redirects";
  };
}
