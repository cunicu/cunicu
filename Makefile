GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_VERSION = $(shell git describe --tags --dirty HEAD || echo unknown)

LDFLAGS = -X main.version=$(GIT_VERSION) -X main.commit=$(GIT_COMMIT) -X main.date=$(shell date -Iseconds)

all: wice

wice:
	go generate ./...
	go build -o $@ -ldflags="$(LDFLAGS)" ./cmd

.PHONY: all wice