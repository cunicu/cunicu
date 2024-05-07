#!/usr/bin/env bash

# SPDX-FileCopyrightText: 2023 Philipp Jungkamp <p.jungkamp@gmx.net>
# SPDX-License-Identifier: Apache-2.0

FILE="${1:-./nix/cunicu.nix}"
VERSION="${2:-}"

if [ -z "$VERSION" ]; then
    echo "Using latest version based on git tags"
    VERSION="$(git tag --merged | sort | grep 'v.*' | tail -n1)"
fi

if [ -z "$VERSION" ]; then
    echo "Could not determine version"
    exit 1
fi

echo "Set version to ${VERSION}"
sed -i "s|version.*|version = \"${VERSION}\";|" "${FILE}"

echo 'Faking go vendorHash'
sed -i 's|vendorHash.*;$|vendorHash = lib.fakeHash;|' "${FILE}"

echo "Calculate go modules hash"
OUTPUT="$(nix build ./nix#cunicu.goModules \
    --extra-experimental-features 'nix-command flakes' \
    --refresh \
    --no-link \
    2>&1)"
CORRECT_HASH="$(sed -n 's|^\s*got:\s*||p' <<<"${OUTPUT}")"
if [ -z "${CORRECT_HASH}" ]; then
	echo -e "Error!\n${OUTPUT}"
	exit 1
fi

echo "Set vendorHash to ${CORRECT_HASH}"
sed -i "s|vendorHash.*|vendorHash = \"${CORRECT_HASH}\";|" "${FILE}"

echo "Check goModules derivation"
nix build ./nix#cunicu.goModules \
  --extra-experimental-features 'nix-command flakes' \
  --no-link
