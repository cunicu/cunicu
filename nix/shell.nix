# SPDX-FileCopyrightText: 2023 Philipp Jungkamp <p.jungkamp@gmx.net>
# SPDX-FileCopyrightText: 2025 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{
  lib,
  stdenv,
  mkShell,

  act,
  coturn,
  cunicu-website,
  cunicu,
  evans,
  ginkgo,
  gnumake,
  gocov-merger,
  golangci-lint,
  goreleaser,
  inotify-tools,
  libpcap,
  nix-update,
  reuse,
  svu,

  cunicu-scripts,

  ...
}:
mkShell {
  packages =
    [
      act
      evans
      ginkgo
      gnumake
      gocov-merger
      golangci-lint
      goreleaser
      libpcap
      nix-update
      reuse
      svu

      cunicu-scripts
    ]
    ++ lib.optional stdenv.isLinux [
      coturn
      inotify-tools
    ];

  inputsFrom = [
    cunicu
    cunicu-website
  ];

  hardeningDisable = [ "fortify" ];
}
