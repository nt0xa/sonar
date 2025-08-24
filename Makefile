SHELL := /bin/bash

# Directories and paths
LOCAL_BIN := $(CURDIR)/.bin
BUILD_DIR := $(CURDIR)/build
COMPLETIONS_DIR := $(CURDIR)/completions
DOCS_DIR := $(CURDIR)/docs

# Build targets
SERVER_BIN := server
CLIENT_BIN := sonar

# Configuration
DB_DSN := postgres://db:db@localhost:5432/db_test?sslmode=disable

#
# Build
#

default: build

.PHONY: build
build: build/server build/client

.PHONY: build/server
build/server:
	$(call INFO,"Building server...")
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(SERVER_BIN) ./cmd/server

.PHONY: build/client
build/client:
	$(call INFO,"Building client...")
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(CLIENT_BIN) ./cmd/client

clean:
	$(call INFO,"Cleaning build artifacts...")
	@rm -rf $(BUILD_DIR) $(COMPLETIONS_DIR) coverage.out

#
# Shell completions
#

.PHONY: completions
completions: build/client
	$(call INFO,"Generating shell completions...")
	@rm -rf $(COMPLETIONS_DIR)
	@mkdir -p $(COMPLETIONS_DIR)
	@$(BUILD_DIR)/$(CLIENT_BIN) completion bash > $(COMPLETIONS_DIR)/$(CLIENT_BIN).bash
	@$(BUILD_DIR)/$(CLIENT_BIN) completion zsh > $(COMPLETIONS_DIR)/$(CLIENT_BIN).zsh
	@$(BUILD_DIR)/$(CLIENT_BIN) completion fish > $(COMPLETIONS_DIR)/$(CLIENT_BIN).fish

#
# Release
#

.PHONY: release
release:
	@$(LOCAL_BIN)/goreleaser release --clean

.PHONY: release/snapshot
release/snapshot:
	@$(LOCAL_BIN)/goreleaser release --clean --snapshot

#
# Dev
#

.PHONY: dev/server
dev/server:
	@$(LOCAL_BIN)/air \
		-build.bin ./server \
		-build.cmd "make build/server" \
		-build.exclude_dir docs \
		-misc.clean_on_exit true

.PHONY: dev/client
dev/client:
	@$(LOCAL_BIN)/air \
		-build.bin "" \
		-build.cmd "make build/client" \
		-build.exclude_dir docs \
		-misc.clean_on_exit true

.PHONY: dev/docs
dev/docs:
	cd $(DOCS_DIR) && npm start


#
# Tools
#

override GOBIN := $(LOCAL_BIN)
export GOBIN

.PHONY: tools
tools: 
	$(call INFO,"Installing development tools...")
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/air-verse/air@latest
	@go install github.com/abice/go-enum@latest
	@go install github.com/vektra/mockery/v2@latest 
	@go install github.com/goreleaser/goreleaser/v2@latest

#
# Test
#

.PHONY: test
test:
	$(call INFO,"Running tests...")
	@go test ./... -v -p 1 -coverprofile coverage.out

.PHONY: coverage/html
coverage/html:
	@go tool cover -html=coverage.out

#
# Lint & format
#

.PHONY: fmt
fmt:
	$(call INFO,"Formatting code...")
	@go fmt ./...

.PHONY: lint
lint:
	$(call INFO,"Linting code...")
	@golangci-lint run

#
# Code generation
#

.PHONY: generate
generate: generate/api generate/cmd generate/client generate/mocks

.PHONY: generate/api
generate/api:
	$(call INFO,"Generating API...")
	@go run ./internal/codegen/*.go -type api > internal/modules/api/generated.go
	@$(MAKE) fmt

.PHONY: generate/cmd
generate/cmd:
	$(call INFO,"Generating CLI...")
	@go run ./internal/codegen/*.go -type cmd > internal/cmd/generated.go
	@$(MAKE) fmt

.PHONY: generate/client
generate/client:
	$(call INFO,"Generating API client...")
	@go run ./internal/codegen/*.go -type apiclient > internal/modules/api/apiclient/generated.go
	@$(MAKE) fmt

.PHONY: generate/mocks
generate/mocks:
	$(call INFO,"Generating mocks...")
	@$(LOCAL_BIN)/mockery \
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
		echo "Environment variable $* not set"; \
		exit 1; \
	fi
