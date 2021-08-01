export GO111MODULE=on

TOOLS := $(notdir $(wildcard cmd/*))

OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)
ALL_ARCH := amd64 arm arm64
ALL_OS := linux freebsd openbsd darwin windows

BINS := $(foreach X,$(ALL_OS),$(foreach Y,$(ALL_ARCH),wice-$X-$Y))

temp = $(subst -, ,$@)
cmd = $(word 1, $(temp))
os = $(word 2, $(temp))
arch = $(word 3, $(temp))

all: wice-$(OS)-$(ARCH)

release: $(PLATFORMS)

$(BINS):
	GOOS=$(os) \
	GOARCH=$(arch) \
	go build -o 'build/$(cmd)-$(os)-$(arch)' ./cmd/$(cmd)

.PHONY: release $(PLATFORMS)
