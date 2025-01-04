# SPDX-FileCopyrightText: 2025 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{
  lib,
  buildGoModule,
  fetchFromGitHub,
}:
buildGoModule {
  pname = "gocov-merger";
  version = "0.10.0";

  src = fetchFromGitHub {
    owner = "amobe";
    repo = "gocov-merger";
    rev = "5494981677165bdf08c8c0595c3b6ed246cb77de";
    hash = "sha256-zec5gKWbZBAIqlxRS811AwSZxNjmbIsE5/zInp94kR8=";
  };

  vendorHash = "sha256-6DznXSmQkb91GJZ2WMAIg558y+8a46KjRKfWRHsvus0=";

  meta = {
    description = "Merge coverprofile results from multiple go cover runs";
    homepage = "https://github.com/amobe/gocov-merger";
    license = lib.licenses.bsd2;
    maintainers = with lib.maintainers; [ stv0g ];
  };
}
