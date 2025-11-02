#!/bin/bash

set -e

echo "Running gofmt..."
go fmt ./...

echo "Running go vet..."
go vet ./...

echo "Running tests..."
go test -v -race -coverprofile=coverage.out ./...

echo "Building..."
go build -o minipay ./main.go

echo "✓ All checks passed!"
