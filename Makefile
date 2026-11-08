.PHONY: all build test test-race lint clean run

BINARY_NAME=searchforge
SRC_DIR=./cmd/searchforge

all: lint test-race build

build:
	@mkdir -p bin
	go build -ldflags="-s -w" -o bin/$(BINARY_NAME) $(SRC_DIR)

run: build
	./bin/$(BINARY_NAME)

test:
	go test -v ./...

test-race:
	go test -v -race -count=1 ./...

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, running go vet..."; \
		go vet ./...; \
	fi

clean:
	rm -rf bin/ data/ *.out *.test

