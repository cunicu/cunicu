PKG = $(shell grep module go.mod | cut -f2 -d" ")

GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_VERSION = $(shell git describe --tags --dirty || echo unknown)

LDFLAGS = -X main.version=$(GIT_VERSION) -X main.commit=$(GIT_COMMIT) -X main.date=$(shell date -Iseconds)

GINKGO_PKG ?= ./...
GINKGO_OPTS += -v -cover -coverpkg=$(PKG)/... -covermode=count -r -p --randomize-all $(GINKGO_PKG)

all: wice

wice:
	go generate ./...
	go build -o $@ -ldflags="$(LDFLAGS)" ./cmd/wice

tests:
	ginkgo run $(GINKGO_OPTS)

lcov.info: $(wildcard *.out) 
	@echo "Merging and converting coverage data..."
	gocovmerge $(wildcard *.out) | gcov2lcov > $@
	@echo "Done. $@ updated"

coverage: lcov.info

tests-watch:
	( while inotifywait -qqe close_write --include "\.out$$" .; do $(MAKE) -sB coverage; done & )
	ginkgo watch $(GINKGO_OPTS)

clean:
	rm -f *.out wice lcov.info

.PHONY: all wice tests tests-watch coverage clean