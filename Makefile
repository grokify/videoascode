.PHONY: build install test clean check run

# Build the binary
build:
	@echo "Building marp2video..."
	@go build -o bin/marp2video ./cmd/marp2video
	@echo "✓ Build complete: bin/marp2video"

# Install dependencies
install:
	@echo "Installing Go dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Go dependencies installed"
	@echo ""
	@echo "Please also ensure you have:"
	@echo "  - ffmpeg: brew install ffmpeg (macOS) or apt install ffmpeg (Linux)"
	@echo "  - marp CLI: npm install -g @marp-team/marp-cli"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf /tmp/marp2video/
	@echo "✓ Clean complete"

# Check dependencies
check:
	@./bin/marp2video --check || echo "Build the binary first with 'make build'"

# Run example
run: build
	@echo "Running example..."
	@./bin/marp2video \
		--input example_presentation.md \
		--output example_output.mp4 \
		--width 1920 \
		--height 1080

# Show help
help:
	@echo "Available targets:"
	@echo "  make build    - Build the marp2video binary"
	@echo "  make install  - Install Go dependencies"
	@echo "  make test     - Run tests"
	@echo "  make clean    - Clean build artifacts"
	@echo "  make check    - Check system dependencies"
	@echo "  make run      - Build and run example"
	@echo "  make help     - Show this help message"
