# SPDX-FileCopyrightText: 2025 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{ buildGo124Module, cunicu }:
buildGo124Module {
  pname = "scripts";
  inherit (cunicu) version;

  src = ../scripts;

  subPackages = [ "vanity_redirects" ];

  vendorHash = "sha256-SawrbxYWFN03h/ZV1XiOZ4kvKcCF4G0QJY1nMxgwAzI=";

  meta = {
    mainProgram = "vanity_redirects";
  };
}
