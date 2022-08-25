PKG = $(shell grep module go.mod | cut -f2 -d" ")

GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_VERSION = $(shell git describe --tags --dirty || echo unknown)

export CGO_ENABLED = 0

LDFLAGS = -X main.version=$(GIT_VERSION) \
          -X main.commit=$(GIT_COMMIT) \
		  -X main.date=$(shell date -Iseconds)

PKGS ?= ./cmd/... ./pkg/...
ifeq ($(GOOS),linux)
    PKGS += ./test/...
endif

GINKGO_PKG ?= ./...
GINKGO_OPTS += --compilers=$(shell nproc) \
			   --keep-going \
			   --timeout=15m \
			   --trace \
			   --cover \
			   --coverpkg=./... \
			   --coverprofile=coverprofile.out \
			   --keep-separate-coverprofiles \
			   --output-dir=./test/logs \
			   --randomize-all \
			   --randomize-suites \
			   $(GINKGO_EXTRA_OPTS) $(GINKGO_PKG) \
			   -- $(GINKGO_TEST_OPTS)

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

vet:
	go vet --copylocks=false $(PKGS)

staticcheck:
	staticcheck $(PKGS)

install-deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/onsi/ginkgo/v2/ginkgo@latest
	go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/amobe/gocov-merger@latest
	go install github.com/jandelgado/gcov2lcov@latest

clean:
	rm -f *.out wice lcov.info

.PHONY: all wice tests tests-watch coverage clean vet staticcheck install-deps