# SPDX-FileCopyrightText: 2025 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

SHELL := /bin/bash
#
# App Info
#
VERSION                         ?= $(shell cat VERSION)
GIT_HASH_SHORT                  := $(shell git rev-parse --abbrev-ref HEAD)
ROOT_DIR                        := $(shell git rev-parse --show-toplevel)
VERSION_DEV_SUFFIX              := ${GIT_HASH_SHORT}

# Add an identifying suffix for `-dev` builds only.
# Release build versions are verified as unique by the CI build process.
ifneq ($(GIT_HASH_SHORT), main)
	ifeq ($(findstring -dev,$(VERSION)), -dev)
		VERSION := $(VERSION)-$(VERSION_DEV_SUFFIX)
	endif
endif

API_GW_COMPONENT_NAME           ?= api-gw
DOCKER_REPO                     ?= ciscohyperdrive
OS                              ?= $(shell go env GOOS)
ARCH                            ?= $(shell go env GOARCH)
TAG                             ?= latest
API_GW_DOCKER_IMAGE             ?= ${DOCKER_REGISTRY}/${DOCKER_REPO}/nexus/${API_GW_COMPONENT_NAME}:${TAG}
BUILD_DIR                       := build/bin
GOEXTRAFLAGS                    := -buildvcs=false
GOCMD                           := CGO_ENABLED=0 GOOS=$(shell go env GOOS) GOARCH=$(shell go env GOARCH) go
GOCMD_TEST                      := GOPRIVATE="github.com/open-edge-platform/*" go

ifeq (${DOCKER_TAG}, )
ifeq (${DOCKER_REGISTRY}, )
DOCKER_TAG=${API_GW_COMPONENT_NAME}:${TAG}
else
DOCKER_TAG=${DOCKER_REGISTRY}/${DOCKER_REPO}/${API_GW_COMPONENT_NAME}:${TAG}
endif
endif

KUBECONFIG ?= ${ROOT_DIR}/nexus-runtime-manifests/kind/.test/kubeconfig

FUZZ_SECONDS                    ?= 24

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

YAML_FILES := $(shell find . -path './venv_charts' -prune -o -type f \( -name '*.yaml' -o -name '*.yml' \) -print )

# Create the virtualenv with python tools installed
VENV_NAME = venv_apigw

$(VENV_NAME): requirements.txt
	echo "Creating virtualenv $@"
	python3 -m venv $@;\
	. ./$@/bin/activate; set -u;\
	python3 -m pip install --upgrade pip;\
	python3 -m pip install -r requirements.txt

.PHONY: all
all: create

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php


.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Dev

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: test-fmt
test-fmt:
	test -z $$(goimports -w -l cmd pkg)

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: lint
lint: go-lint hadolint yamllint ## Scan for all the golang source files, Dockerfile and yaml files

.PHONY: hadolint
hadolint: ## Scan Dockerfile with Hadolint
	hadolint Dockerfile

.PHONY: go-lint
go-lint: ## Scan golang source files with golangci-lint
	golangci-lint --version
	golangci-lint run --timeout=3m --config .golangci.yml ./...

.PHONY: yamllint
yamllint: ## Lint yaml files using yamllint
	yamllint --version
	yamllint -c yamllint_conf.yml -s $(YAML_FILES)

.PHONY: go-tidy
go-tidy: ## Updates the Go dependencies and regenerate the go.sum file
	go mod tidy

.PHONY: go-vendor
go-vendor:
	go mod vendor

create:
	mkdir -p ${BUILD_DIR}

.PHONY: test
test: build race-unit-test ## Runs unit test

.PHONY: race-unit-test
race-unit-test: init-unit-test
	${GOCMD_TEST} test -coverprofile ${BUILD_DIR}/coverage.out -coverpkg=$$(go list ./... | tr '\n' ,)  -covermode atomic ./...

.PHONY: ginkgo
ginkgo: ## run ginkgo tests
	ginkgo run -v -r -p --race -randomize-all -randomize-suites

clean-testdata:
	rm -rf ./pkg/apiremap/testdata
	rm -rf ./pkg/auth/authn/testdata
	rm -rf ./pkg/auth/authz/testdata

test-fuzz: clean-testdata
	$(GOCMD) test ./pkg/apiremap -fuzz FuzzTenancyAPIRemapping -fuzztime=${FUZZ_SECONDS}s
	$(GOCMD) test ./pkg/auth/authn -fuzz FuzzVerifyJWT -fuzztime=${FUZZ_SECONDS}s
	$(GOCMD) test ./pkg/auth/authz -fuzz FuzzVerifyAuthorization -fuzztime=${FUZZ_SECONDS}s
	$(GOCMD) test ./pkg/auth/authz -fuzz FuzzVerifyAuthzForProjects -fuzztime=${FUZZ_SECONDS}s
	$(GOCMD) test ./pkg/auth/authz -fuzz FuzzAuthzForProjectsMemberRole -fuzztime=${FUZZ_SECONDS}s

##@ Coverage
.PHONY: coverage
coverage: test ## Runs unit test coverage
	gocover-cobertura < ${BUILD_DIR}/coverage.out > ${BUILD_DIR}/coverage.xml
	${GOCMD_TEST} tool cover -html=${BUILD_DIR}/coverage.out -o ${BUILD_DIR}/coverage.html
	${GOCMD_TEST} tool cover -func ${BUILD_DIR}/coverage.out -o ${BUILD_DIR}/function_coverage.log


.PHONY: image_scan
image_scan:
	flash docker scan image ${API_GW_DOCKER_IMAGE}

.PHONY: docker.run.kind
docker.run.kind:
	docker rm -f api-gw 2>/dev/null || true
	docker run -it --name ${API_GW_COMPONENT_NAME} -e KUBECONFIG=/kubeconfig -e GATEWAY_MODE=admin -v ${KUBECONFIG}/Users/dbakiara/src/github.com/xmen4xp/orch-utils/nexus/nexus-runtime-manifests/kind/.test/kubeconfig:/kubeconfig -p 8082:8082 --network kind ${API_GW_DOCKER_IMAGE}

# Standard Makefile targets

.PHONY: build
build: create ## Builds golang source files
	${GOCMD} build ${GOEXTRAFLAGS} -o ${BUILD_DIR}/${API_GW_COMPONENT_NAME}

.PHONY: clean
clean: ## Clean up the build directory
	rm -rf ${BUILD_DIR}
	rm -rf vendor

.PHONY: container-build
container-build: go-vendor
	@echo "Building container images..."
	# docker build, buildah, etc.
	docker build --build-arg API_GW_COMPONENT_NAME=${API_GW_COMPONENT_NAME} -t ${DOCKER_TAG} -f Dockerfile .

.PHONY: container-push
container-push: ## Push container images to registry
	docker push ${DOCKER_TAG}

.PHONY: kind-load-images
kind-load-images:
	$(info Loading api-gw images into kind cluster...)
	docker save ${DOCKER_TAG} | docker exec -i ${CLUSTER_NAME}-control-plane ctr --namespace k8s.io images import -
