ENV_FILE ?= .env
-include $(ENV_FILE)
export

check-database-url = $(if $(strip $(DATABASE_URL)),,$(error DATABASE_URL must be set in $(ENV_FILE) or the environment))

MIGRATE_DIR ?= migrations
MIGRATE ?= migrate
SWAG_VERSION ?= v1.16.6
REDOCLY_VERSION ?= v2.43.1
COVERAGE_DIR ?= coverage
COVERAGE_FILE ?= $(COVERAGE_DIR)/coverage.out

.PHONY: build test coverage integration vet run sqlc swagger swagger-validate migrate-up migrate-down migrate-down-1 fmt

build:
	go build ./cmd/auth-go

test:
	go test ./...

coverage: migrate-up
	mkdir -p "$(COVERAGE_DIR)"
	go test -tags=integration ./... -covermode=atomic -coverprofile="$(COVERAGE_FILE)"

integration:
	$(MAKE) migrate-up
	go test -tags=integration ./...

vet:
	go vet ./...

run:
	go run ./cmd/auth-go

sqlc:
	sqlc generate

swagger:
	go run github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION) init -g cmd/auth-go/main.go --parseDependency --parseInternal -o docs

swagger-validate: swagger
	npx --yes @redocly/cli@$(REDOCLY_VERSION) lint docs/swagger.yaml

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
