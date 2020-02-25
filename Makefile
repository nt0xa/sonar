project = sonar
owner = bi-zone
build = $(shell git rev-parse --short HEAD)
version = $(shell git describe --tags | cut -c 2-)
version_patch = ${version}
version_minor = $(shell echo ${version} | cut -d '.' -f1-2)
version_major = $(shell echo ${version} | cut -d '.' -f1)
go_version = 1.13

#
# Build
#

.PHONY: build
build: build-server

.PHONY: build-server
build-server:
	@go build -ldflags '-s -w' -o server ./cmd/server

#
# Docker
#

docker_registry = docker.pkg.github.com
docker_image = ${docker_registry}/${owner}/${project}/server
docker_image_patch = ${docker_image}:${version_patch}
docker_image_minor = ${docker_image}:${version_minor}
docker_image_major = ${docker_image}:${version_major}

.PHONY: docker-login
docker-login:
	@docker login --username ${docker_login} --password ${docker_password} ${docker_registry}

.PHONY: docker-build
docker-build:
	@docker build --build-arg go_version=${go_version} --tag ${docker_image} .

.PHONY: docker-push
	@docker tag ${docker_image} ${docker_image_patch}
	@docker tag ${docker_image} ${docker_image_minor}
	@docker tag ${docker_image} ${docker_image_major}
	@docker push ${docker_image}

.PHONY: docker-release
docker-release: docker-login docker-build docker-push

#
# Test
#

.PHONY: test
test:
	@go test ./... -v -coverprofile coverage.out

.PHONY: coverage
coverage:
	@go tool cover -func=coverage.out

.PHONY: coverage-html
coverage-html:
	@go tool cover -html=coverage.out

#
# Migrations
#

migrations = ./internal/database/migrations
db_url = ${SONAR_DB}

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

.PHONY: migrations-pack
migrations-pack:
	@go-bindata -pkg migrations \
		-ignore ".*\.go" \
		-prefix ${migrations} \
		-o ${migrations}/bindata.go \
		${migrations}/...
	@go fmt ${migrations}

#
# Changelog
#

.PHONY: changelog
changelog:
	@git-chglog --config .github/git-chglog.yml --output CHANGELOG.md v${version}

