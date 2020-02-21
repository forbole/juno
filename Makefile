VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT  := $(shell git log -1 --format='%H')

export GO111MODULE = on

all: install

LD_FLAGS = -X github.com/angelorc/desmos-parser/version.Version=$(VERSION) \
	-X github.com/angelorc/desmos-parser/version.Commit=$(COMMIT)

BUILD_FLAGS := -ldflags '$(LD_FLAGS)'

########################################
### Build

build: go.sum
ifeq ($(OS),Windows_NT)
	@echo "--> Building the parser binaries..."
	@go build -mod=readonly $(BUILD_FLAGS) -o build/desmosp.exe ./cmd/desmosp
else
	@echo "--> Building the parser binaries..."
	@go build -mod=readonly $(BUILD_FLAGS) -o build/desmosp ./cmd/desmosp
endif

########################################
### Install

install: go.sum
	@echo "--> Installing the parser binaries..."
	@go install -mod=readonly $(BUILD_FLAGS) ./cmd/desmosp

########################################
### Tools & dependencies

go.sum: go.mod
	@echo "--> Ensuring the dependencies have not been modified"
	@go mod verify

.PHONY: install build go.sum
