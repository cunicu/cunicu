# SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0

PKG = $(shell grep module go.mod | cut -f2 -d" ")

export CGO_ENABLED = 0

LINT_OPTS ?= --timeout=15m

LDFLAGS = -X cunicu.li/cunicu/pkg/buildinfo.Version=$(shell git describe --tags --dirty || echo unknown) \
          -X cunicu.li/cunicu/pkg/buildinfo.Tag=$(shell git describe --tags) \
          -X cunicu.li/cunicu/pkg/buildinfo.Commit=$(shell git rev-parse HEAD) \
          -X cunicu.li/cunicu/pkg/buildinfo.Branch=$(shell git rev-parse --abbrev-ref HEAD) \
          -X cunicu.li/cunicu/pkg/buildinfo.DateStr=$(shell date -Iseconds) \
          -X cunicu.li/cunicu/pkg/buildinfo.BuiltBy=makefile \

PKGS ?= ./cmd/... ./pkg/... ./test
ifeq ($(GOOS),linux)
    PKGS += ./test/e2e/...
endif

ifeq ($(CI),true)
	GINKGO_OPTS += \
			   --keep-going \
			   --timeout=15m \
			   --trace \
			   --cover \
			   --coverpkg=./... \
			   --keep-separate-coverprofiles \
			   --randomize-all \
			   --randomize-suites
endif


all: cunicu

cunicu:
	go generate ./...
	go build -o $@ -ldflags="$(LDFLAGS)" ./cmd/cunicu

tests:
	ginkgo run $(GINKGO_OPTS) --coverprofile=coverprofile.out ./pkg/... -- $(GINKGO_ARGS)

tests-e2e:
	ginkgo run $(GINKGO_OPTS) --output-dir=./test/e2e/logs --coverprofile=coverprofile_e2e.out ./test/e2e -- $(GINKGO_ARGS)

coverprofile_merged.out: $(shell find . -type f -name "*.out" -and -not -name "*_merged.out")
	gocov-merger -o $@ $^

coverage: coverprofile_merged.out

tests-watch:
	( while inotifywait -qqe close_write --include "\.out$$" .; do $(MAKE) -sB coverage; done & )
	ginkgo watch $(GINKGO_OPTS)

tidy:
	go mod tidy

generate:
	go generate ./...

lint:
	golangci-lint run $(LINT_OPTS) $(PKGS)

install-deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/amobe/gocov-merger@latest
	go install github.com/goreleaser/goreleaser@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/onsi/ginkgo/v2/ginkgo

completions: completions/cunicu.bash completions/cunicu.zsh completions/cunicu.fish

completions-dir:
	mkdir completions

completions/cunicu.%: completions-dir
	go run ./cmd/cunicu completion $* > $@

docs:
	go run ./cmd/cunicu docs --with-frontmatter --output-dir docs/usage

prepare: clean tidy generate lint completions docs

ci:
	CLICOLOR_FORCE=1 \
	act push $(ACT_OPTS) \
		--remote-name $(shell git config branch.main.remote) \
		--platform ubuntu-24.04=catthehacker/ubuntu:act-latest \
	| grep -v '::' # Suppress debug output

clean:
	find . -name "*.out" -exec rm {} \;
	rm -rf cunicu test/logs/ completions/

.PHONY: all cunicu tests tests-watch coverage clean lint install-deps completions prepare generate ci tidy docs
