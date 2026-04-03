BINARY_NAME := ailint
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"
GOFLAGS := -trimpath

.PHONY: build test lint clean install release fmt vet coverage

## build: Build the ailint binary
build:
	go build $(GOFLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/ailint

## install: Install ailint to $GOPATH/bin
install:
	go install $(GOFLAGS) $(LDFLAGS) ./cmd/ailint

## test: Run all tests
test:
	go test -race -count=1 ./...

## coverage: Run tests with coverage report
coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: Format code
fmt:
	gofmt -s -w .
	goimports -w .

## vet: Run go vet
vet:
	go vet ./...

## clean: Remove build artifacts
clean:
	rm -rf bin/ coverage.out dist/

## release: Create a release with goreleaser
release:
	goreleaser release --clean

## snapshot: Create a snapshot release (no publish)
snapshot:
	goreleaser release --snapshot --clean

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /'
