#
# Build
#

.PHONY: build
build: build-server build-client

.PHONY: build-server
build-server:
	@go build -o server ./cmd/server

.PHONY: build-client
build-client:
	@go build -o sonar ./cmd/client


#
# Test
#

.PHONY: test
test:
	@go test ./... -v -p 1 -coverprofile coverage.out

.PHONY: coverage
coverage:
	@go tool cover -func=coverage.out

.PHONY: coverage-html
coverage-html:
	@go tool cover -html=coverage.out

.PHONY: mock
mock: mock-actions

.PHONY: mock-actions
mock-actions:
	@mockery \
		--dir internal/actions \
		--output internal/actions/mock \
		--outpkg actions_mock \
		--name Actions
	@mockery \
		--dir internal/actions \
		--output internal/actions/mock \
		--outpkg actions_mock \
		--name ResultHandler

.PHONY: mock-deps
	@go install github.com/vektra/mockery/v2@latest

#
# Code generation
#

.PHONY: gen
gen: gen-api gen-cmd gen-apiclient

.PHONY: gen-api
gen-api:
	@go run ./internal/codegen/*.go -type api > internal/modules/api/generated.go
	@go fmt ./internal/modules/api

.PHONY: gen-cmd
gen-cmd:
	@go run ./internal/codegen/*.go -type cmd > internal/cmd/generated.go
	@go fmt ./internal/cmd

.PHONY: gen-apiclient
gen-apiclient:
	@go run ./internal/codegen/*.go -type apiclient > internal/modules/api/apiclient/generated.go
	@go fmt ./internal/modules/api/apiclient

#
# Migrations
#

migrations = ./internal/database/migrations
db_url = ${SONAR_DB_DSN}

.PHONY: migrations-create
migrations-create:
	@migrate create -ext sql -dir ${migrations} -seq ${name}

.PHONY: migrations-list
migrations-list:
	@ls ${migrations} | grep '.sql' | cut -d '.' -f 1 | sort | uniq

.PHONY: migrations-up
migrations-up:
	@migrate -path ${migrations} -database ${db_url} up ${n}

.PHONY: migrations-down
migrations-down:
	@migrate -path ${migrations} -database ${db_url} down ${n}

.PHONY: migrations-goto
migrations-goto:
	@migrate -path ${migrations} -database ${db_url} goto ${v}

.PHONY: migrations-drop
migrations-drop:
	@migrate -path ${migrations} -database ${db_url} drop

.PHONY: migrations-force
migrations-force:
	@migrate -path ${migrations} -database ${db_url} force ${v}
