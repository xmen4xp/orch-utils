BUILD_DIR ?= build
CRD_MODULE_PATH ?= $(shell go list -m)/${BUILD_DIR}/
COMPILER_TAG ?= latest
CONTAINER_ID ?= ""
DATAMODEL_LOCAL_PATH ?= $(realpath .)
BUCKET ?= nexus-template-downloads
DOCKER_REPO ?= $(shell cat nexus.yaml | grep dockerRepo | awk '{print $$2}' | tr -d '"')
DATAMODEL_DOCKER_REGISTRY ?= $(shell cat nexus.yaml | grep dockerRepo | awk '{print $$2}' | tr -d '"')
DATAMODEL_GROUP_NAME ?= $(shell cat nexus.yaml | grep groupName | awk '{print $$2}' | tr -d '"')
TAG ?= latest
NAME ?= $(shell cat nexus.yaml | grep groupName | awk '{print $$2}' | tr -d '"')
NEXUS_API_GW_K8S_NAMESPACE ?= "default"
DOCKER_BASE_IMAGE=gcr.io/nsx-sm/photon-kubectl:latest
DOCKER_NETWORK ?= nexus
MOUNTED_KUBECONFIG ?= /etc/config/kubeconfig

ifeq (${DOCKER_TAG}, )
ifeq (${DATAMODEL_DOCKER_REGISTRY}, )
DOCKER_TAG=${DATAMODEL_GROUP_NAME}:${TAG}
else
DOCKER_TAG=${DATAMODEL_DOCKER_REGISTRY}/${DOCKER_REPO}/${DATAMODEL_GROUP_NAME}:${TAG}
endif
endif

COMPILER_DOCKER_REGISTRY ?=
IMAGE_NAME ?= nexus/compiler
GOARCH ?= $(shell go env GOARCH)
COMPILER_NAME ?= ${IMAGE_NAME}/${GOARCH}
ifeq (${COMPILER_DOCKER_TAG}, )
ifeq (${COMPILER_DOCKER_REGISTRY}, )
COMPILER_DOCKER_TAG=${COMPILER_NAME}:${COMPILER_TAG}
else
COMPILER_DOCKER_TAG=${COMPILER_DOCKER_REGISTRY}/${COMPILER_NAME}:${COMPILER_TAG}
endif
endif

.PHONY: datamodel_build
datamodel_build:
	@echo "CRD and API Generated Output Directory: ${BUILD_DIR}"
	@echo "OPENAPISpec Generated Output Directory: ${BUILD_DIR}/crds/"
	rm -rf ${DATAMODEL_LOCAL_PATH}/build
	mkdir -p ${BUILD_DIR}
	if [ -z $(CONTAINER_ID) ]; then \
		docker run --pull=missing\
			--volume $(realpath .):/go/src/github.com/vmware-tanzu/graph-framework-for-microservices/compiler/datamodel/ \
			-v /go/src/github.com/vmware-tanzu/graph-framework-for-microservices/compiler/datamodel/build/ \
			--volume $(realpath .)/build:/go/src/github.com/vmware-tanzu/graph-framework-for-microservices/compiler/generated/ \
			--volume ~/.ssh:/home/$(shell id -u)/.ssh \
			--volume ${HOME}/.cache/go-build:/home/$(shell id -u)/.cache/go-build \
			-e GOCACHE=/home/$(shell id -u)/.cache/go-build \
			-e CRD_MODULE_PATH=${CRD_MODULE_PATH} \
			-e CONFIG_FILE=datamodel/nexus.yaml \
			--user $(shell id -u):$(shell id -g) \
			${COMPILER_DOCKER_TAG} /bin/bash -c 'make generate_code'; \
	else \
		docker run --pull=missing\
			--volumes-from=$(CONTAINER_ID) \
			-e DATAMODEL_PATH=$(DATAMODEL_LOCAL_PATH) \
			-e GENERATED_OUTPUT_DIRECTORY=$(DATAMODEL_LOCAL_PATH)/build \
			-e CONFIG_FILE=${DATAMODEL_LOCAL_PATH}/nexus.yaml \
			-e CRD_MODULE_PATH=${CRD_MODULE_PATH} \
			-e GOPRIVATE=*.eng.vmware.com \
			-e GOINSECURE=*.eng.vmware.com \
			-e CICD_TOKEN=${CICD_TOKEN} \
			${COMPILER_DOCKER_TAG}; \
	fi

.PHONY: docker.build
docker.build:
	@echo "${DOCKER_TAG}" > IMAGE
	@echo "${DATAMODEL_GROUP_NAME}" > NAME
	docker build -t ${DOCKER_TAG} . -f Dockerfile

.PHONY: docker.publish
docker.publish:
	docker push $(DOCKER_TAG)

.PHONY: dm.install
dm.install:
	docker run \
		--net ${DOCKER_NETWORK} \
		--pull=missing \
		--mount type=bind,source=${HOST_KUBECONFIG},target=${MOUNTED_KUBECONFIG},readonly \
		-e KUBECONFIG=${MOUNTED_KUBECONFIG} \
		-e NAME=${DATAMODEL_GROUP_NAME} \
		-e IMAGE=${DOCKER_TAG} \
		-e https_proxy= \
		-e http_proxy= \
		-e HTTPS_PROXY= \
		-e HTTP_PROXY= \
		${DOCKER_TAG}

.PHONY: kind.load
kind.load: docker.build
	kind load docker-image ${DOCKER_TAG} --name ${CLUSTER_NAME}

.PHONY: dm.install.helm
dm.install.helm:
	cd helm-chart; helm install api -n ${NEXUS_API_GW_K8S_NAMESPACE} --set global.datamodelName=api --set image=${DOCKER_TAG} --set imagePull=${DOCKER_IMAGE_PULL} .

.PHONY: dm.uninstall.helm
dm.uninstall.helm:
	helm uninstall api

.PHONY: container-build
container-build: datamodel_build docker.build

.PHONY: container-push
container-push: docker.publish

.PHONY: kind-load-images
kind-load-images:
	$(info Loading api image into kind cluster...)
	docker save ${DOCKER_TAG} | docker exec -i ${CLUSTER_NAME}-control-plane ctr --namespace k8s.io images import -
