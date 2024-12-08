.PHONY: build run generate clean

# Default binary name
BINARY_NAME=outreach-generator

build: generate
	go build -o $(BINARY_NAME) cmd/server/main.go

run: generate
	go run cmd/server/main.go

generate:
	templ generate

clean:
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe
	rm -f $(BINARY_NAME)-mac-intel
	rm -f $(BINARY_NAME)-mac-apple
	rm -f $(BINARY_NAME)-mac-universal
	rm -f $(BINARY_NAME)-linux

# Cross compilation
build-all: generate
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-mac-intel cmd/server/main.go
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-mac-apple cmd/server/main.go
	# Create universal macOS binary
	lipo -create -output $(BINARY_NAME)-mac-universal \
		$(BINARY_NAME)-mac-intel \
		$(BINARY_NAME)-mac-apple
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux cmd/server/main.go
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME).exe cmd/server/main.go

# Development tools
dev-deps:
	go install github.com/a-h/templ/cmd/templ@latest 