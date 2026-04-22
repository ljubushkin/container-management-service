ifneq (,$(wildcard .env))
include .env
export
endif

APP_MAIN=./cmd/app/main.go
MIGRATIONS_DIR=./migrations


.PHONY: docker-up
docker-up:
	docker compose up -d


.PHONY: docker-down
docker-down:
	docker compose down


.PHONY: docker-down-v
docker-down-v:
	docker compose down -v


.PHONY: docker-logs
docker-logs:
	docker compose logs -f


.PHONY: goose-up
goose-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(POSTGRES_DSN)" up


.PHONY: goose-down
goose-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(POSTGRES_DSN)" down


.PHONY: goose-status
goose-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(POSTGRES_DSN)" status


.PHONY: goose-create
goose-create:
	goose -dir $(MIGRATIONS_DIR) create $(name) sql


.PHONY: run
run:
	go run $(APP_MAIN)


.PHONY: run-inmemory
run-inmemory:
	STORAGE=inmemory go run $(APP_MAIN)


.PHONY: test
test:
	go test ./...


.PHONY: coverage
coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out


.PHONY: run-all
run-all: docker-up goose-up run