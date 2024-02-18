# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.22-alpine AS builder

RUN apk add \
    git \
    make \
    protoc

COPY Makefile .
RUN make install-deps

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
RUN make

FROM alpine:3.19

COPY --from=builder /app/cunicu /

ENTRYPOINT ["/cunicu"]
