GOPATH?=`realpath workspace`
BIN="./bin/bot"

# package versions
AIR_VERSION=v1.49.0
SQLC_VERSION=v1.27.0
MIGRATE_VERSION=v4.17.1

AIR_TEST := $(shell command -v air 2> /dev/null)
SQLC_TEST := $(shell command -v sqlc 2> /dev/null)
MIGRATE_TEST := $(shell command -v migrate 2> /dev/null)

.PHONY: dev
dev: install-builddeps db
ifndef AIR_TEST
	go install github.com/cosmtrek/air@${AIR_VERSION}
endif
	@air \
	  -root "." \
		-tmp_dir ".tmp" \
		-build.cmd "make build" \
		-build.bin "DEBUG=true ./bin/bot" \
		-build.delay "1000" \
		-build.exclude_dir 'logs,node_modules,bin' \
		-build.exclude_file \
		  'Dockerfile,docker-compose.yaml,internal/db/queries.sql.go,internal/db/models.go,internal/db/db.go' \
		-build.exclude_regex '_test.go,.null-ls' \
		-build.include_ext 'go,md,yaml,sql' \
		-build.log "logs/build-errors.log" \
		-misc.clean_on_exit "false"

.PHONY: install-builddeps
install-builddeps:
ifndef SQLC_TEST
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@${SQLC_VERSION}
endif
ifndef AIR_TEST
	go install github.com/cosmtrek/air@${AIR_VERSION}
endif
ifndef MIGRATE_TEST
	CGO_ENABLED=0 go install --tags "sqlite" github.com/golang-migrate/migrate/v4/cmd/migrate@${MIGRATE_VERSION}
endif

.PHONY: build
build: install-builddeps clean db vet
	@CGO_ENABLED=0 go build -tags debug -o ${BIN}

.PHONY: release
release: install-builddeps clean db
	@CGO_ENABLED=0 go build -tags release -ldflags "-s -w" -o ${BIN} .

.PHONY: db
db:
	@sqlc generate

.PHONY: clean
clean:
	@rm ${BIN} 2> /dev/null || true

.PHONY: test
test:
	@go test -race ./... | tc

.PHONY: vet
vet:
	@go vet ./...

.PHONY: build-container
build-container:
	docker build . -t "buggins-bot"
