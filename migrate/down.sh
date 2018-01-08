#!/usr/bin/env bash

source ./migrate/defaults.sh

## Migrate is needed:
# go get -u github.com/golang-migrate/migrate
# go build -tags 'postgres' -o /usr/local/bin/migrate github.com/golang-migrate/migrate/cli

migrate -database postgres://${PRIZEM_POSTGRES_USER}:${PRIZEM_POSTGRES_PASSWORD}@${PRIZEM_POSTGRES_HOST}:${PRIZEM_POSTGRES_PORT}/${PRIZEM_POSTGRES_NAME}?sslmode=disable -path ./sql down
