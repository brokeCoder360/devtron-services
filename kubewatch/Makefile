.PHONY: default build docker-image test stop clean-images clean

BINARY = kubewatch

VERSION=
BUILD=

PKG            = github.com/devtron-labs/kubewatch
TRAVIS_COMMIT ?= `git describe --tags`
GOCMD          = go
BUILD_DATE     = `date +%FT%T%z`
GOFLAGS       ?= $(GOFLAGS:)
LDFLAGS       := "-X '$(PKG)/cmd.gitCommit=$(TRAVIS_COMMIT)' \
		          -X '$(PKG)/cmd.buildDate=$(BUILD_DATE)'"

default: fetch-all-env build

build:
	CGO_ENABLED=0 GOOS=linux "$(GOCMD)" build ${GOFLAGS} -ldflags ${LDFLAGS} -o "${BINARY}"

docker-image:
	@docker build -t "${BINARY}" .

stop:
	@docker stop "${BINARY}"

clean-images: stop
	@docker rmi "${BUILDER}" "${BINARY}"

clean:
	"$(GOCMD)" clean -i


TARGET_BRANCH?=main
dep-update-oss:
	go mod edit -replace=github.com/devtron-labs/common-lib=github.com/devtron-labs/devtron-services/common-lib@$(TARGET_BRANCH)
	go mod tidy
	go mod vendor

fetch-all-env:
	go run fetchAllEnv/fetchAllEnv.go
