VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT  := $(shell git log -1 --format='%H')

export GO111MODULE = on

all: ci-lint ci-test install

###############################################################################
# Build / Install
###############################################################################

LD_FLAGS = -X github.com/angelorc/desmos-parser/version.Version=$(VERSION) \
	-X github.com/angelorc/desmos-parser/version.Commit=$(COMMIT)

BUILD_FLAGS := -ldflags '$(LD_FLAGS)'

build: go.sum
ifeq ($(OS),Windows_NT)
	@echo "building juno binary..."
	@go build -mod=readonly $(BUILD_FLAGS) -o build/juno.exe ./cmd/juno
else
	@echo "building juno binary..."
	@go build -mod=readonly $(BUILD_FLAGS) -o build/juno ./cmd/juno
endif

install: go.sum
	@echo "installing juno binary..."
	@go install -mod=readonly $(BUILD_FLAGS) ./cmd/juno

###############################################################################
# Tests / CI
###############################################################################

coverage:
	@echo "viewing test coverage..."
	@go tool cover --html=coverage.out

ci-test:
	@echo "executing unit tests..."
	@go test -mod=readonly -v -coverprofile coverage.txt ./...

ci-lint:
	@echo "running GolangCI-Lint..."
	@GO111MODULE=on golangci-lint run
	@echo "formatting..."
	@find . -name '*.go' -type f -not -path "*.git*" | xargs gofmt -d -s
	@echo "verifying modules..."
	@go mod verify

clean:
	rm -f tools-stamp ./build/**

.PHONY: install build ci-test ci-lint coverage clean
