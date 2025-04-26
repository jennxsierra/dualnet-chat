APP_NAME := dualnet-chat
GO := go
PKG := ./...
PREFIX := [make]
BUILD_DIR := bin/

# Default target
.DEFAULT_GOAL := build

# Declare phony targets
.PHONY: fmt vet build test clean check run

fmt:
	@echo "$(PREFIX) Formatting source code..."
	@$(GO) fmt $(PKG)

vet: fmt
	@echo "$(PREFIX) Running vet to check code..."
	@$(GO) vet $(PKG)

build: vet
	@echo "$(PREFIX) Building $(APP_NAME) binaries in $(BUILD_DIR)..."
	@$(GO) build -o $(BUILD_DIR)

test:
	@echo "$(PREFIX) Running tests..."
	@$(GO) test $(PKG) -v

clean:
	@echo "$(PREFIX) Cleaning build artifacts..."
	@rm -rfv $(BUILD_DIR)

check: fmt vet test
	@echo "$(PREFIX) Quality checks (format, vet, tests) complete!"

# run: build
# 	@echo "$(PREFIX) Running $(APP_NAME)..."
# 	@$(BUILD_DIR)$(APP_NAME)