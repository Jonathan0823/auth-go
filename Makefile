DATABASE_URL ?=
ifeq ($(strip $(DATABASE_URL)),)
$(error DATABASE_URL must be set in the environment)
endif
MIGRATE_DIR ?= migrations
MIGRATE ?= migrate

.PHONY: build test vet run sqlc migrate-up migrate-down migrate-down-1 fmt

build:
	go build ./cmd/auth-go

test:
	go test ./...

vet:
	go vet ./...

run:
	go run ./cmd/auth-go

sqlc:
	sqlc generate

migrate-up:
	$(MIGRATE) -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" up

migrate-down:
	$(MIGRATE) -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" down

migrate-down-1:
	$(MIGRATE) -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" down 1

fmt:
	gofmt -w $$(find cmd internal -name '*.go')
