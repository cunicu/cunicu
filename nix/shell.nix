# SPDX-FileCopyrightText: 2023 Philipp Jungkamp <p.jungkamp@gmx.net>
# SPDX-FileCopyrightText: 2025 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{
  mkShell,

  coturn,
  cunicu,
  cunicu-website,
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

  ...
}:
mkShell {
  packages = [
    coturn
    evans
    ginkgo
    gnumake
    gocov-merger
    golangci-lint
    goreleaser
    inotify-tools
    libpcap
    nix-update
    reuse
    svu
  ];

  inputsFrom = [
    cunicu
    cunicu-website
  ];

  hardeningDisable = [ "fortify" ];
}
