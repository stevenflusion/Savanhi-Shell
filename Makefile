# Savanhi Shell Makefile
# Build, test, and development targets

# Binary name
BINARY_NAME=savanhi-shell
BINARY_UNIX=$(BINARY_NAME)_unix

# Build directory
BUILD_DIR=./bin
DIST_DIR=./dist

# Main package
MAIN_PACKAGE=./cmd/savanhi-shell

# Version info (set at build time)
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildDate=$(BUILD_DATE)"

# Platform targets
PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64

.PHONY: all build clean test coverage lint fmt help run install release cross-compile e2e

all: clean deps build

## build: Build the binary for current platform
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

## build-all: Build for all platforms (deprecated, use cross-compile)
build-all: cross-compile

## cross-compile: Build for all supported platforms
cross-compile:
	@echo "Cross-compiling for all platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		os=`echo $$platform | cut -d/ -f1`; \
		arch=`echo $$platform | cut -d/ -f2`; \
		output="$(DIST_DIR)/$(BINARY_NAME)-$$os-$$arch"; \
		if [ "$$os" = "windows" ]; then \
			output="$$output.exe"; \
		fi; \
		echo "Building for $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch $(GOBUILD) $(LDFLAGS) -o $$output $(MAIN_PACKAGE); \
	done
	@echo "Cross-compilation complete. Binaries in $(DIST_DIR)/"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) verify

## test: Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -count=1 ./...

## test-short: Run short tests (skip integration)
test-short:
	@echo "Running short tests..."
	$(GOTEST) -v -short -race -count=1 ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## coverage: Alias for test-coverage
coverage: test-coverage

## e2e: Run end-to-end tests
e2e:
	@echo "Running E2E tests..."
	$(GOTEST) -v -count=1 ./tests/e2e/...

## lint: Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD)fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

## run: Build and run locally
run: build
	@echo "Running..."
	$(BUILD_DIR)/$(BINARY_NAME)

## install: Install binary to GOPATH/bin
install: build
	@echo "Installing..."
	$(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) $(MAIN_PACKAGE)

## install-local: Install binary to /usr/local/bin (requires sudo)
install-local: build
	@echo "Installing to /usr/local/bin..."
	sudo install -m 755 $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

## uninstall: Remove binary from /usr/local/bin
uninstall:
	@echo "Uninstalling..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)

## check: Run linter and tests
check: lint test

## ci: Run CI checks
ci: deps fmt vet lint test test-coverage

## release: Create release using goreleaser (requires goreleaser)
release:
	@echo "Creating release..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --rm-dist; \
	else \
		echo "goreleaser not installed. Run: go install github.com/goreleaser/goreleaser/v2@latest"; \
		exit 1; \
	fi

## release-snapshot: Create snapshot release (for testing)
release-snapshot:
	@echo "Creating snapshot release..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --snapshot --rm-dist; \
	else \
		echo "goreleaser not installed. Run: go install github.com/goreleaser/goreleaser/v2@latest"; \
		exit 1; \
	fi

## checksums: Generate checksums for release binaries
checksums:
	@echo "Generating checksums..."
	cd $(DIST_DIR) && sha256sum $(BINARY_NAME)* > checksums.txt

## version: Print version info
version:
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"

## help: Show this help message
help:
	@echo "Savanhi Shell - Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'

## docker-test: Run E2E tests in Docker
docker-test:
	@echo "Running E2E tests in Docker..."
	@for dockerfile in tests/e2e/docker/Dockerfile.*; do \
		echo "Testing with $$dockerfile..."; \
		docker build -t savanhi-test-$$(@F) -f $$dockerfile . || exit 1; \
	done

## docker-build: Build in Docker
docker-build:
	@echo "Building in Docker..."
	docker build -t savanhi-shell:$(VERSION) .

.DEFAULT_GOAL := build