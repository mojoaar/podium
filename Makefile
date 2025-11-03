.PHONY: build install uninstall start stop restart clean run build-all help

# Variables
BIN_DIR = bin
BINARY_NAME = podium
BINARY_UNIX = $(BIN_DIR)/$(BINARY_NAME)
BINARY_WIN = $(BIN_DIR)/$(BINARY_NAME).exe

# Build the binary for current platform
build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BINARY_UNIX) .

# Build for specific platforms
build-linux:
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -o $(BIN_DIR)/$(BINARY_NAME)-linux-arm64 .

build-windows:
	@mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(BINARY_NAME)-windows-amd64.exe .

build-macos:
	@mkdir -p $(BIN_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64 .

# Build for all platforms using build script
build-all:
	@./build.sh all

# Service management (requires sudo on Linux)
install: build
	$(BINARY_UNIX) -service install

uninstall:
	$(BINARY_UNIX) -service uninstall

start:
	$(BINARY_UNIX) -service start

stop:
	$(BINARY_UNIX) -service stop

restart:
	$(BINARY_UNIX) -service restart

# Run in development mode (not as service)
run:
	go run main.go

# Clean build artifacts
clean:
	rm -rf $(BIN_DIR)
	rm -f podium podium-linux podium.exe podium-macos

# Help
help:
	@echo "Podium Makefile Commands:"
	@echo "  make build          - Build the binary for current platform"
	@echo "  make build-all      - Build for all platforms (uses build.sh)"
	@echo "  make build-linux    - Build for Linux (amd64 and arm64)"
	@echo "  make build-macos    - Build for macOS (amd64 and arm64)"
	@echo "  make build-windows  - Build for Windows (amd64)"
	@echo "  make install        - Install as system service"
	@echo "  make uninstall      - Uninstall system service"
	@echo "  make start          - Start the service"
	@echo "  make stop           - Stop the service"
	@echo "  make restart        - Restart the service"
	@echo "  make run            - Run in development mode"
	@echo "  make clean          - Remove build artifacts"
	@echo ""
	@echo "Build outputs are in the '$(BIN_DIR)/' directory"
