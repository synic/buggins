GOPATH?=`realpath workspace`
BIN=./bin/bot
CGO_ENABLED=1

# package versions
AIR_VERSION=v1.49.0
SQLC_VERSION=v1.27.0
GOOSE_VERSION=v3.21.1

AIR_TEST := $(shell command -v air 2> /dev/null)
SQLC_TEST := $(shell command -v sqlc 2> /dev/null)
GOOSE_TEST := $(shell command -v goose 2> /dev/null)

.PHONY: dev
dev: install-builddeps
	@DEBUG=true air

.PHONY: install-devdeps
install-devdeps:
ifndef AIR_TEST
	go install github.com/cosmtrek/air@${AIR_VERSION}
endif
ifndef GOOSE_TEST
	go install github.com/pressly/goose/v3/cmd/goose@${GOOSE_VERSION}
endif
ifndef SQLC_TEST
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@${SQLC_VERSION}
endif

.PHONY: build
build:
	go build -tags "debug" -o ${BIN} ./

.PHONY: release
release:
	go build -a -tags "release" -ldflags "-s -w" -o ${BIN} ./

.PHONY: release-docker
release-docker:
	go build -a -tags "release" \
		-ldflags '-s -w -linkmode external -extldflags "-static"' -o ${BIN} .

.PHONY: codegen
codegen:
	go generate ./...

.PHONY: test
test:
	go test -race ./... | tc

.PHONY: vet
vet:
	go vet ./...

.PHONY: build-container
build-container:
	docker build . -t "buggins-bot"

.PHONY: migrate
migrate:
	./scripts/migrate.sh

.PHONY: migration-create
migration-create:
	./scripts/migration-create.sh
