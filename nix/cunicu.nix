# SPDX-FileCopyrightText: 2023 Philipp Jungkamp <p.jungkamp@gmx.net>
# SPDX-FileCopyrightText: 2025 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{
  lib,
  stdenv,
  buildGo124Module,
  installShellFiles,
  freebsd,
  versionCheckHook,
  protobuf,
  protoc-gen-go,
  protoc-gen-go-grpc,
  nix-update-script,
}:
let
  version = "0.13.0";
  src = ./..;
in
buildGo124Module {
  pname = "cunicu";
  inherit version src;

  vendorHash = "sha256-8MqGcIkYJdWtLsV1/dNaMvxmGf9/qoZmHmBjtBtvpfE=";

  nativeBuildInputs = [
    installShellFiles

    protobuf
    protoc-gen-go
    protoc-gen-go-grpc
  ];

  buildInputs = lib.optionals stdenv.isFreeBSD [
    freebsd.libnv
  ];

  nativeInstallCheckInputs = [ versionCheckHook ];

  env.CGO_ENABLED = if stdenv.isFreeBSD then "1" else "0";

  # These packages contain networking dependent tests which fail in the sandbox
  excludedPackages =
    [
      "pkg/config"
      "pkg/selfupdate"
      "pkg/tty"
      "scripts"
    ]
    ++ lib.optionals (!stdenv.isLinux) [
      "test/e2e"
    ];

  ldflags = [
    "-X cunicu.li/cunicu/pkg/buildinfo.Version=${version}"
    "-X cunicu.li/cunicu/pkg/buildinfo.BuiltBy=Nix"
  ];

  doInstallCheck = true;
  versionCheckProgramArg = "version";

  passthru.updateScript = nix-update-script { };

  preBuild = ''
    go generate ./...
  '';

  postInstall = lib.optionalString (stdenv.buildPlatform.canExecute stdenv.hostPlatform) ''
    $out/bin/cunicu docs --with-frontmatter
    installManPage ./docs/usage/man/*.1
    installShellCompletion --cmd cunicu \
      --bash <($out/bin/cunicu completion bash) \
      --zsh <($out/bin/cunicu completion zsh) \
      --fish <($out/bin/cunicu completion fish)
  '';

  meta = {
    description = "Zeroconf peer-to-peer mesh VPN using Wireguard and Interactive Connectivity Establishment (ICE)";
    homepage = "https://cunicu.li";
    license = lib.licenses.asl20;
    platforms = lib.platforms.linux ++ lib.platforms.darwin ++ lib.platforms.freebsd;
    maintainers = [ lib.maintainers.stv0g ];
    mainProgram = "cunicu";
  };
}
