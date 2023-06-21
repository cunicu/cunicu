#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2023 Philipp Jungkamp <p.jungkamp@gmx.net>
# SPDX-License-Identifier: Apache-2.0

FILE=./nix/cunicu.nix

echo 'Faking the hash'

sed -i 's|vendorHash.*;$|vendorHash = lib.fakeHash;|' "${FILE}"

echo "Evaluating the derivation"

OUTPUT="$(nix build ./nix#cunicu.go-modules \
    --extra-experimental-features 'nix-command flakes' \
    --refresh \
    --no-link \
    2>&1)"


echo "Extract correct hash"

CORRECT_HASH=$(sed -n 's|^\s*got:\s*||p' <<<"${OUTPUT}")

if [ -z "${CORRECT_HASH}" ]; then
	echo -e "Error!\n${OUTPUT}"
	exit 1
fi

echo "Set hash to ${CORRECT_HASH}"

sed -i "s|vendorHash.*|vendorHash = \"${CORRECT_HASH}\";|" "${FILE}"

nix build ./nix#cunicu \
  --extra-experimental-features 'nix-command flakes' \
  --no-link

