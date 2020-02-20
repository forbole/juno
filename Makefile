VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT  := $(shell git log -1 --format='%H')

all: install

build: go.sum
ifeq ($(OS),Windows_NT)
	@echo "building desmosp binary..."
	@go build -mod=readonly $(BUILD_FLAGS) -o build/desmosp.exe .
else
	@echo "building desmosp binary..."
	@go build -mod=readonly $(BUILD_FLAGS) -o build/desmosp .
endif

install: go.sum
	@echo "installing desmosp binary..."
	@go install -mod=readonly $(BUILD_FLAGS) .

.PHONY: install build