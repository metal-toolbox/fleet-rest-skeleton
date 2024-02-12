export DOCKER_BUILDKIT=1
SERVICE_NAME ?= fleet-rest-skeleton
GIT_COMMIT  := $(shell git rev-parse --short HEAD)
GIT_BRANCH  := $(shell git symbolic-ref -q --short HEAD)
GIT_SUMMARY := $(shell git describe --tags --dirty --always)
VERSION     := $(shell git describe --tags 2> /dev/null)
BUILD_DATE  := $(shell date +%s)
GO_VERSION := $(shell expr `go version |cut -d ' ' -f3 |cut -d. -f2` \>= 20)
LDFLAG_LOCATION := github.com/metal-toolbox/${SERVICE_NAME}/internal/version
DOCKER_IMAGE  ?= ghcr.io/metal-toolbox/${SERVICE_NAME}

.DEFAULT_GOAL := help

## lint
lint:
	golangci-lint run --config .golangci.yml

## Go test
test: lint
	CGO_ENABLED=0 go test -timeout 1m -v -covermode=atomic ./...

build: lint
	CGO_ENABLED=0 go build -o ${SERVICE_NAME} 

clean:
	rm -rf ${SERVICE_NAME}

image:
	docker build -t ${DOCKER_IMAGE}:latest . \
		--build-arg APP_NAME=${SERVICE_NAME} \
		--build-arg LDFLAG_LOCATION=${LDFLAG_LOCATION} \
		--build-arg GIT_COMMIT=${GIT_COMMIT} --build-arg GIT_BRANCH=${GIT_BRANCH} \
		--build-arg GIT_SUMMARY=${GIT_SUMMARY} --build-arg VERSION=${VERSION} \
		--build-arg BUILD_DATE=${BUILD_DATE} 

push-image-devel: image
	docker tag ${DOCKER_IMAGE}:latest localhost:5001/${SERVICE_NAME}:latest
	docker push localhost:5001/${SERVICE_NAME}:latest
	kind load docker-image localhost:5001/${SERVICE_NAME}:latest

push-image: image
	docker push ${DOCKER_IMAGE}:latest

# https://gist.github.com/prwhite/8168133
# COLORS
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)


TARGET_MAX_CHAR_NUM=20
## Show help
help:
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  ${YELLOW}%-$(TARGET_MAX_CHAR_NUM)s${RESET} ${GREEN}%s${RESET}\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)
