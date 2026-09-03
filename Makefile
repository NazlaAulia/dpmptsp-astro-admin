.DEFAULT_GOAL := help
SHELL := /bin/bash

# A leading dash so a fresh clone with no .env still runs `make help` instead of
# failing on the include.
-include .env
export

COMPOSE     := docker compose
COMPOSE_DEV := docker compose -f docker-compose.yml -f docker-compose.dev.yml

DB_CONNECTION ?= mysql
DB_DATABASE   ?= ladpm
DB_USERNAME   ?= root
# migrate/migrate needs the engine name for the migrations directory.
DB_ENGINE     := $(DB_CONNECTION)

# Go and golang-migrate are deliberately NOT installed on the host. Both live in
# containers, so every target here works on a machine with only docker.
GO_IMAGE      := golang:1.25-alpine
MIGRATE_IMAGE := migrate/migrate:v4.19.0
NETWORK       := dpmptsp_backend

ifeq ($(DB_CONNECTION),postgres)
  DB_URL := postgres://$(DB_USERNAME):$(DB_PASSWORD)@database:5432/$(DB_DATABASE)?sslmode=disable
else
  DB_URL := mysql://$(DB_USERNAME):$(DB_PASSWORD)@tcp(database:3306)/$(DB_DATABASE)
endif

HAS_API := $(shell test -f apps/api/go.mod && echo yes)

# --user keeps generated files owned by you. Without it the container writes
# root-owned migrations into the working tree and you cannot edit or stage them.
DOCKER_RUN_GO := docker run --rm \
  -v "$(PWD)/apps/api":/src -w /src \
  --user "$(shell id -u):$(shell id -g)" \
  -e HOME=/tmp -e GOCACHE=/tmp/gocache -e GOMODCACHE=/tmp/gomod \
  $(GO_IMAGE)

MIGRATE := docker run --rm \
  --network $(NETWORK) \
  --user "$(shell id -u):$(shell id -g)" \
  -v "$(PWD)/apps/api/migrations/$(DB_ENGINE)":/migrations \
  $(MIGRATE_IMAGE) -path=/migrations -database "$(DB_URL)"

.PHONY: help dev up down logs ps build test lint fmt \
        migrate-up migrate-down migrate-create migrate-version migrate-fresh \
        seed seed-list seed-only seed-fresh schema-check smoke clean \
        legacy-media legacy-media-archive

help: ## Show this help
	@grep -hE '^[a-z-]+:.*?##' $(MAKEFILE_LIST) \
	  | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-16s\033[0m %s\n",$$1,$$2}'

dev: ## Run the whole stack with hot reload
	$(COMPOSE_DEV) up --build

up: ## Run the production-shaped stack detached
	$(COMPOSE) up -d --build

down: ## Stop everything
	$(COMPOSE) down

logs: ## Tail all logs
	$(COMPOSE) logs -f --tail=100

ps: ## Show service status
	$(COMPOSE) ps

build: ## Build every workspace and the api image
	pnpm -r build
ifeq ($(HAS_API),yes)
	$(COMPOSE) build api
endif

test: ## Run all tests
	pnpm -r test
ifeq ($(HAS_API),yes)
	$(DOCKER_RUN_GO) go test ./...
endif

lint: ## Typecheck and vet
	pnpm -r lint
ifeq ($(HAS_API),yes)
	$(DOCKER_RUN_GO) sh -c 'gofmt -l . && go vet ./...'
endif

fmt: ## Format Go sources
ifeq ($(HAS_API),yes)
	$(DOCKER_RUN_GO) gofmt -w .
endif

migrate-up: ## Apply all pending migrations
	$(MIGRATE) up

# `down 1`, not bare `down`. Bare down reverts EVERYTHING and prompts
# interactively, which is far too easy to blow past from a Makefile.
migrate-down: ## Roll back exactly one migration
	$(MIGRATE) down 1

migrate-create: ## Create a migration: make migrate-create name=add_users
	@test -n "$(name)" || { echo "usage: make migrate-create name=xxx"; exit 1; }
	@for engine in postgres mysql; do \
	  docker run --rm --user "$(shell id -u):$(shell id -g)" \
	    -v "$(PWD)/apps/api/migrations/$$engine":/migrations \
	    $(MIGRATE_IMAGE) create -ext sql -dir /migrations -seq $(name); \
	done
	@echo "Created in both migrations/postgres and migrations/mysql."
	@echo "Write both. 'make schema-check' proves they agree."

# --- seeding, in the shape `php artisan db:seed` takes -------------------------
# Runs inside a container on the compose network, with the same environment the
# api service gets, so it reaches `database` by the same name.
SEED_RUN = docker run --rm \
  --network $(NETWORK) \
  -v "$(PWD)/apps/api":/src -w /src \
  --user "$(shell id -u):$(shell id -g)" \
  -e HOME=/tmp -e GOCACHE=/tmp/gocache -e GOMODCACHE=/tmp/gomod \
  -e DB_CONNECTION -e DB_HOST=database -e DB_PORT -e DB_DATABASE \
  -e DB_USERNAME -e DB_PASSWORD -e DB_SSLMODE -e DATABASE_URL \
  -e API_SERVICE_KEY -e HASH_DRIVER -e BCRYPT_ROUNDS -e SEED_ADMIN_PASSWORD \
  $(GO_IMAGE) go run ./cmd/seed

seed: ## Insert reference data. Safe to re-run: existing rows are skipped.
	@test -n "$(SEED_ADMIN_PASSWORD)" || { \
	  echo "SEED_ADMIN_PASSWORD is required — the admin password is hashed at"; \
	  echo "seed time and never stored in a file."; exit 1; }
	$(SEED_RUN)

seed-list: ## Show which seeders would run, without touching the database
	$(SEED_RUN) -list

seed-only: ## Run one seeder: make seed-only name=users
	@test -n "$(name)" || { echo "usage: make seed-only name=users"; exit 1; }
	$(SEED_RUN) -only=$(name)

seed-fresh: ## Delete the rows each seeder owns, then re-insert them
	$(SEED_RUN) -fresh

migrate-fresh: ## Drop everything, migrate from empty, then seed
	$(MIGRATE) down -all
	$(MIGRATE) up
	$(MAKE) seed

migrate-version: ## Show the current migration version
	$(MIGRATE) version

# Two migration sets exist because no single SQL file is valid on both engines.
# The failure mode is that they drift silently. This spins up a throwaway
# database of each kind, migrates both from empty, and diffs the result.
schema-check: ## Prove the postgres and mysql migrations produce the same schema
	@set -e; \
	NET=schemacheck-$$$$; \
	trap "docker rm -f sc-pg-$$$$ sc-my-$$$$ >/dev/null 2>&1; docker network rm $$NET >/dev/null 2>&1" EXIT; \
	docker network create $$NET >/dev/null; \
	docker run -d --name sc-pg-$$$$ --network $$NET \
	  -e POSTGRES_PASSWORD=p -e POSTGRES_USER=u -e POSTGRES_DB=d postgres:16-alpine >/dev/null; \
	docker run -d --name sc-my-$$$$ --network $$NET \
	  -e MYSQL_ROOT_PASSWORD=p -e MYSQL_DATABASE=d mysql:8.0 >/dev/null; \
	echo "waiting for throwaway databases..."; \
	for i in $$(seq 1 60); do docker exec sc-pg-$$$$ pg_isready -h 127.0.0.1 -U u >/dev/null 2>&1 && break; sleep 1; done; \
	for i in $$(seq 1 90); do docker exec sc-my-$$$$ mysqladmin ping -h 127.0.0.1 -uroot -pp >/dev/null 2>&1 && break; sleep 2; done; \
	docker run --rm --network $$NET -v "$(PWD)/apps/api/migrations/postgres":/m $(MIGRATE_IMAGE) \
	  -path=/m -database "postgres://u:p@sc-pg-$$$$:5432/d?sslmode=disable" up; \
	docker run --rm --network $$NET -v "$(PWD)/apps/api/migrations/mysql":/m $(MIGRATE_IMAGE) \
	  -path=/m -database "mysql://root:p@tcp(sc-my-$$$$:3306)/d" up; \
	docker run --rm --network $$NET -v "$(PWD)/apps/api":/src -w /src \
	  --user "$(shell id -u):$(shell id -g)" \
	  -e HOME=/tmp -e GOCACHE=/tmp/gocache -e GOMODCACHE=/tmp/gomod \
	  $(GO_IMAGE) go run ./cmd/schemadiff \
	    -postgres "postgres://u:p@sc-pg-$$$$:5432/d?sslmode=disable" \
	    -mysql "root:p@tcp(sc-my-$$$$:3306)/d"

legacy-media: ## Report which legacy article images are still reachable
	./scripts/legacy-media.sh check

legacy-media-archive: ## Download every reachable legacy image before the old host goes away
	./scripts/legacy-media.sh archive

smoke: ## Check the running stack, including that the api is NOT reachable from the gateway
	@set -e; \
	echo "web via gateway:   $$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:8017/ || echo down)"; \
	echo "admin via gateway: $$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:8018/admin/login || echo down)"; \
	echo -n "api reachable from web (must succeed):     "; \
	$(COMPOSE) exec -T web  wget -qO- http://api:8080/healthz >/dev/null 2>&1 && echo yes || echo NO; \
	echo -n "api reachable from gateway (must be NO):   "; \
	$(COMPOSE) exec -T gateway wget -qO- http://api:8080/healthz >/dev/null 2>&1 && echo "YES - ISOLATION BROKEN" || echo no

clean: ## Stop and remove volumes. Destroys local database contents.
	$(COMPOSE) down -v
