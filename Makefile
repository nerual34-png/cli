.PHONY: build install clean test run dev convex-dev convex-deploy

# Binary name
BINARY=grind
VERSION?=dev

# Build flags
LDFLAGS=-ldflags "-X grind/cmd.Version=$(VERSION)"

# Build the binary
build:
	go build $(LDFLAGS) -o $(BINARY) .

# Install to GOPATH/bin
install:
	go install $(LDFLAGS) .

# Clean build artifacts
clean:
	rm -f $(BINARY)
	rm -rf dist/

# Run tests
test:
	go test -v ./...

# Run the CLI
run: build
	./$(BINARY)

# Development mode - rebuild on changes (requires entr)
dev:
	find . -name '*.go' | entr -r make run

# Start Convex dev server
convex-dev:
	cd convex && npm install && npx convex dev

# Deploy Convex backend
convex-deploy:
	cd convex && npx convex deploy

# Build for all platforms
release:
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-amd64.exe .

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Check for updates
update:
	go get -u ./...
	go mod tidy
