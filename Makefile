# --- PROJECT MAKEFILE ---

APP_NAME := dualnet-chat
GO := go
PKG := ./...
PREFIX := [make]
BUILD_DIR := bin/

# default server/client binaries (TCP)
SERVER_BINARY ?= $(BUILD_DIR)tcp-server
CLIENT_BINARY ?= $(BUILD_DIR)tcp-client

# default
.DEFAULT_GOAL := build

# dynamically find all directories inside cmd/ that have a main.go
BINARIES := $(shell find cmd -type f -name main.go | sed 's|/main.go||' | sed 's|^cmd/||')

# --- Build Section ---

.PHONY: fmt vet build test clean check run

fmt:
	@echo "$(PREFIX) Formatting source code..."
	@$(GO) fmt $(PKG)

vet: fmt
	@echo "$(PREFIX) Running vet to check code..."
	@$(GO) vet $(PKG)

build: vet
	@echo "$(PREFIX) Building $(APP_NAME) binaries in $(BUILD_DIR)..."
	@mkdir -p $(BUILD_DIR)
	@for bin in $(BINARIES); do \
		binary_name=$$(echo $$bin | tr '/' '-'); \
		echo "$(PREFIX) Building $$bin as $$binary_name..."; \
		$(GO) build -o $(BUILD_DIR)$$binary_name ./cmd/$$bin; \
	done

clean:
	@echo "$(PREFIX) Cleaning build artifacts..."
	@rm -rfv $(BUILD_DIR)

check: fmt vet test
	@echo "$(PREFIX) Quality checks (format, vet, tests) complete!"

# --- Network Tests Section ---

SERVER_PORT = 4000

# good network (fast and stable)
GOOD_LATENCY = 10ms
GOOD_LOSS = 0%
GOOD_RATE = 100mbit

# normal network (moderate)
NORMAL_LATENCY = 50ms
NORMAL_LOSS = 0.5%
NORMAL_RATE = 10mbit

# bad network (slow and unreliable)
BAD_LATENCY = 300ms
BAD_LOSS = 5%
BAD_RATE = 1mbit

.PHONY: impair-network clean-network run-server run-client test-network

# uses tc to define latency, loss, and rate
impair-network:
	@echo "$(PREFIX) Applying network impairments (latency=$(LATENCY), loss=$(LOSS), rate=$(RATE))..."
	sudo tc qdisc add dev lo root netem delay $(LATENCY) loss $(LOSS) rate $(RATE)

# -sudo signifies to continue even if command fails (e.g. network already clean)
clean-network:
	@echo "$(PREFIX) Cleaning network impairments..."
	-sudo tc qdisc del dev lo root 2>/dev/null

# tests current network
test-network: build clean-network
	@echo "$(PREFIX) Launching server and client under real (no impairment) network..."
	@{ \
		$(SERVER_BINARY) --port $(SERVER_PORT) & \
		SERVER_PID=$$!; \
		sleep 1; \
		go test -v -count=1 ./tests/network; \
		kill $$SERVER_PID; \
		wait $$SERVER_PID 2>/dev/null || true; \
	}

# internal helper for impaired tests
_test-network-impaired: build clean-network impair-network
	@echo "$(PREFIX) Launching server and client under impaired network..."
	@{ \
		$(SERVER_BINARY) --port $(SERVER_PORT) & \
		SERVER_PID=$$!; \
		sleep 1; \
		go test -v -count=1 ./tests/network; \
		kill $$SERVER_PID; \
		wait $$SERVER_PID 2>/dev/null || true; \
	}
	@$(MAKE) clean-network

# --- TCP Network Test Shortcuts ---

test-tcp-network: test-network

test-tcp-network-good:
	@$(MAKE) -s SERVER_BINARY=$(BUILD_DIR)tcp-server CLIENT_BINARY=$(BUILD_DIR)tcp-client \
		LATENCY=10ms LOSS=0.1% RATE=10mbit \
		_test-network-impaired

test-tcp-network-normal:
	@$(MAKE) -s SERVER_BINARY=$(BUILD_DIR)tcp-server CLIENT_BINARY=$(BUILD_DIR)tcp-client \
		LATENCY=50ms LOSS=1% RATE=5mbit \
		_test-network-impaired

test-tcp-network-bad:
	@$(MAKE) -s SERVER_BINARY=$(BUILD_DIR)tcp-server CLIENT_BINARY=$(BUILD_DIR)tcp-client \
		LATENCY=200ms LOSS=5% RATE=1mbit \
		_test-network-impaired

# --- UDP Network Test Shortcuts ---

test-udp-network: test-network SERVER_BINARY=$(BUILD_DIR)udp-server CLIENT_BINARY=$(BUILD_DIR)udp-client

test-udp-network-good:
	@$(MAKE) -s SERVER_BINARY=$(BUILD_DIR)udp-server CLIENT_BINARY=$(BUILD_DIR)udp-client \
		LATENCY=10ms LOSS=0.1% RATE=10mbit \
		_test-network-impaired

test-udp-network-normal:
	@$(MAKE) -s SERVER_BINARY=$(BUILD_DIR)udp-server CLIENT_BINARY=$(BUILD_DIR)udp-client \
		LATENCY=50ms LOSS=1% RATE=5mbit \
		_test-network-impaired

test-udp-network-bad:
	@$(MAKE) -s SERVER_BINARY=$(BUILD_DIR)udp-server CLIENT_BINARY=$(BUILD_DIR)udp-client \
		LATENCY=200ms LOSS=5% RATE=1mbit \
		_test-network-impaired
