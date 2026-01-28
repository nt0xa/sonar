SHELL := /bin/bash

# Docker compose
COMPOSE := docker compose --project-name sonar --project-directory . --file dev/docker-compose.yml
IN_DOCKER := $(shell [ -f /.dockerenv ] && echo 1)

# Required because some commands need to be run inside the container.
ifdef IN_DOCKER
    EXEC ?=
else
    EXEC ?= $(COMPOSE) exec dev
endif

#
# Build
#

default: build

.PHONY: build
build: build/server build/client

.PHONY: build/server
build/server:
	@echo "Building server..."
	@$(EXEC) mkdir -p build
	@$(EXEC) go build -o build/server ./cmd/server

.PHONY: build/client
build/client:
	@echo "Building client..."
	@$(EXEC) mkdir -p build
	@$(EXEC) go build -o build/sonar ./cmd/client

.PHONY: clean/build
clean/build:
	@echo "Cleaning build artifacts..."
	@$(EXEC) rm -rf build completions coverage.out

#
# Shell completions
#

.PHONY: completions
completions: build/client
	@echo "Generating shell completions..."
	@$(EXEC) rm -rf completions
	@$(EXEC) mkdir -p completions
	@$(EXEC) build/sonar completion bash > completions/sonar.bash
	@$(EXEC) build/sonar completion zsh > completions/sonar.zsh
	@$(EXEC) build/sonar completion fish > completions/sonar.fish

#
# Release
#

.PHONY: release/snapshot
release/snapshot:
	@$(EXEC) goreleaser release --clean --snapshot

#
# Docker compose commands.
#

# Start the development environment.
.PHONY: up
up:
	@$(COMPOSE) up

# Stop the development environment.
.PHONY: down
down:
	@$(COMPOSE) down

# Restart the development container.
.PHONY: restart
restart:
	@$(COMPOSE) restart dev

# Recreates the development container.
.PHONY: recreate
recreate:
	@$(COMPOSE) rm --force --stop dev
	@$(COMPOSE) up --detach

# Clean docker volumes.
.PHONY: clean/volumes
clean/volumes:
	@$(COMPOSE) down --volumes

# Clean docker images.
.PHONY: clean/images
clean/images:
	@$(COMPOSE) down --rmi local

# Start a shell inside the development container.
.PHONY: enter
enter:
	@$(COMPOSE) exec dev bash

#
# Watch (auto-restart)
#

# Rebuilds both server and client on code changes and restarts the server.
# Used as a dev container command in dev/docker-compose.yml.
.PHONY: watch
watch:
	@$(EXEC) air \
		-build.bin build/server \
		-build.cmd "make build" \
		-build.exclude_dir docs \
		-misc.clean_on_exit true

#
# Tools
#

# Used in dev/Dockerfile to install devtools inside a development container image.
.PHONY: devtools
devtools: 
	@echo "Installing development tools..."
	@go install -tags "postgres" github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/air-verse/air@latest
	@go install github.com/abice/go-enum@latest
	@go install github.com/vektra/mockery/v2@latest 
	@go install github.com/goreleaser/goreleaser/v2@latest
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest


#
# Test
#

.PHONY: test
test:
	@echo "Running tests..."
	@$(EXEC) go test ./... -v -p 1 -coverprofile coverage.out

.PHONY: coverage
coverage:
	@$(EXEC) go tool cover -html coverage.out -o coverage.html

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
	@$(EXEC) go run ./internal/codegen/*.go -type api > internal/modules/api/generated.go
	@$(MAKE) fmt

.PHONY: generate/cmd
generate/cmd:
	@echo "Generating CLI..."
	@$(EXEC) go run ./internal/codegen/*.go -type cmd > internal/cmd/generated.go
	@$(MAKE) fmt

.PHONY: generate/client
generate/client:
	@echo "Generating API client..."
	@$(EXEC) go run ./internal/codegen/*.go -type apiclient > internal/modules/api/apiclient/generated.go
	@$(MAKE) fmt

.PHONY: generate/mocks
generate/mocks:
	@echo "Generating mocks..."
	@$(EXEC) mockery \
		--dir internal/actions \
		--output internal/actions/mock \
		--outpkg actions_mock \
		--name Actions

.PHONY: generate/db
generate/db:
	@echo "Generating datatabase access code..."
	@$(EXEC) sqlc vet
	@$(EXEC) sqlc generate

#
# Migrations
#

MIGRATE = $(EXEC) sh -c 'migrate -path internal/database/migrations -database "$$SONAR_DB_DSN" "$$@"' _

.PHONY: migrations/create
migrations/create: guard-NAME
	@$(EXEC) migrate create -ext sql -dir internal/database/migrations -seq $(NAME)

.PHONY: migrations/up
migrations/up:
	@$(MIGRATE) up $(N)

.PHONY: migrations/down
migrations/down: guard-N
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
		echo "ERROR: Environment variable \"$*\" not set"; \
		exit 1; \
	fi
