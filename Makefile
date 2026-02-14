.PHONY: build release clean help

# Binary name
BINARY_NAME=regcrawler

# Build the application
build:
	go build -o $(BINARY_NAME) ./cmd/regcrawler

# Build for production (smaller binary, no debug info)
release:
	go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/regcrawler

# Clean build artifacts and generated files
clean:
	rm -f $(BINARY_NAME)

# Show help
help:
	@echo "Usage:"
	@echo "  make build    - Build the binary (with debug info)"
	@echo "  make release  - Build for production (smaller file size)"
	@echo "  make clean    - Remove build artifacts and generated files"
