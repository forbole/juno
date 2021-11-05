VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT  := $(shell git log -1 --format='%H')

export GO111MODULE = on

all: lint test-unit install

###############################################################################
# Build / Install
###############################################################################

LD_FLAGS = -X github.com/forbole/juno/v2/cmd.Version=$(VERSION) \
	-X github.com/forbole/juno/v2/cmd.Commit=$(COMMIT)

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

test-unit:
	@echo "executing unit tests..."
	@go test -mod=readonly -v -coverprofile coverage.txt ./...

lint:
	golangci-lint run --out-format=tab

lint-fix:
	golangci-lint run --fix --out-format=tab --issues-exit-code=0
.PHONY: lint lint-fix


format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' | xargs misspell -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' | xargs goimports -w -local github.com/forbole/juno
.PHONY: format

clean:
	rm -f tools-stamp ./build/**

.PHONY: install build ci-test ci-lint coverage clean
