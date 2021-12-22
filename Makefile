export GO111MODULE=on

TOOLS := $(notdir $(wildcard cmd/*))

OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)
ALL_ARCH := amd64 arm arm64
ALL_OS := linux freebsd openbsd darwin windows

BINS := $(foreach X,$(ALL_OS),$(foreach Y,$(ALL_ARCH),wice-$X-$Y))

PROTOBUFS := pkg/pb/socket.pb.go \
			 pkg/pb/offer.pb.go \
			 pkg/pb/common.pb.go \
			 pkg/pb/event.pb.go \
			 pkg/pb/socket_grpc.pb.go

temp = $(subst -, ,$@)
cmd = $(word 1, $(temp))
os = $(word 2, $(temp))
arch = $(word 3, $(temp))

all: wice-$(OS)-$(ARCH)

release: $(PLATFORMS)

$(BINS): $(PROTOBUFS)
	GOOS=$(os) \
	GOARCH=$(arch) \
	go build -o 'build/$(cmd)-$(os)-$(arch)' ./cmd/$(cmd)

%.pb.go: %.proto
	protoc \
		--proto_path=$(dir $^) \
		--go_out=$(dir $^) \
		--go_opt=paths=source_relative $^

%_grpc.pb.go: %.proto
	protoc \
		--proto_path=$(dir $^) \
		--go-grpc_out=$(dir $^) \
		--go-grpc_opt=paths=source_relative  $^

.PHONY: release $(PLATFORMS) proto
