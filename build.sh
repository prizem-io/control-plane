#!/bin/sh
CGO_ENABLED=0 go build -a --installsuffix cgo --ldflags="-s" -o build/control-plane cmd/control-plane/main.go
docker build -t prizem-io/control-plane .