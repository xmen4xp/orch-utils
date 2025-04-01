# SPDX-FileCopyrightText: 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

TAG ?= $(shell cat TAG | awk '{ print $1 }')
ARCH ?= $(shell go env GOARCH)

OPENAPI_GENERATOR_COMPONENT_NAME ?= openapi-generator
OPENAPI_GENERATOR_DOCKER_IMAGE?= ${DOCKER_REGISTRY}/nexus/${OPENAPI_GENERATOR_COMPONENT_NAME}:${TAG}

.PHONY: check.tag
check.tag:
ifndef TAG
	$(error TAG is mandatory and should be set to the desired docker image tag)
endif

.PHONY: compiler.builder
compiler.builder:
	cd compiler; DOCKER_REGISTRY=${DOCKER_REGISTRY} BUILDER_TAG=${TAG} make docker.builder

.PHONY: compiler.builder.publish
compiler.builder.publish:
	cd compiler; DOCKER_REGISTRY=${DOCKER_REGISTRY} BUILDER_TAG=${TAG} make docker.builder.publish

.PHONY: compiler.build
compiler.build:
	cd compiler; DOCKER_REGISTRY=${DOCKER_REGISTRY} BUILDER_TAG=${TAG} TAG=${TAG} make docker

.PHONY: compiler.build.publish
compiler.build.publish:
	cd compiler; DOCKER_REGISTRY=${DOCKER_REGISTRY} BUILDER_TAG=${TAG} TAG=${TAG} make docker.publish

.PHONY: openapi.generator.docker
openapi.generator.docker:
	docker build --no-cache -t ${OPENAPI_GENERATOR_DOCKER_IMAGE} -f openapi-generator/Dockerfile .
