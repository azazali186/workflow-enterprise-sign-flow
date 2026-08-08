# AeroXe SignFlow backend
SHELL := /bin/bash
GO ?= go
GOARCH ?= amd64

.PHONY: build run test test-race vet fmt tidy swagger migrate infra-up infra-down docker-build clean

## build: compile the server and migration binaries
build:
	GOARCH=$(GOARCH) $(GO) build -o bin/server ./cmd/server
	GOARCH=$(GOARCH) $(GO) build -o bin/migrate ./cmd/migrate

## run: run the API server (requires infra)
run:
	GOARCH=$(GOARCH) $(GO) run ./cmd/server

## migrate: apply database migrations (GORM AutoMigrate)
migrate:
	GOARCH=$(GOARCH) $(GO) run ./cmd/migrate

## test: run the unit and integration test suite
test:
	GOARCH=$(GOARCH) $(GO) test ./... -count=1

## test-race: run tests with the race detector
test-race:
	GOARCH=$(GOARCH) $(GO) test -race ./... -count=1

## vet: static analysis
vet:
	GOARCH=$(GOARCH) $(GO) vet ./...

## fmt: format all sources
fmt:
	gofmt -l -w cmd internal

## tidy: sync go.mod / go.sum
tidy:
	$(GO) mod tidy

## swagger: regenerate OpenAPI docs from annotations
swagger:
	GOARCH=$(GOARCH) swag init -g cmd/server/main.go --output docs

## infra-up: start Postgres, Redis and NATS via docker compose
infra-up:
	docker compose up -d postgres redis nats

## infra-down: stop the local infrastructure
infra-down:
	docker compose down

## docker-build: build the server container image
docker-build:
	docker build -t signflow-backend:latest .

## clean: remove build artifacts
clean:
	rm -rf bin
