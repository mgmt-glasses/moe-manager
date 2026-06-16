.PHONY: build test vet run verify

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

run:
	go run ./cmd/api

verify: vet test build
