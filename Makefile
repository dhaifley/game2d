VERSION="0.1.1"

GO_FILES := $(shell find . -name "*.go")

YAML_FILES := $(shell find ./api -name "*.yaml")

-include ./.env

all: build

clean:
	rm -f game2d
	rm -f game2d-api
	rm -rf app/dist
.PHONY: clean

static/openapi.yaml: $(YAML_FILES)
	@./api/generate.sh

docs: static/openapi.yaml
.PHONY: docs

game2d: $(GO_FILES)
	CGO_ENABLED=1 go build -v -o game2d \
	-ldflags="-X github.com/dhaifley/game2d/client.Version=${VERSION}" \
	./cmd/game2d

app/dist/index.html: app/src/* app/index.html
	cd app && \
	npm run build && \
	cd ..

game2d-app: app/dist/index.html
.PHONY: game2d-app	

game2d-api: $(GO_FILES) static/* game2d-app
	CGO_ENABLED=0 go build -v -o game2d-api \
	-ldflags="-X github.com/dhaifley/game2d/server.Version=${VERSION}" \
	./cmd/game2d-api

build: game2d game2d-api
.PHONY: build

docker.test: game2d-api Dockerfile tests/docker-compose.yml
	docker compose -f tests/docker-compose.yml build
	touch docker.test

build-docker: docker.test
.PHONY: build-docker

certs/tls.key:
	@sh certs/generate.sh

certs/tls.crt: certs/tls.key

build-certs: certs/tls.crt
.PHONY: build-certs

clean-certs:
	@rm -f certs/*.crt certs/*.key certs/*.csr certs/*.srl
.PHONY: clean-certs

start.test: build-certs
	docker compose -f tests/docker-compose.yml up -d --force-recreate
	@touch start.test
	@echo "Test services started."

start: start.test
.PHONY: start

stop:
	docker compose -f tests/docker-compose.yml down --remove-orphans --volumes
	@rm -f start.test
	@echo "All test services stopped."
.PHONY: stop

test:
	@make start
	@echo "set -a && . ./.env && go test -race -cover ./..." | ${SHELL}
	@make stop
.PHONY: test

test-no-cache:
	@make start
	@echo "set -a && . ./.env && go test -count=1 -race -cover ./..." | ${SHELL}
	@make stop
.PHONY: test-no-cache

test-quick:
	go test -race -cover -short ./...
.PHONY: test-quick

test-quick-no-cache:
	go test -count=1 -race -cover -short ./...
.PHONY: test-quick-no-cache

run: build start
	@echo "set -a && . ./.env && ./game2d-api" | ${SHELL}
.PHONY: run
