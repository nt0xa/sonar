SHELL := /bin/bash

# Docker compose
COMPOSE := docker compose --project-name sonar --project-directory . --file dev/docker-compose.yml
IN_DOCKER := $(shell [ -f /.dockerenv ] && echo 1)

ifdef IN_DOCKER
    EXEC :=
else
    EXEC := $(COMPOSE) exec server
endif

#
# Build
#

default: build

.PHONY: build
build: build/server build/client

.PHONY: build/server
build/server: require-container-server
	@echo "Building server..."
	@$(EXEC) mkdir -p build
	@$(EXEC) go build -o build/server ./cmd/server

.PHONY: build/client
build/client: require-container-server
	@echo "Building client..."
	@$(EXEC) mkdir -p build
	@$(EXEC) go build -o build/sonar ./cmd/client

clean: require-container-server
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
release/snapshot: require-container-server
	@$(EXEC) goreleaser release --clean --snapshot

#
# Docker compose
#

.PHONY: up
up:
	@$(COMPOSE) up

.PHONY: down
down:
	@$(COMPOSE) down

.PHONY: restart/server
restart/server:
	@$(COMPOSE) restart server

.PHONY: restart/client
restart/client:
	@$(COMPOSE) restart client

.PHONY: recreate/server
recreate/server:
	@$(COMPOSE) rm --force --stop server
	@$(COMPOSE) up --detach

.PHONY: clean/volumes
clean/volumes:
	@$(COMPOSE) down --volumes

.PHONY: clean/images
clean/images:
	@$(COMPOSE) down --rmi local

.PHONY: exec/client
exec/client:
	@$(COMPOSE) exec client bash

.PHONY: exec/server
exec/server:
	@$(COMPOSE) exec server bash

#
# Watch (auto-restart)
#

.PHONY: watch/server
watch/server:
	@$(EXEC) air \
		-build.bin build/server \
		-build.cmd "make build/server" \
		-build.exclude_dir docs \
		-misc.clean_on_exit true

.PHONY: watch/client
watch/client:
	@$(EXEC) air \
		-build.bin /usr/bin/true \
		-build.cmd "make build/client" \
		-build.exclude_dir docs \
		-misc.clean_on_exit true

#
# Docs
#

.PHONY: docs/dev
docs/dev:
	@cd docs && npm run start -- --host 0.0.0.0 --port 3000

.PHONY: docs/deps
docs/deps:
	@cd docs && npm install


#
# Tools
#

.PHONY: devtools
devtools: 
	@echo "Installing development tools..."
	@go install -tags "postgres" github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/air-verse/air@latest
	@go install github.com/abice/go-enum@latest
	@go install github.com/vektra/mockery/v2@latest 
	@go install github.com/goreleaser/goreleaser/v2@latest

#
# Test
#

.PHONY: test
test: require-container-server
	@echo "Running tests..."
	@$(EXEC) go test ./... -v -p 1 -coverprofile coverage.out

.PHONY: coverage
coverage: require-container-server
	@go tool cover -html=coverage.out

#
# Lint & format
#

.PHONY: fmt
fmt: require-container-server
	@echo "Formatting code..."
	@go fmt ./...

.PHONY: lint
lint: require-container-server
	@echo "Linting code..."
	@golangci-lint run

#
# Code generation
#

.PHONY: generate
generate: generate/api generate/cmd generate/client generate/mocks

.PHONY: generate/api
generate/api: require-container-server
	@echo "Generating API..."
	@$(EXEC) go run ./internal/codegen/*.go -type api > internal/modules/api/generated.go
	@$(MAKE) fmt

.PHONY: generate/cmd
generate/cmd: require-container-server
	@echo "Generating CLI..."
	@$(EXEC) go run ./internal/codegen/*.go -type cmd > internal/cmd/generated.go
	@$(MAKE) fmt

.PHONY: generate/client
generate/client: require-container-server
	@echo "Generating API client..."
	@$(EXEC) go run ./internal/codegen/*.go -type apiclient > internal/modules/api/apiclient/generated.go
	@$(MAKE) fmt

.PHONY: generate/mocks
generate/mocks: require-container-server
	@echo "Generating mocks..."
	@$(EXEC) mockery \
		--dir internal/actions \
		--output internal/actions/mock \
		--outpkg actions_mock \
		--name Actions ;

#
# Migrations
#

DB_MIGRATIONS_DIR := internal/database/migrations
override MIGRATE := $(LOCAL_BIN)/migrate -path $(DB_MIGRATIONS_DIR) -database $(DB_DSN)

.PHONY: migrations/create
migrations/create: guard-NAME
	@$(LOCAL_BIN)/migrate create -ext sql -dir $(DB_MIGRATIONS_DIR) -seq $(NAME) ;

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

define INFO
	@echo -e "\033[1m$(1)\033[0m"
endef

guard-%:
	@if [ "${${*}}" = "" ]; then \
		echo "ERROR: Environment variable \"$*\" not set"; \
		exit 1; \
	fi

require-container-%:
	@if [ -f /.dockerenv ]; then \
		exit 0; \
	fi; \
	if [ -z "$$($(COMPOSE) ps -q $* --status running 2>/dev/null)" ]; then \
		echo "ERROR: Container \"$*\" is not running"; \
		exit 1; \
	fi
