#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2023 Philipp Jungkamp <p.jungkamp@gmx.net>
# SPDX-License-Identifier: Apache-2.0

file=./nix/cunicu.nix

printf '%s\n' 'Faking the hash'

sed -i 's|vendorHash.*;$|vendorHash = lib.fakeHash;|' "$file"

printf '%s\n' 'Evaluating the derivation'

output="$(
  nix build ./nix#cunicu.go-modules \
    --extra-experimental-features 'nix-command flakes' \
    --refresh \
    --no-link \
    2>&1
)"

printf '%s\n' 'Extract correct hash'

correct_hash="$(sed -n '$s|^\s*got:\s*||p' <<<"$output")"

if [ -z "$correct_hash" ]; then
	printf '%s\n' 'Error!' "$output"
	exit 1
fi

printf '%s\n' "Set hash to $correct_hash"

sed -i "s|vendorHash.*;$|vendorHash = \"$correct_hash\";|" "$file"

nix build ./nix#cunicu \
  --extra-experimental-features 'nix-command flakes' \
  --no-link

