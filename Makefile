# Makefile for QNAP Display Control (Go + Bazelisk)
# ================================================

# Project configuration
PROJECT_NAME := qnap_display_control
VERSION := 0.1.0
GO_VERSION := 1.21.5
BAZELISK_VERSION := 1.19.0

# Directories
BIN_DIR := bin
BUILD_DIR := bazel-bin
INSTALL_DIR := /usr/local/bin
TEMP_DIR := /tmp/qnap-build

# Platform detection
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_S),Linux)
    PLATFORM := linux
endif
ifeq ($(UNAME_S),Darwin)
    PLATFORM := darwin
endif

ifeq ($(UNAME_M),x86_64)
    ARCH := amd64
endif
ifeq ($(UNAME_M),aarch64)
    ARCH := arm64
endif
ifeq ($(UNAME_M),arm64)
    ARCH := arm64
endif

# URLs
BAZELISK_URL := https://github.com/bazelbuild/bazelisk/releases/download/v$(BAZELISK_VERSION)/bazelisk-$(PLATFORM)-$(ARCH)
GO_URL := https://go.dev/dl/go$(GO_VERSION).$(PLATFORM)-$(ARCH).tar.gz

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
PURPLE := \033[0;35m
CYAN := \033[0;36m
NC := \033[0m # No Color

# Default target
.PHONY: help
help: ## Show this help message
	@echo "$(CYAN)QNAP Display Control - Go/Bazelisk Build System$(NC)"
	@echo "$(CYAN)=================================================$(NC)"
	@echo ""
	@echo "$(YELLOW)Setup Commands:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST) | grep -E "(setup|install|deps)"
	@echo ""
	@echo "$(YELLOW)Build Commands:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST) | grep -E "(build|test|clean)"
	@echo ""
	@echo "$(YELLOW)Development Commands:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST) | grep -E "(run|dev|format|lint)"
	@echo ""
	@echo "$(YELLOW)Deployment Commands:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST) | grep -E "(deploy|package|release)"
	@echo ""

# ============================================================================
# Setup and Installation Commands
# ============================================================================

.PHONY: setup
setup: ## Complete development environment setup
	@echo "$(CYAN)Setting up QNAP Display Control development environment...$(NC)"
	@$(MAKE) check-system
	@$(MAKE) install-bazelisk
	@$(MAKE) install-go
	@$(MAKE) verify-deps
	@echo "$(GREEN)✅ Setup completed successfully!$(NC)"

.PHONY: check-system
check-system: ## Check system requirements
	@echo "$(BLUE)Checking system requirements...$(NC)"
	@echo "Platform: $(PLATFORM)-$(ARCH)"
	@which curl > /dev/null || (echo "$(RED)❌ curl is required$(NC)" && exit 1)
	@which wget > /dev/null || which curl > /dev/null || (echo "$(RED)❌ wget or curl is required$(NC)" && exit 1)
	@[ "$(PLATFORM)" = "linux" ] || [ "$(PLATFORM)" = "darwin" ] || (echo "$(RED)❌ Unsupported platform: $(UNAME_S)$(NC)" && exit 1)
	@echo "$(GREEN)✅ System requirements met$(NC)"

.PHONY: install-bazelisk
install-bazelisk: ## Install Bazelisk (Bazel version manager)
	@echo "$(BLUE)Installing Bazelisk v$(BAZELISK_VERSION)...$(NC)"
	@if command -v bazel > /dev/null 2>&1; then \
		echo "$(YELLOW)⚠️  Bazel/Bazelisk already installed: $$(bazel version | head -1)$(NC)"; \
	else \
		echo "$(BLUE)Downloading Bazelisk...$(NC)"; \
		mkdir -p $(TEMP_DIR); \
		curl -fsSL $(BAZELISK_URL) -o $(TEMP_DIR)/bazelisk; \
		chmod +x $(TEMP_DIR)/bazelisk; \
		if [ -w $(INSTALL_DIR) ]; then \
			mv $(TEMP_DIR)/bazelisk $(INSTALL_DIR)/bazel; \
		else \
			echo "$(YELLOW)Installing to $(INSTALL_DIR) (requires sudo)...$(NC)"; \
			sudo mv $(TEMP_DIR)/bazelisk $(INSTALL_DIR)/bazel; \
		fi; \
		echo "$(GREEN)✅ Bazelisk installed successfully$(NC)"; \
	fi

.PHONY: install-go
install-go: ## Install Go toolchain
	@echo "$(BLUE)Installing Go toolchain...$(NC)"
	@if command -v go > /dev/null 2>&1; then \
		echo "$(YELLOW)⚠️  Go already installed: $$(go version)$(NC)"; \
	else \
		echo "$(BLUE)Downloading and installing Go $(GO_VERSION)...$(NC)"; \
		curl -L $(GO_URL) -o $(TEMP_DIR)/go.tar.gz; \
		sudo rm -rf /usr/local/go; \
		sudo tar -C /usr/local -xzf $(TEMP_DIR)/go.tar.gz; \
		rm $(TEMP_DIR)/go.tar.gz; \
		echo "$(GREEN)✅ Go installed successfully$(NC)"; \
		echo "$(YELLOW)Please add /usr/local/go/bin to your PATH$(NC)"; \
	fi

.PHONY: install-deps
install-deps: ## Install additional system dependencies
	@echo "$(BLUE)Installing system dependencies...$(NC)"
	@if command -v apt-get > /dev/null 2>&1; then \
		echo "$(BLUE)Installing dependencies for Ubuntu/Debian...$(NC)"; \
		sudo apt-get update; \
		sudo apt-get install -y build-essential curl wget git pkg-config; \
	elif command -v yum > /dev/null 2>&1; then \
		echo "$(BLUE)Installing dependencies for RHEL/CentOS...$(NC)"; \
		sudo yum groupinstall -y "Development Tools"; \
		sudo yum install -y curl wget git pkgconfig; \
	elif command -v pacman > /dev/null 2>&1; then \
		echo "$(BLUE)Installing dependencies for Arch Linux...$(NC)"; \
		sudo pacman -S --needed base-devel curl wget git pkgconf; \
	else \
		echo "$(YELLOW)⚠️  Unknown package manager. Please install build tools manually.$(NC)"; \
	fi
	@echo "$(GREEN)✅ System dependencies installed$(NC)"

.PHONY: verify-deps
verify-deps: ## Verify all dependencies are installed correctly
	@echo "$(BLUE)Verifying dependencies...$(NC)"
	@command -v bazel > /dev/null || (echo "$(RED)❌ Bazelisk/Bazel not found$(NC)" && exit 1)
	@command -v go > /dev/null || (echo "$(RED)❌ Go not found$(NC)" && exit 1)
	@echo "$(GREEN)✅ Bazel: $$(bazel version | head -1)$(NC)"
	@echo "$(GREEN)✅ Go: $$(go version)$(NC)"

# ============================================================================
# Build Commands
# ============================================================================

.PHONY: build
build: ## Build all targets
	@echo "$(BLUE)Building all targets...$(NC)"
	@bazel build //...
	@$(MAKE) copy-binaries
	@echo "$(GREEN)✅ Build completed$(NC)"

.PHONY: build-opt
build-opt: ## Build optimized release version
	@echo "$(BLUE)Building optimized release...$(NC)"
	@bazel build --config=opt //...
	@$(MAKE) copy-binaries
	@echo "$(GREEN)✅ Optimized build completed$(NC)"

.PHONY: build-static
build-static: ## Build static binary for deployment
	@echo "$(BLUE)Building static binary...$(NC)"
	@bazel build --config=static //...
	@$(MAKE) copy-binaries
	@echo "$(GREEN)✅ Static build completed$(NC)"

.PHONY: build-debug
build-debug: ## Build debug version with symbols
	@echo "$(BLUE)Building debug version...$(NC)"
	@bazel build --config=debug //...
	@$(MAKE) copy-binaries
	@echo "$(GREEN)✅ Debug build completed$(NC)"

.PHONY: build-go
build-go: ## Build with Go directly (without Bazel)
	@echo "$(BLUE)Building with Go directly...$(NC)"
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/qnap-display-control cmd/main.go
	@echo "$(GREEN)✅ Go build completed: $(BIN_DIR)/qnap-display-control$(NC)"
	@ls -la $(BIN_DIR)/qnap-display-control

.PHONY: copy-binaries
copy-binaries: ## Copy built binaries to bin/ directory
	@echo "$(BLUE)Copying binaries...$(NC)"
	@mkdir -p $(BIN_DIR)
	@if [ -f "$(BUILD_DIR)/qnap-display-control_/qnap-display-control" ]; then \
		cp $(BUILD_DIR)/qnap-display-control_/qnap-display-control $(BIN_DIR)/; \
		chmod +x $(BIN_DIR)/qnap-display-control; \
	fi
	@if [ -f "$(BUILD_DIR)/cmd/cmd_/cmd" ]; then \
		cp $(BUILD_DIR)/cmd/cmd_/cmd $(BIN_DIR)/cmd; \
		chmod +x $(BIN_DIR)/cmd; \
	fi
	@echo "$(GREEN)✅ Binaries copied to $(BIN_DIR)/$(NC)"

.PHONY: test
test: ## Run all tests
	@echo "$(BLUE)Running tests...$(NC)"
	@bazel test //...
	@echo "$(GREEN)✅ All tests passed$(NC)"

.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo "$(BLUE)Running unit tests...$(NC)"
	@bazel test //:unit_tests
	@echo "$(GREEN)✅ Unit tests passed$(NC)"

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "$(BLUE)Running integration tests...$(NC)"
	@bazel test //:integration_tests
	@echo "$(GREEN)✅ Integration tests passed$(NC)"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	@bazel clean
	@rm -rf $(BIN_DIR)
	@rm -rf $(TEMP_DIR)
	@echo "$(GREEN)✅ Cleaned$(NC)"

.PHONY: clean-all
clean-all: ## Clean everything including Bazel cache
	@echo "$(BLUE)Cleaning everything...$(NC)"
	@bazel clean --expunge
	@rm -rf $(BIN_DIR)
	@rm -rf $(TEMP_DIR)
	@echo "$(GREEN)✅ Deep clean completed$(NC)"

# ============================================================================
# Development Commands
# ============================================================================

.PHONY: run
run: build ## Build and run the main program
	@echo "$(BLUE)Running QNAP Display Control...$(NC)"
	@echo "$(YELLOW)Note: Requires root privileges for hardware access$(NC)"
	@sudo ./$(BIN_DIR)/qnap_display_control

.PHONY: run-usb-test
run-usb-test: build ## Build and run USB COPY button test
	@echo "$(BLUE)Running USB COPY button test...$(NC)"
	@echo "$(YELLOW)Note: Requires root privileges for I/O port access$(NC)"
	@sudo ./$(BIN_DIR)/qnap_display_control --test-usb-copy

.PHONY: run-example
run-example: build ## Build and run USB copy example
	@echo "$(BLUE)Running USB copy example...$(NC)"
	@echo "$(YELLOW)Note: Requires root privileges for hardware access$(NC)"
	@sudo ./$(BIN_DIR)/usb_copy_example

.PHONY: dev
dev: ## Quick development build
	@echo "$(BLUE)Development build...$(NC)"
	@bazel build --config=dev //...
	@$(MAKE) copy-binaries

.PHONY: format
format: ## Format Go code
	@echo "$(BLUE)Formatting Go code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✅ Code formatted$(NC)"

.PHONY: lint
lint: ## Run Go linter (golangci-lint)
	@echo "$(BLUE)Running Go linter...$(NC)"
	@if command -v golangci-lint > /dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)Installing golangci-lint...$(NC)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi
	@echo "$(GREEN)✅ Linting completed$(NC)"

.PHONY: check
check: format lint test ## Run all code quality checks
	@echo "$(GREEN)✅ All checks passed$(NC)"

# ============================================================================
# Deployment Commands
# ============================================================================

.PHONY: package
package: build-static ## Create deployment package
	@echo "$(BLUE)Creating deployment package...$(NC)"
	@mkdir -p dist/$(PROJECT_NAME)-$(VERSION)
	@cp $(BIN_DIR)/* dist/$(PROJECT_NAME)-$(VERSION)/
	@cp README.md dist/$(PROJECT_NAME)-$(VERSION)/
	@cd dist && tar -czf $(PROJECT_NAME)-$(VERSION)-$(PLATFORM)-$(ARCH).tar.gz $(PROJECT_NAME)-$(VERSION)
	@echo "$(GREEN)✅ Package created: dist/$(PROJECT_NAME)-$(VERSION)-$(PLATFORM)-$(ARCH).tar.gz$(NC)"

.PHONY: deploy-local
deploy-local: build-static ## Deploy binaries to local system
	@echo "$(BLUE)Deploying to local system...$(NC)"
	@echo "$(YELLOW)Installing to $(INSTALL_DIR) (requires sudo)...$(NC)"
	@sudo cp $(BIN_DIR)/qnap_display_control $(INSTALL_DIR)/
	@sudo cp $(BIN_DIR)/usb_copy_example $(INSTALL_DIR)/
	@sudo chmod +x $(INSTALL_DIR)/qnap_display_control
	@sudo chmod +x $(INSTALL_DIR)/usb_copy_example
	@echo "$(GREEN)✅ Deployed successfully$(NC)"

.PHONY: release
release: clean build-static test package ## Build release version
	@echo "$(GREEN)🚀 Release $(VERSION) ready!$(NC)"
	@ls -la dist/

# ============================================================================
# Utility Commands
# ============================================================================

.PHONY: info
info: ## Show build information
	@echo "$(CYAN)Project Information$(NC)"
	@echo "$(CYAN)==================$(NC)"
	@echo "Project: $(PROJECT_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Platform: $(PLATFORM)-$(ARCH)"
	@echo "Go Version: $(GO_VERSION)"
	@echo "Bazelisk Version: $(BAZELISK_VERSION)"
	@echo ""
	@if command -v bazel > /dev/null 2>&1; then \
		echo "$(CYAN)Bazel Information$(NC)"; \
		echo "$(CYAN)=================$(NC)"; \
		bazel version | head -5; \
		echo ""; \
	fi
	@if command -v go > /dev/null 2>&1; then \
		echo "$(CYAN)Go Information$(NC)"; \
		echo "$(CYAN)==============$(NC)"; \
		go version; \
		echo ""; \
	fi

.PHONY: deps-graph
deps-graph: ## Show dependency graph
	@echo "$(BLUE)Generating dependency graph...$(NC)"
	@bazel mod graph

.PHONY: workspace-status
workspace-status: ## Show workspace status
	@echo "$(BLUE)Workspace status...$(NC)"
	@bazel info

.PHONY: verify
verify: ## Verify Bzlmod configuration
	@echo "$(BLUE)Verifying Bzlmod configuration...$(NC)"
	@./verify_bzlmod.sh

# ============================================================================
# Special targets
# ============================================================================

.PHONY: bootstrap
bootstrap: ## Bootstrap development environment (same as setup)
	@$(MAKE) setup

.PHONY: docker
docker: ## Build in Docker container (for CI/CD)
	@echo "$(BLUE)Building in Docker container...$(NC)"
	@docker build -t $(PROJECT_NAME):$(VERSION) .
	@echo "$(GREEN)✅ Docker build completed$(NC)"

# Help is the default target
.DEFAULT_GOAL := help
