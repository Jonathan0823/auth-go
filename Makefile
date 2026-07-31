ENV_FILE ?= .env
-include $(ENV_FILE)
export

check-database-url = $(if $(strip $(DATABASE_URL)),,$(error DATABASE_URL must be set in $(ENV_FILE) or the environment))

MIGRATE_DIR ?= migrations
MIGRATE ?= migrate

.PHONY: build test integration vet run sqlc migrate-up migrate-down migrate-down-1 fmt

build:
	go build ./cmd/auth-go

test:
	go test ./...

integration:
	$(MAKE) migrate-up
	go test -tags=integration ./...

vet:
	go vet ./...

run:
	go run ./cmd/auth-go

sqlc:
	sqlc generate

migrate-up:
	$(check-database-url)
	$(MIGRATE) -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" up

migrate-down:
	$(check-database-url)
	$(MIGRATE) -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" down

migrate-down-1:
	$(check-database-url)
	$(MIGRATE) -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" down 1

fmt:
	gofmt -w $$(find cmd internal -name '*.go')
