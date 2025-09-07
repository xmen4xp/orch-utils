# SPDX-FileCopyrightText: 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

### configuration ###
# safe shell invocations
SHELL           := bash -eu -o pipefail

# default goal to show help
.DEFAULT_GOAL   := help

# default configuration variables
VERSION         ?= $(shell cat VERSION)
ARCH            ?= $(shell go env GOARCH)
DOCKER_REGISTRY ?=
CLI_DIR ?= cli
DOCKER_BUILD_MOUNT_DIR ?= /go/src

# Detect platform and populate env variables.
ifeq ($(shell uname), Darwin)
	GOOS=darwin
else
	GOOS=linux
endif

# Detect GOARCH.
GOARCH=$(shell go env GOARCH)

### Nexus API GW build targets ###
.PHONY: gw.docker
gw.docker:
	@echo "Building nexus/cli for darwin..."
	docker build -f api-gw/Dockerfile -t api-gw:latest .

### Nexus CLI build targets ###
.PHONY: cli.build.darwin
cli.build.darwin:
	@echo "Building nexus/cli for darwin..."
	docker run \
		--pull=missing \
		--volume $(realpath .):${DOCKER_BUILD_MOUNT_DIR} \
		--volume ${HOME}/.cache/go-build:/root/.cache/go-build \
		-e GOCACHE=/root/.cache/go-build \
		-w ${DOCKER_BUILD_MOUNT_DIR}/${CLI_DIR} \
		golang:1.19.8 \
		/bin/bash -c "git config --global --add safe.directory '*' && go mod download && make build.darwin";

.PHONY: cli.build.linux
cli.build.linux:
	@echo "Building nexus/cli for linux..."
	docker run \
		--pull=missing \
		--volume $(realpath .):${DOCKER_BUILD_MOUNT_DIR} \
		-w ${DOCKER_BUILD_MOUNT_DIR}/${CLI_DIR} \
		--volume ${HOME}/.cache/go-build:/root/.cache/go-build \
		-e GOCACHE=/root/.cache/go-build \
		golang:1.19.8 \
		/bin/bash -c "git config --global --add safe.directory '*' && go mod download && make build.linux";

.PHONY: cli.build
cli.build:
	@echo "Building nexus/cli for $(GOOS)..."
	@if [ "$(GOOS)" = "darwin" ]; then \
		make cli.build.darwin; \
	elif [ "$(GOOS)" = "linux" ]; then \
		make cli.build.linux; \
	else \
		echo "Unknown GOOS: $(GOOS)"; \
		exit 1; \
	fi

.PHONY: cli.install.darwin
cli.install.darwin:
	cd cli; make install.darwin

.PHONY: cli.install.linux
cli.install.linux:
	cd cli; make install.linux

.PHONY: cli.install
cli.install:
	@echo "Installing nexus/cli for $(GOOS)..."
	@if [ "$(GOOS)" = "darwin" ]; then \
		make cli.install.darwin; \
	elif [ "$(GOOS)" = "linux" ]; then \
		make cli.install.linux; \
	else \
		echo "Unknown GOOS: $(GOOS)"; \
		exit 1; \
	fi

.PHONY: cli.uninstall
cli.uninstall:
	cd cli; make uninstall

### Docker build targets ###
.PHONY: compiler.builder
compiler.builder:
	cd compiler; DOCKER_REGISTRY=${DOCKER_REGISTRY} BUILDER_VERSION=${VERSION} make docker.builder

.PHONY: compiler.builder.publish
compiler.builder.publish:
	cd compiler; DOCKER_REGISTRY=${DOCKER_REGISTRY} BUILDER_VERSION=${VERSION} make docker.builder.publish

.PHONY: compiler.builder.tag
compiler.builder.tag:
	cd compiler; DOCKER_REGISTRY=${DOCKER_REGISTRY} BUILDER_VERSION=${VERSION} CUSTOM_BUILDER_DOCKER_TAG=${CUSTOM_BUILDER_DOCKER_TAG} make docker.builder.tag

.PHONY: compiler.build
compiler.build:
	cd compiler; DOCKER_REGISTRY=${DOCKER_REGISTRY} BUILDER_VERSION=${VERSION} VERSION=${VERSION} make docker

.PHONY: compiler.build.publish
compiler.build.publish:
	cd compiler; DOCKER_REGISTRY=${DOCKER_REGISTRY} BUILDER_VERSION=${VERSION} VERSION=${VERSION} make docker.publish

.PHONY: compiler.tag
compiler.tag:
	cd compiler; DOCKER_REGISTRY=${DOCKER_REGISTRY} BUILDER_VERSION=${VERSION} CUSTOM_COMPILER_DOCKER_TAG=${CUSTOM_COMPILER_DOCKER_TAG} make docker.tag

.PHONY: compiler.clean
compiler.clean:
	cd compiler; DOCKER_REGISTRY=${DOCKER_REGISTRY} BUILDER_VERSION=${VERSION} VERSION=${VERSION} make docker.clean
	cd compiler; DOCKER_REGISTRY=${DOCKER_REGISTRY} BUILDER_VERSION=${VERSION} VERSION=${VERSION} make docker.builder.clean

.PHONY: compiler.test
compiler.test:
	cd compiler; make generate_example_in_container

OPENAPI_GENERATOR_COMPONENT_NAME ?= openapi-generator
OPENAPI_GENERATOR_DOCKER_IMAGE   ?= ${DOCKER_REGISTRY}nexus/${OPENAPI_GENERATOR_COMPONENT_NAME}:${VERSION}

.PHONY: openapi.generator.docker
openapi.generator.docker:
	docker build --load -t ${OPENAPI_GENERATOR_DOCKER_IMAGE} -f openapi-generator/Dockerfile .

.PHONY: ginkgo
## Note: should fix/enable these tests
ginkgo: ## run ginkgo tests
	ginkgo run -v -r -p --race -randomize-all -randomize-suites \
  --skip-package="compiler/example/tests,compiler/pkg/generator,compiler/pkg/openapi_generator,compiler/pkg/parser,compiler/pkg/parser/rest,compiler/pkg/preparser,gqlgen/api,gqlgen/codegen/templates,gqlgen/codegen/testserver/followschema,gqlgen/codegen/testserver/singlefile,gqlgen/internal/code,gqlgen/internal/rewrite,gqlgen/graphql/handler/transport,gqlgen/plugin/federation,gqlgen/plugin/modelgen,gqlgen/plugin/resolvergen,kube-openapi/pkg/idl,kube-openapi/pkg/util/proto,kube-openapi/pkg/util/proto/validation,kube-openapi/test/integration,install-validator/pkg/dir"

compiler-docker-list:
	@make -C compiler docker-list

docker-list: compiler-docker-list ## Print names of docker container images
	@echo "  nexus/$(OPENAPI_GENERATOR_COMPONENT_NAME):"
	@echo "    name: '$(OPENAPI_GENERATOR_DOCKER_IMAGE)'"
	@echo "    version: '$(VERSION)'"
	@echo "    gitTagPrefix: 'nexus/$(VERSION)'"
	@echo "    buildTarget: 'docker-build-nexus/openapi-generator'"

#### Help Target ####
help: ## Print help for each target
	@echo $(PROJECT_NAME) make targets
	@echo "Target               Makefile:Line    Description"
	@echo "-------------------- ---------------- -----------------------------------------"
	@grep -H -n '^[[:alnum:]_-]*:.* ##' $(MAKEFILE_LIST) \
    | sort -t ":" -k 3 \
    | awk 'BEGIN  {FS=":"}; {sub(".* ## ", "", $$4)}; {printf "%-20s %-16s %s\n", $$3, $$1 ":" $$2, $$4};'

# K8s Runtime Commands

.PHONY: kubectl.build
	kubectl.build:
		@if [ -z "$$(docker images -q nexus/kubectl:latest)" ]; then \
			echo "Building nexus/kubectl:latest image..."; \
			cd nexus-runtime-manifests/Dockerfiles && docker build -t nexus/kubectl:latest -f Dockerfile.kubectl .; \
		else \
			echo "nexus/kubectl:latest image already exists."; \
		fi

.PHONY: kind.install
kind.install: kubectl.build
	cd nexus-runtime-manifests/kind && ./run_kind.sh -n ${CLUSTER_NAME} -p ${CLUSTER_PORT}

.PHONY: kind.uninstall
kind.uninstall:
	cd nexus-runtime-manifests/kind && ./run_kind.sh -n ${CLUSTER_NAME} -d

.PHONY: check.kind
check.kind:
	$(info Checking if kind requirements are met...)
	@kind version > /dev/null

# Nexus Runtime Commands

.PHONY: runtime.uninstall.kind
runtime.uninstall.kind: check.kind kind.uninstall api-gw.kind.stop
	$(info )
	$(info ====================================================)
	$(info Runtime is now uninstalled)
	$(info ====================================================)

.PHONY: container-build
container-build:
	cd nexus-runtime-manifests && CLUSTER_NAME=${CLUSTER_NAME} make container-build
	cd api-gw && CLUSTER_NAME=${CLUSTER_NAME} make container-build
	cd api && CLUSTER_NAME=${CLUSTER_NAME} make container-build

.PHONY: runtime-kind-load-images
runtime-kind-load-images:
	$(info Loading images into kind cluster...)
	cd nexus-runtime-manifests && CLUSTER_NAME=${CLUSTER_NAME} make kind-load-images
	cd api-gw && CLUSTER_NAME=${CLUSTER_NAME} make kind-load-images
	cd api && CLUSTER_NAME=${CLUSTER_NAME} make kind-load-images

.PHONY: runtime.install.kind
runtime.install.kind: HOST_KUBECONFIG=$(realpath .)/nexus-runtime-manifests/kind/.${CLUSTER_NAME}/kubeconfig
runtime.install.kind: check.kind kind.install api.install.kind api-gw.kind.run
	$(info )
	$(info ====================================================)
	$(info To access runtime, you can execute kubectl as:)
	$(info     kubectl -s localhost:${K8S_RUNTIME_PORT} ...)
	$(info )
	$(info To access nexus api gateway using kubeconfig, export:)
	$(info     export HOST_KUBECONFIG=${HOST_KUBECONFIG})
	$(info     export DOCKER_NETWORK=kind)
	$(info )
	$(info ====================================================)
