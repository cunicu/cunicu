FROM golang:1.19-alpine AS builder

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

FROM alpine:3.16

COPY --from=builder /app/cunicu /

ENTRYPOINT ["/cunicu"]
