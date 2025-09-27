BINARY_NAME=ai-team

.PHONY: all test build clean

all: test build

test:
	@echo "Running tests..."
	@go test -v ./...

build:
	@echo "Building binary..."
	@go build -o $(BINARY_NAME) main.go

clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
