.DEFAULT_GOAL := bin/screensaver

GOPATH := $(shell go env GOPATH)
VERSION ?= master

PLATFORMS := linux darwin windows
ARCHITECTURES := amd64 arm64

.PHONY: build test run lint clean install-tools certs help demo release

##@ Packaging

bin/screensaver:
	@go build -o bin/screensaver main.go

# Build binary
build: bin/screensaver ## Build binary

release: clean ## Build release binaries for all platforms
	@for platform in $(PLATFORMS); do \
		for arch in $(ARCHITECTURES); do \
			ext=""; \
			if [ "$$platform" = "windows" ]; then ext=".exe"; fi; \
			output="bin/screensaver_$(VERSION)_$${platform}.$${arch}$${ext}"; \
			echo "Building $$output..."; \
			GOOS=$$platform GOARCH=$$arch go build -o $$output main.go; \
		done; \
	done
	@go run .

##@ Development commands

# Install
install: ## Install dependencies
	@go install .
	@echo "Installed to $(GOPATH)/bin/screensaver"

.PHONY: uninstall
uninstall:
	@rm -f $(GOPATH)/bin/screensaver
	@echo "Uninstalled"

# Lint code
lint: ## Lint code
	go vet ./...
	go fmt ./...

# Clean build artifacts
clean: ## Clean build artifacts
	@rm -rf bin/

install-tools: ## Install development tools
	@go install github.com/charmbracelet/vhs@latest
	@echo "Development tools installed!"

##@ Other

demo: ## Run vhs demo
	@cd assets && vhs demo.tape

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
