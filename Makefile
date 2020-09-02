DOCKER_NAME := jdlubrano/reverse-proxy
GIT_SHA := $$(git log -1 --pretty=%H)
DOCKER_IMG := ${DOCKER_NAME}:${GIT_SHA}
DOCKER_LATEST_TAG := ${DOCKER_NAME}:latest

default: test

test:
	@go test -cover ./...

build:
	@go build

format:
	@go fmt ./...

start-dev: build
	./reverse-proxy --routes-file config/routes.yml.example

docker-build: build
	@docker build -t ${DOCKER_IMG} .
	@docker tag ${DOCKER_IMG} ${DOCKER_LATEST_TAG}

docker-push: docker-build
	@docker push ${DOCKER_NAME}

modupdate: ## Bump the minor for all dependencies in the go.mod
	@go get -u ./...

.PHONY: test build format start-dev modupdate
