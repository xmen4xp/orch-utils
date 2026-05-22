DOCKER_REPO                     ?= ciscohyperdrive
TAG                             ?= latest
K8S_PROXY_COMPONENT_NAME        ?= nexus-k8s-proxy

ifeq (${DOCKER_TAG}, )
ifeq (${DOCKER_REGISTRY}, )
DOCKER_TAG=${K8S_PROXY_COMPONENT_NAME}:${TAG}
else
DOCKER_TAG=${DOCKER_REGISTRY}/${DOCKER_REPO}/${K8S_PROXY_COMPONENT_NAME}:${TAG}
endif
endif

.PHONY: container-build
container-build:
	@echo "Building k8s proxy container image..."
	docker build -t ${DOCKER_TAG} -f ./Dockerfiles/Dockerfile.kubectl .

.PHONY: kind-load-images
kind-load-images:
	$(info Loading k8s proxy images into kind cluster...)
	docker save ${DOCKER_TAG} | docker exec -i ${CLUSTER_NAME}-control-plane ctr --namespace k8s.io images import -
