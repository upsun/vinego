#!/usr/bin/bash
set -xeu
go test ./...
golangci-lint run --timeout 10m -v