PKG = $(shell grep module go.mod | cut -f2 -d" ")

export CGO_ENABLED = 0

LDFLAGS = -X github.com/stv0g/cunicu/pkg/util/buildinfo.Version=$(shell git describe --tags --dirty || echo unknown) \
		  -X github.com/stv0g/cunicu/pkg/util/buildinfo.Tag=$(shell git describe --tags) \
          -X github.com/stv0g/cunicu/pkg/util/buildinfo.Commit=$(shell git rev-parse HEAD) \
		  -X github.com/stv0g/cunicu/pkg/util/buildinfo.Branch=$(shell git rev-parse --abbrev-ref HEAD) \
		  -X github.com/stv0g/cunicu/pkg/util/buildinfo.DateStr=$(shell date -Iseconds) \
		  -X github.com/stv0g/cunicu/pkg/util/buildinfo.BuiltBy=makefile \

PKGS ?= ./cmd/... ./pkg/... ./test
ifeq ($(GOOS),linux)
    PKGS += ./test/e2e/...
endif

GINKGO_OPTS =  --compilers=2 \
			   --keep-going \
			   --timeout=15m \
			   --trace \
			   --cover \
			   --coverpkg=./... \
			   --keep-separate-coverprofiles \
			   --randomize-all \
			   --randomize-suites \
			   $(GINKGO_EXTRA_OPTS)

all: cunicu

cunicu:
	go generate ./...
	go build -o $@ -ldflags="$(LDFLAGS)" ./cmd/cunicu

tests:
	ginkgo run $(GINKGO_OPTS) --coverprofile=coverprofile.out ./pkg/...

tests-e2e:
	ginkgo run $(GINKGO_OPTS) --output-dir=./test/e2e/logs --coverprofile=coverprofile_e2e.out ./test/e2e

coverprofile_merged.out: $(shell find . -name "*.out" -type f)
	gocov-merger -o $@ $^

lcov.info: coverprofile_merged.out
	gcov2lcov > $@ < $^ 

coverage: lcov.info

tests-watch:
	( while inotifywait -qqe close_write --include "\.out$$" .; do $(MAKE) -sB coverage; done & )
	ginkgo watch $(GINKGO_OPTS)

tidy:
	go mod tidy

generate:
	go generate ./...

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
	go install github.com/goreleaser/goreleaser@latest

website: docs
	cd website && \
	yarn build

docs: $(wildcard cmd/cunicu/*.go)
	rm -rf ./docs/usage/{man,md}
	go run ./cmd/cunicu/ docs --with-frontmatter
	# find ./docs/usage/md -name "*.md" ! -name "cunicu_completion_*.md" -exec sed -i 's/</\\</g;s/>/\\>/g;' {} \;

completions: completions/cunicu.bash completions/cunicu.zsh completions/cunicu.fish

completions-dir:
	mkdir completions

completions/cunicu.%: completions-dir
	go run ./cmd/cunicu/ completion $* > $@

prepare: clean tidy generate vet staticcheck docs completions

ci: install-deps vet staticcheck tests

clean:
	find . -name "*.out" -exec rm {} \;
	rm -rf cunicu lcov.info test/logs/ completions/

.PHONY: all cunicu tests tests-watch coverage clean vet staticcheck install-deps ci completions docs prepare generate website