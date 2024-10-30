# Use winpty if present
WINPTY = $(shell basename `which winpty 2> /dev/null` 2> /dev/null)

# Golang caches
LOCAL_GOCACHE = $(shell go env GOCACHE)
LOCAL_GOMODCACHE = $(shell go env GOMODCACHE)

# Docker commands
DOCKER = $(WINPTY) docker
DOCKER_COMPOSE = $(WINPTY) env LOCAL_GOCACHE=$(LOCAL_GOCACHE) LOCAL_GOMODCACHE=$(LOCAL_GOMODCACHE) docker compose

.DEFAULT_GOAL := help
help: init
	@printf " \033[33mCommands\033[0m:\n"
	@grep -E '(^[a-zA-Z0-9_-]+:.*?##.*$$)|(^##)' Makefile | awk 'BEGIN {FS = ":.*?##s*?"}; {printf "  \033[32m%-20s\033[0m %s\n", $$1, $$2}' | sed -e 's/\[32m##/[33m/'

.PHONY: help

info: ## Display the docker compose configuration
info: init
	$(DOCKER_COMPOSE) config

init: ## Create network if it doesn't exist
ifeq ($(shell docker network ls --filter "name=goauth_local_network" --quiet), )
	-$(DOCKER) network create goauth_local_network
endif

pull: ## Pull the images (this will refresh them if needed)
pull: init
	$(DOCKER_COMPOSE) pull --ignore-pull-failures

build: ## Pull and build the images
build: pull
	$(DOCKER_COMPOSE) build --pull

run: ## Run the containers... (no detach --> all logs are visible in the console)
run: init
	$(DOCKER_COMPOSE) up --remove-orphans --force-recreate

start: ## Start the containers in the background (detach the processes)
start: init
	$(DOCKER_COMPOSE) up -d --remove-orphans --force-recreate --wait goauth mocks mongodb

tests: ## Launch all the tests
tests: qa-tests unit-tests

qa-tests: ## Launch the QA tests
qa-tests:
	$(DOCKER_COMPOSE) exec goauth sh -c "cd /src && gofmt -d ."
	$(DOCKER_COMPOSE) exec goauth sh -c "cd /src && go vet ./..."
	$(DOCKER_COMPOSE) run --rm golangci

unit-tests: ## Launch the unit tests
unit-tests:
	$(DOCKER_COMPOSE) exec goauth sh -c "cd /src && go test ./..."

logs: ## Display the logs... (with the --follow option)
logs: init
	$(DOCKER_COMPOSE) logs -f goauth

stop: ## Stop the containers
stop: init
	$(DOCKER_COMPOSE) stop

kill: ## Force to stop and remove the containers and volumes
kill: init
	$(DOCKER_COMPOSE) kill
	$(DOCKER_COMPOSE) down --volumes --remove-orphans

restart: ## Stop and start the containers
restart: kill
	$(MAKE) --no-print-directory start

prune: ## Prune the Docker system (including the volumes)
prune: init
	$(DOCKER) system prune --volumes
	$(DOCKER) image prune

.PHONY: info init pull build run start logs stop kill restart prune
