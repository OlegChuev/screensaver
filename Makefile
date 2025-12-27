# Binary name
BINARY_NAME=screensaver

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build directory
BUILD_DIR=build

# Version info
VERSION?=0.1.0
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION)"

.PHONY: all build build-all clean run test deps install uninstall

# Default target
all: clean build

# Build for current platform
build:
	@echo "Building $(BINARY_NAME)..."
	@$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

# Build for all platforms
build-all: clean
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@echo "  -> darwin/amd64"
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	@echo "  -> darwin/arm64"
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	@echo "  -> linux/amd64"
	@GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	@echo "  -> linux/arm64"
	@GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	@echo "  -> windows/amd64"
	@GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Done!"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)

# Run without building
run:
	@$(GOCMD) run .

# Run tests
test:
	@$(GOTEST) -v ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@$(GOMOD) tidy

# Install to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME) 2>/dev/null || \
		cp $(BINARY_NAME) $(HOME)/go/bin/$(BINARY_NAME) 2>/dev/null || \
		(echo "Could not find Go bin directory. Please copy $(BINARY_NAME) manually." && exit 1)
	@echo "Installed to Go bin directory"

# Uninstall from GOPATH/bin
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f $(GOPATH)/bin/$(BINARY_NAME) 2>/dev/null || \
		rm -f $(HOME)/go/bin/$(BINARY_NAME) 2>/dev/null || \
		echo "Binary not found in Go bin directory"
	@echo "Done!"

# Help
help:
	@echo "Available targets:"
	@echo "  build      - Build for current platform"
	@echo "  build-all  - Build for all platforms (darwin, linux, windows)"
	@echo "  clean      - Remove build artifacts"
	@echo "  run        - Run without building"
	@echo "  test       - Run tests"
	@echo "  deps       - Download dependencies"
	@echo "  install    - Install to GOPATH/bin"
	@echo "  uninstall  - Remove from GOPATH/bin"
	@echo "  help       - Show this help"
