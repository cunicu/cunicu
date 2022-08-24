PKG = $(shell grep module go.mod | cut -f2 -d" ")

GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_VERSION = $(shell git describe --tags --dirty || echo unknown)

LDFLAGS = -X main.version=$(GIT_VERSION) -X main.commit=$(GIT_COMMIT) -X main.date=$(shell date -Iseconds)

GINKGO_PKG ?= ./...
GINKGO_OPTS += --covermode=count --coverpkg=./... --coverprofile=coverprofile.out --randomize-all $(GINKGO_EXTRA_OPTS) $(GINKGO_PKG) -- $(GINKGO_TEST_OPTS)

all: wice

wice:
	go generate ./...
	go build -o $@ -ldflags="$(LDFLAGS)" ./cmd/wice

tests:
	ginkgo run $(GINKGO_OPTS)

coverprofile_merged.out: $(shell find . -name "*.out" -type f)
	gocov-merger -o $@ $^

lcov_merged.info: coverprofile_merged.out
	gcov2lcov > $@ < $^ 

coverage: lcov.info

tests-watch:
	( while inotifywait -qqe close_write --include "\.out$$" .; do $(MAKE) -sB coverage; done & )
	ginkgo watch $(GINKGO_OPTS)

clean:
	rm -f *.out wice lcov.info

.PHONY: all wice tests tests-watch coverage clean