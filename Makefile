#
# Build
#

.PHONY: build
build: build-server

.PHONY: build-server
build-server:
	@go build -o server ./cmd/server


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

.PHONY: mocks
mock: mock-server mock-actions

.PHONY: mock-server
mock-server:
	@mockery -dir pkg/server \
		-output pkg/server/mocks \
		-outpkg server_mocks \
		-name RequestNotifier

.PHONY: mock-actions
mock-actions:
	@mockery -dir internal/actions \
		-output internal/actions/mock \
		-outpkg actions_mock \
		-name Actions

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
