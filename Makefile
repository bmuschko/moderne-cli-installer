.PHONY: build build-all clean run test

BINARY_NAME=moderne-cli-installer
VERSION?=dev

# Build for current platform
build:
	go build -o $(BINARY_NAME) .

# Build for all platforms
build-all:
	./build.sh $(VERSION)

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

# Run locally (requires version argument)
run:
	go run . -version $(VERSION)

# Test the build
test:
	go build -o /dev/null .
	@echo "Build successful"
