# SPDX-FileCopyrightText: 2023 Philipp Jungkamp <p.jungkamp@gmx.net>
# SPDX-License-Identifier: Apache-2.0
{
  lib,
  src,
  buildGoModule,
}:
buildGoModule rec {
  pname = "cunicu";
  version = "0.3.4";
  vendorHash = "sha256-KXLKCwV3/9+PRQ1mYCZBvaV914WN68FgJXNWZizOmkA=";
  inherit src;
  CGO_ENABLED = 0;
  # these packages contain networking dependent tests which fail in the sandbox
  excludedPackages = ["pkg/config" "pkg/selfupdate" "pkg/tty"];
  postBuild = ''
    cunicu=$GOPATH/bin/cunicu
    $cunicu docs --with-frontmatter
  '';
  postInstall = ''
    install -d $out/usr/share/man/man1
    install ./docs/usage/man/*.1 $out/usr/share/man/man1
    install -D <($cunicu completion bash) $out/share/bash-completion/completions/cunicu
    install -D <($cunicu completion fish) $out/share/fish/vendor_completions.d/cunicu.fish
    install -D <($cunicu completion zsh) $out/share/zsh/vendor-completions/_cunicu
  '';
  meta = with lib; {
    description = "A zeroconf peer-to-peer mesh VPN using Wireguard® and Interactive Connectivity Establishment (ICE)";
    homepage = "https://cunicu.li";
    license = licenses.asl20;
  };
}
