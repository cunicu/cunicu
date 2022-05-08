GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_VERSION = $(shell git describe --tags --dirty || echo unknown)

LDFLAGS = -X main.version=$(GIT_VERSION) -X main.commit=$(GIT_COMMIT) -X main.date=$(shell date -Iseconds)

GINKGO_OPTS = -cover -r -p --randomize-all

all: wice

wice:
	go generate ./...
	go build -o $@ -ldflags="$(LDFLAGS)" ./cmd/wice

tests:
	ginkgo run $(GINKGO_OPTS)

lcov.info: coverprofile.out
	gcov2lcov -infile $^ -outfile $@

coverprofile.out: tests

coverage: lcov.info

tests-watch:
	( while inotifywait -qqe close_write coverprofile.out; do $(MAKE) -s lcov.info; done & )
	ginkgo watch $(GINKGO_OPTS)

.PHONY: all wice tests tests-watch coverage