.DEFAULT_GOAL := _setup


.PHONY: _setup
_setup:
	@node .github/setup.js
SHELL := /bin/bash

.PHONY: build test lint security run release-snapshot clean

build:
	CGO_ENABLED=0 go build -o bin/memoire ./cmd/memoire

test:
	go test ./... -count=1 -race

test-cover:
	go test ./... -count=1 -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out | tail -1

lint:
	go vet ./...
	gofmt -l .
	@if [ -n "$$(gofmt -l .)" ]; then echo "files not gofmt'd" >&2; exit 1; fi

security:
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	$$(go env GOPATH)/bin/govulncheck ./...

run:
	go run ./cmd/memoire

release-snapshot:
	@go install github.com/goreleaser/goreleaser/v2@latest
	$$(go env GOPATH)/bin/goreleaser release --snapshot --clean

clean:
	rm -rf bin/ dist/ coverage.out
