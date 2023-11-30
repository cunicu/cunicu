# SPDX-FileCopyrightText: 2023 Philipp Jungkamp <p.jungkamp@gmx.net>
# SPDX-License-Identifier: Apache-2.0
{
  self',
  pkgs,
  ...
}:
pkgs.mkShell {
  packages = with pkgs; [
    yarn-berry
    protobuf
    gnumake
    libpcap
    reuse
    ginkgo
    gcov2lcov
    goreleaser
    golangci-lint
    protoc-gen-go
    protoc-gen-go-grpc

    (buildGoModule
      {
        name = "gocov-merger";

        src = fetchFromGitHub {
          owner = "amobe";
          repo = "gocov-merger";
          rev = "5494981677165bdf08c8c0595c3b6ed246cb77de";
          hash = "sha256-zec5gKWbZBAIqlxRS811AwSZxNjmbIsE5/zInp94kR8=";
        };

        vendorHash = "sha256-6DznXSmQkb91GJZ2WMAIg558y+8a46KjRKfWRHsvus0=";
      })

    # coturn
  ];

  inputsFrom = [
    self'.packages.cunicu
  ];
}
