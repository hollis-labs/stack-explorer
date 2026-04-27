.PHONY: build build-full install test lint eval

build:
	go build -o stack-explorer ./cmd/stack-explorer

build-full:
	go build -tags treesitter -o stack-explorer ./cmd/stack-explorer

install:
	go install ./cmd/stack-explorer

test:
	go test ./...

lint:
	go vet ./...

eval:
	go run ./cmd/stack-explorer eval --queries eval/queries.yaml --baseline eval/baseline.json
