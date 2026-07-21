.PHONY: run build test test-race lint vet fmt tidy up down migrate-up migrate-down

run:
	go run ./cmd/auth

build:
	go build ./...

test:
	go test -coverprofile=coverage.out ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w cmd configs internal pkg

tidy:
	go mod tidy

lint:
	golangci-lint run

up:
	docker compose up --build

down:
	docker compose down

migrate-up:
	docker compose run --rm migrate

migrate-down:
	docker compose run --rm migrate -path /migrations -database "postgres://postgres:postgres@postgres:5432/auth?sslmode=disable" down 1
