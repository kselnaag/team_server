#!/usr/bin/env bash
set -xe
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/stoilo_server ./cmd/main.go
