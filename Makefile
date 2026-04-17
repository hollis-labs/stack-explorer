.PHONY: build install test lint

build:
	go build -o stack-explorer ./cmd/stack-explorer

install:
	go install ./cmd/stack-explorer

test:
	go test ./...

lint:
	go vet ./...
