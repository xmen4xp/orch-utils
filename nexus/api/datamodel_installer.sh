#!/bin/bash

set -ex

NAME=${NAME:-}
DATAMODEL_IMAGE=${IMAGE}
IMAGE="file://${NAME}"
# TITLE resolution order (first non-empty wins):
#   1. TITLE env var (deploy-time override)
#   2. `title:` field in /nexus.yaml when present (datamodels that
#      bundle their nexus.yaml into the installer image get a single
#      source of truth shared with the build-time openapi-generator)
#   3. Script default (applied via the empty-TITLE branch below)
# The awk extractor is POSIX-only (no yq dependency). The trailing
# `|| true` is REQUIRED: this script runs under `set -e`, and awk
# exits nonzero when /nexus.yaml is absent (which is the common case
# for the orch-utils template image — only datamodels that bundle
# their nexus.yaml into the installer image will find the file). The
# `|| true` keeps the assignment quiet and the script alive in both
# cases, with TITLE remaining empty so the existing fallback below
# applies.
TITLE=${TITLE:-$(awk -F'"' '/^title:[[:space:]]*"/ {print $2; exit}' /nexus.yaml 2>/dev/null || true)}
SKIP_CRD_INSTALLATION=${SKIP_CRD_INSTALLATION:-false}
GRAPHQL_ENABLED=${GRAPHQL_ENABLED:-false}

### User can pass the custom HTTP URL where the graphql plugin can be downloaded via internet
GRAPHQL_PATH=${GRAPHQL_PATH:-NA}

echo '
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  creationTimestamp: null
  name: datamodels.nexus.com
spec:
  conversion:
    strategy: None
  group: nexus.com
  names:
    kind: Datamodel
    listKind: DatamodelList
    plural: datamodels
    shortNames:
    - datamodel
    singular: datamodel
  scope: Cluster
  versions:
  - name: v1
    schema:
      openAPIV3Schema:
        type: object
        properties:
          apiVersion:
            type: string
          kind:
            type: string
          metadata:
            type: object
          spec:
            properties:
              name:
                type: string
              url:
                type: string
              title:
                type: string
                default: "Nexus API GW APIs"
              enableGraphql:
                type: boolean
                default: false
              graphqlPath:
                type: string
                default: ""
            type: object
    served: true
    storage: true
' | kubectl apply -f -

echo '
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: extensionrestapis.nexus.com
spec:
  group: nexus.com
  scope: Cluster
  names:
    plural: extensionrestapis
    singular: extensionrestapi
    kind: ExtensionRestAPI
    shortNames:
      - erapi
  versions:
    - name: v1
      served: true
      storage: true
      additionalPrinterColumns:
        - name: URI
          type: string
          jsonPath: .spec.uri
        - name: Age
          type: date
          jsonPath: .metadata.creationTimestamp
      schema:
        openAPIV3Schema:
          type: object
          description: ExtensionRestAPI stores one REST URI and its OpenAPI path fragment.
          properties:
            apiVersion:
              type: string
            kind:
              type: string
            metadata:
              type: object
            spec:
              type: object
              required:
                - uri
              properties:
                uri:
                  type: string
                  minLength: 1
                  description: REST API URI path with optional path parameters
                methods:
                  type: array
                  description: HTTP methods to proxy (e.g., ["GET", "POST"]). Empty = all methods.
                  items:
                    type: string
                    enum: ["GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"]
                openAPIPathSpec:
                  type: string
                  description: Raw OpenAPI Path Item fragment as YAML or JSON string
            status:
              type: object
              properties:
                phase:
                  type: string
                  enum: ["Registered", "Rejected", "Pending"]
                  description: Current registration phase
                message:
                  type: string
                  description: Human-readable status message
                registeredRoutes:
                  type: array
                  description: Successfully registered URI+Method combinations
                  items:
                    type: object
                    properties:
                      uri:
                        type: string
                      method:
                        type: string
                collisions:
                  type: array
                  description: Collision details if phase is Rejected
                  items:
                    type: object
                    properties:
                      uri:
                        type: string
                      method:
                        type: string
                      conflictingCR:
                        type: string
                      conflictingSource:
                        type: string
                        enum: ["nexus-crd", "extension-rest-api"]
                lastUpdated:
                  type: string
                  format: date-time
                  description: Timestamp of last status update
      subresources:
        status: {}
' | kubectl apply -f -

echo '
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: extensionrestapiendpoints.nexus.com
spec:
  group: nexus.com
  scope: Cluster
  names:
    plural: extensionrestapiendpoints
    singular: extensionrestapiendpoint
    kind: ExtensionRestAPIEndpoint
    shortNames:
      - erapiendpoint
  versions:
    - name: v1
      served: true
      storage: true
      additionalPrinterColumns:
        - name: ExtensionRestAPI
          type: string
          jsonPath: .spec.extensionRestAPIRef
        - name: Service
          type: string
          jsonPath: .spec.service
        - name: Port
          type: string
          jsonPath: .spec.port
        - name: Age
          type: date
          jsonPath: .metadata.creationTimestamp
      schema:
        openAPIV3Schema:
          type: object
          description: ExtensionRestAPIEndpoint provides backend service configuration for an ExtensionRestAPI.
          properties:
            apiVersion:
              type: string
            kind:
              type: string
            metadata:
              type: object
            spec:
              type: object
              required:
                - extensionRestAPIRef
                - service
                - port
              properties:
                extensionRestAPIRef:
                  type: string
                  minLength: 1
                  description: Name of the ExtensionRestAPI CR this endpoint provides backend for
                service:
                  type: string
                  minLength: 1
                  description: Fully qualified service DNS name (e.g., metrics-api.hdai-system.svc.cluster.local)
                port:
                  type: string
                  minLength: 1
                  description: Service port number or name (e.g., 8080 or http)
' | kubectl apply -f -

### This is to support older way of installing datamodel from local folder
if [[ $SKIP_CRD_INSTALLATION == "false" ]]; then
    kubectl apply -f /crds --recursive
    [[ $GRAPHQL_PATH != NA ]] && GRAPHQL_ENABLED=true
    if  test -f /build/server; then
        GRAPHQL_ENABLED=true
    fi
fi
echo $NAME
### We will create datamodel object
if [[ -n $NAME ]] && [[ -n $IMAGE ]]; then
  if [[ -n $TITLE ]]; then
    echo '
      apiVersion: nexus.com/v1
      kind: Datamodel
      metadata:
        name: '"$NAME"'
      spec:
        name: '"$NAME"'
        url: '"$IMAGE"'
        title: '"$TITLE"'
        enableGraphql: '"$GRAPHQL_ENABLED"'' | kubectl apply -f -
  else
    echo '
    apiVersion: nexus.com/v1
    kind: Datamodel
    metadata:
      name: '"$NAME"'
    spec:
      name: '"$NAME"'
      url: '"$IMAGE"'
      enableGraphql: '"$GRAPHQL_ENABLED"'' | kubectl apply -f -
  fi
fi

# Apply ExtensionRestAPI CRs if they exist
if [[ -d /extensionrestapi ]]; then
    echo "Applying ExtensionRestAPI CRs from /extensionrestapi"
    kubectl apply -f /extensionrestapi --recursive || true
fi
