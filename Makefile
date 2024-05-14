LOCAL_BIN := $(CURDIR)/.bin

DB_DSN := postgres://db:db@localhost:5432/db_test?sslmode=disable
DB_MIGRATIONS_DIR := internal/database/migrations

export PATH := $(LOCAL_BIN):$(PATH)

#
# Build
#

default: build

.PHONY: build
build: build/server build/client

.PHONY: build/server
build/server:
	@echo "Building server..."
	@go build -o server ./cmd/server

.PHONY: build/client
build/client:
	@echo "Building client..."
	@go build -o sonar ./cmd/client


#
# Dev
#

.PHONY: dev
dev:
	@$(LOCAL_BIN)/air

#
# Tools
#

override GOBIN := $(LOCAL_BIN)

.PHONY: tools
tools: 
	@GOBIN=$(GOBIN) go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@GOBIN=$(GOBIN) go install github.com/cosmtrek/air@latest
	@GOBIN=$(GOBIN) go install github.com/abice/go-enum@latest
	@GOBIN=$(GOBIN) go install github.com/vektra/mockery/v2@latest 

#
# Test
#

.PHONY: test
test:
	@go test ./... -v -p 1 -coverprofile coverage.out

.PHONY: coverage/html
coverage/html:
	@go tool cover -html=coverage.out

#
# Lint & format
#

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

.PHONY: lint
lint:
	@echo "Linting code..."
	@golangci-lint run

#
# Code generation
#

.PHONY: generate
generate: generate/api generate/cmd generate/client generate/mocks

.PHONY: generate/api
generate/api:
	@echo "Generating API..."
	@go run ./internal/codegen/*.go -type api > internal/modules/api/generated.go
	@go fmt ./internal/modules/api

.PHONY: generate/cmd
generate/cmd:
	@echo "Generating CLI..."
	@go run ./internal/codegen/*.go -type cmd > internal/cmd/generated.go
	@go fmt ./internal/cmd

.PHONY: generate/client
generate/client:
	@echo "Generating API client..."
	@go run ./internal/codegen/*.go -type apiclient > internal/modules/api/apiclient/generated.go
	@go fmt ./internal/modules/api/apiclient

.PHONY: generate/mocks
generate/mocks:
	@echo "Generating mocks..."
	@$(LOCAL_BIN)/mockery \
		--dir internal/actions \
		--output internal/actions/mock \
		--outpkg actions_mock \
		--name Actions

#
# Migrations
#

override MIGRATE := $(LOCAL_BIN)/migrate -path $(DB_MIGRATIONS_DIR) -database $(DB_DSN)

.PHONY: migrations/create
migrations/create: guard-NAME
	@$(LOCAL_BIN)/migrate create -ext sql -dir $(DB_MIGRATIONS_DIR) -seq $(NAME)

.PHONY: migrations/up
migrations/up:
	@$(MIGRATE) up $(N)

.PHONY: migrations/down
migrations/down:
	@$(MIGRATE) down $(N)

.PHONY: migrations/goto
migrations/goto: guard-V
	@$(MIGRATE) goto $(V)

.PHONY: migrations/drop
migrations/drop:
	@$(MIGRATE) drop

.PHONY: migrations/force
migrations/force: guard-V
	@$(MIGRATE) force $(V)

#
# Helpers
#

guard-%:
	@if [ "${${*}}" = "" ]; then \
		echo "Environment variable $* not set"; \
		exit 1; \
	fi
