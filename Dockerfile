# SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0

FROM nixos/nix:2.29.1 AS builder

WORKDIR /src

COPY . .
RUN nix \
    --extra-experimental-features "nix-command flakes" \
    build

FROM alpine:3.22

COPY --from=builder /src/result/bin/cunicu /

ENTRYPOINT ["/cunicu"]
