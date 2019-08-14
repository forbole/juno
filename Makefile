VERSION               := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT                := $(shell git log -1 --format='%H')
TOOLS_DESTDIR         ?= $(GOPATH)/bin
GOLANGCI_LINT         = $(TOOLS_DESTDIR)/golangci-lint
GOIMPORTS             = $(TOOLS_DESTDIR)/goimports
GOLANGCI_LINT_VERSION := v1.16.0
GOLANGCI_LINT_HASHSUM := ac897cadc180bf0c1a4bf27776c410debad27205b22856b861d41d39d06509cf

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir 	:= $(dir $(mkfile_path))

all: ci-lint ci-test install

###############################################################################
# Build / Install
###############################################################################

LD_FLAGS = -X github.com/alexanderbez/juno/cmd.Version=$(VERSION) \
	-X github.com/alexanderbez/juno/cmd.Commit=$(COMMIT)

BUILD_FLAGS := -ldflags '$(LD_FLAGS)'

build: go.sum
ifeq ($(OS),Windows_NT)
	@echo "Building juno binary..."
	@go build -mod=readonly $(BUILD_FLAGS) -o build/juno.exe .
else
	@echo "Building juno binary..."
	@go build -mod=readonly $(BUILD_FLAGS) -o build/juno .
endif

install: go.sum
	@echo "Installing juno binary..."
	@go install -mod=readonly $(BUILD_FLAGS) .

###############################################################################
# Tools
###############################################################################

tools-stamp: $(GOLANGCI_LINT) $(GOIMPORTS)
	@touch $@

tools: tools-stamp

$(GOLANGCI_LINT): $(mkfile_dir)contrib/install-golangci-lint.sh
	bash $(mkfile_dir)contrib/install-golangci-lint.sh $(TOOLS_DESTDIR) $(GOLANGCI_LINT_VERSION) $(GOLANGCI_LINT_HASHSUM)

$(STATIK):
	$(call go_install,rakyll,statik,v0.1.5)

$(GOIMPORTS):
	go get golang.org/x/tools/cmd/goimports@v0.0.0-20190114222345-bf090417da8b

###############################################################################
# Tests / CI
###############################################################################

coverage:
	@echo "Viewing test coverage..."
	@go tool cover --html=coverage.out

ci-test:
	@echo "Executing unit tests..."
	@go test -mod=readonly -v -coverprofile coverage.out ./... 

ci-lint: tools
	@echo "Running GolangCI-Lint..."
	@golangci-lint run
	@echo "Formatting..."
	@find . -name '*.go' -type f -not -path "*.git*" | xargs gofmt -d -s
	@echo "Verifying modules..."
	@go mod verify

.PHONY: ci-lint tools tools-stamp coverage
