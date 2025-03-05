# SPDX-FileCopyrightText: 2025 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0
{
  lib,
  stdenv,
  stdenvNoCC,

  yarn-berry,
  cacert,
  nodejs-slim,
  cunicu,
  cunicu-scripts,

  # Options
  goModules ? builtins.fromJSON (builtins.readFile ./modules.json),

  ...
}:
let
  yarnOfflineCache = stdenvNoCC.mkDerivation {
    name = "yarn-deps";
    inherit (cunicu) version src;

    nativeBuildInputs = [ yarn-berry ];

    dontInstall = true;

    NODE_EXTRA_CA_CERTS = "${cacert}/etc/ssl/certs/ca-bundle.crt";
    YARN_ENABLE_TELEMETRY = "0";

    buildPhase = ''
      runHook preBuild

      cd website

      export HOME=$(mktemp -d)

      YARN_CACHE="$(yarn config get cacheFolder)"
      yarn install --immutable --mode skip-build

      mkdir -p $out
      cp -r $YARN_CACHE/* $out/

      runHook postBuild
    '';

    outputHash = "sha256-akMIrajGv6mZJecQaLgu1IjxvYpfA51SqryRxa5m58U";
    outputHashMode = "recursive";
  };
in
stdenv.mkDerivation (finalAttrs: {
  name = "cunicu-docs";
  inherit (cunicu) version src;
  inherit yarnOfflineCache;

  nativeBuildInputs = [
    nodejs-slim
    yarn-berry
  ];

  NODE_ENV = "production";
  YARN_ENABLE_TELEMETRY = "0";

  buildPhase = ''
    runHook preBuild

    shopt -s globstar

    cd website

    export HOME=$(mktemp -d)
    export npm_config_nodedir=${nodejs-slim}

    mkdir -p ~/.yarn/berry
    ln -s $yarnOfflineCache ~/.yarn/berry/cache

    echo "== Generate redirects"
    ${cunicu-scripts}/bin/vanity_redirects -static-dir static -modules-file ${builtins.toFile "modules.json" (builtins.toJSON goModules)}

    echo "== Generate usage docs"
    ${lib.getExe cunicu} docs --with-frontmatter --output-dir docs/usage

    echo "-- Fix generated docs"
    substituteInPlace docs/usage/md/**/*.md \
      --replace-quiet '<' '\<' \
      --replace-quiet '{' '\{'

    echo "== Yarn install"
    yarn install --immutable --immutable-cache

    patchShebangs ~/node_modules

    echo "== Yarn build"
    yarn build

    runHook postBuild
  '';

  installPhase = ''
    runHook preInstall

    cp -r build $out

    runHook postInstall
  '';
})
