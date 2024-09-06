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
dev: install-builddeps install-builddeps-dev db
	@DEBUG=true air

.PHONY: install-builddeps
install-builddeps:
ifndef SQLC_TEST
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@${SQLC_VERSION}
endif

.PHONY: install-builddeps-dev
install-builddeps-dev:
ifndef AIR_TEST
	go install github.com/cosmtrek/air@${AIR_VERSION}
endif
ifndef GOOSE_TEST
	go install github.com/pressly/goose/v3/cmd/goose@${GOOSE_VERSION}
endif

.PHONY: build
build: install-builddeps clean db vet
	go build -tags "debug" -o ${BIN} ./cmd/bot/

.PHONY: release
release: install-builddeps clean db
	go build -a -tags "release" -ldflags "-s -w" -o ${BIN} ./cmd/bot

.PHONY: release-docker
release-docker: install-builddeps clean db
	go build -a -tags "release" \
		-ldflags '-s -w -linkmode external -extldflags "-static"' \
		-o ${BIN} ./cmd/bot

.PHONY: db
db:
	sqlc generate

.PHONY: clean
clean:
	@rm ${BIN} 2> /dev/null || true

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
