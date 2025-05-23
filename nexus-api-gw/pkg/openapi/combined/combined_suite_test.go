// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package combined_test

import (
	"testing"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/open-edge-platform/orch-utils/nexus-api-gw/pkg/config"
	log "github.com/sirupsen/logrus"
)

const (
	URI         = "/v1alpha1/project/{projectId}/global-namespaces"
	ResourceURI = "/v1alpha1/project/{projectId}/global-namespaces/{id}"
	ListURI     = "/v1alpha1/global-namespaces/test"
)

var spec = []byte(`openapi: 3.0.0
info:
  version: 1.0.0
  title: NSX-SM <Tenant/Operator> APIs
  description: <Tenant/Operator> APIs for NSX service mesh.
  termsOfService: 'http://nsxservicemesh.vmware.com/terms/'
  contact:
    name: VMware NSX-ServiceMesh Team
    email: support@nsxservicemesh.vmware.com
    url: 'http://nsxservicemesh.vmware.com/'
  license:
    name: VMWare
    url: 'https://nsxservicemesh.vmware.com/licenses/LICENSE.html'
servers:
  - url: 'http://127.0.0.1:3000'
basePath: /v1
paths:
  '/v1alpha1/project/{projectId}/global-namespaces/{id}':
    put:
      x-controller-name: GlobalNamespaceControllerV1Alpha1
      x-operation-name: putGlobalNamespaceV1
      tags:
        - Global Namespaces (v1alpha1)
      description: Create the global namespace
      x-nexus-kind-name: GlobalNamespace
      x-nexus-group-name: gns.vmware.org
      x-nexus-identifier: id
      x-nexus-short-name: gns
      responses:
        '200':
          description: 'global namespace updated '
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GlobalNamespaceConfig'
        '201':
          description: 'global namespace created '
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GlobalNamespaceConfig'
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ApiHttpError'
      parameters:
        - name: id
          in: path
          schema:
            type: string
          required: true
      requestBody:
        description: Global namespace Config
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GlobalNamespaceConfig'
        x-parameter-index: 1
      operationId: GlobalNamespaceControllerV1Alpha1.putGlobalNamespaceV1
    get:
      x-controller-name: GlobalNamespaceControllerV1Alpha1
      x-operation-name: getGlobalNamespaceV1
      x-nexus-identifier: id
      tags:
        - Global Namespaces (v1alpha1)
      description: Return the config for a global namespace
      x-nexus-kind-name: GlobalNamespace
      x-nexus-group-name: gns.vmware.org
      x-nexus-short-name: gns
      responses:
        '200':
          description: global namespace config
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GlobalNamespaceConfig'
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ApiHttpError'
      parameters:
        - name: id
          in: path
          schema:
            type: string
          required: true
      operationId: GlobalNamespaceControllerV1Alpha1.getGlobalNamespaceV1
    delete:
      x-controller-name: GlobalNamespaceControllerV1Alpha1
      x-operation-name: deleteGlobalNamespaceV1
      tags:
        - Global Namespaces (v1alpha1)
      description: Delete the global namespace
      x-nexus-kind-name: GlobalNamespace
      x-nexus-group-name: gns.vmware.org
      x-nexus-identifier: id
      x-nexus-short-name: gns
      responses:
        '200':
          description: 'global namespace delete '
          content:
            application/json:
              schema:
                type: string
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ApiHttpError'
      parameters:
        - name: id
          in: path
          schema:
            type: string
          required: true
      operationId: GlobalNamespaceControllerV1Alpha1.deleteGlobalNamespaceV1
  /v1alpha1/project/{projectId}/global-namespaces:
    post:
      x-controller-name: GlobalNamespaceControllerV1Alpha1
      x-operation-name: postGlobalNamespaceV1
      tags:
        - Global Namespaces (v1alpha1)
      description: Create the global namespace
      responses:
        '200':
          description: 'global namespace created '
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GlobalNamespaceConfig'
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ApiHttpError'
      requestBody:
        description: Global namespace Config
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/GlobalNamespaceConfig'
      operationId: GlobalNamespaceControllerV1Alpha1.postGlobalNamespaceV1
    get:
      x-controller-name: GlobalNamespaceControllerV1Alpha1
      x-operation-name: getGNSListV1
      tags:
        - Global Namespaces (v1alpha1)
      description: Get a list of GNS IDs that are defined
      x-nexus-kind-name: GlobalNamespace
      x-nexus-group-name: gns.vmware.org
      x-nexus-short-name: gns
      responses:
        '200':
          description: list of gns defined
          content:
            application/json:
              schema:
                type: array
                items:
                  type: string
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ApiHttpError'
      operationId: GlobalNamespaceControllerV1Alpha1.getGNSListV1
  /v1alpha1/global-namespaces/test:
    get:
      x-controller-name: GlobalNamespaceControllerV1Alpha1
      x-operation-name: getGNSListV1
      tags:
        - Global Namespaces (v1alpha1)
      description: Get a list of GNS IDs that are defined
      x-nexus-kind-name: GlobalNamespaceList
      x-nexus-group-name: gns.vmware.org
      x-nexus-list-endpoint: true
      x-nexus-short-name: gns
      responses:
        '200':
          description: list of gns defined
          content:
            application/json:
              schema:
                type: array
                items:
                  type: string
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ApiHttpError'
      operationId: GlobalNamespaceControllerV1Alpha1.getGNSListV1
components:
  securitySchemes:
    BasicAuth:
      type: http
      scheme: basic
  examples:
   objectExample:
      value:
        id: 1
        name: new object
      summary: A sample object
  callbacks:
    myEvent:   # Event name
          '{$request.body#/callbackUrl}':   # The callback URL,
                                            # Refers to the passed URL
            post:
              requestBody:   # Contents of the callback message
                required: true
                content:
                  application/json:
                    schema:
                      type: object
                      properties:
                        message:
                          type: string
                          example: Some event happened
                      required:
                        - message
              responses:   # Expected responses to the callback message
                '200':
                  description: Your server returns this code if it accepts the callback
  schemas:
    ApiHttpError:
      title: ApiHttpError
      properties:
        code:
          type: number
        message:
          type: string
      required:
        - code
        - message
      additionalProperties: false
    Condition:
      title: Condition
      properties:
        type:
          type: string
          description: START_WITH | EXACT
        match:
          type: string
      additionalProperties: false
    MatchCondition:
      title: MatchCondition
      properties:
        namespace:
          $ref: '#/components/schemas/Condition'
        cluster:
          $ref: '#/components/schemas/Condition'
        service: 
          $ref: '#/components/schemas/MatchCondition'
      required:
        - namespace
      additionalProperties: false
    GlobalNamespaceConfig:
      title: GlobalNamespaceConfig
      properties:
        name:
          type: string
          pattern: '^[a-z0-9][a-z0-9-.]*[a-z0-9]$'
          minLength: 2
          maxLength: 253
        display_name:
          type: string
        domain_name:
          type: string
        use_shared_gateway:
          type: boolean
        mtls_enforced:
          type: boolean
        ca_type:
          type: string
          enum:
            - PreExistingCA
            - self-signed
        ca:
          type: string
        description:
          type: string
        color:
          type: string
        version:
          type: string
        match_conditions:
          type: array
          items:
            $ref: '#/components/schemas/MatchCondition'
        api_discovery_enabled:
          type: boolean
      required:
        - name
        - domain_name
        - match_conditions
      additionalProperties: false`)

func TestDeclarative(t *testing.T) {
	log.StandardLogger().ExitFunc = nil
	gomega.RegisterFailHandler(ginkgo.Fail)
	config.Cfg = &config.Config{
		TenantAPIGwDomain: "http://test",
	}
	ginkgo.RunSpecs(t, "Declarative Suite")
}

var crdExample = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    nexus: |
      {"name":"management.Leader","hierarchy":["roots.orgchart.vmware.org"],
      "children":{"humanresourceses.hr.vmware.org":{"fieldName":"HR","fieldNameGvk":"hRGvk","isNamed":false},
      "mgrs.management.vmware.org":{"fieldName":"EngManagers","fieldNameGvk":"engManagersGvk","isNamed":true}
      },
      "links":{"Role":{"fieldName":"Role","fieldNameGvk":"roleGvk","isNamed":false}
      },
      "is_singleton":true,
      "nexus-rest-api-gen":{
      "uris":[
      {"uri":"/root/{orgchart.Root}/leader/{management.Leader}",
      "methods":{
      "DELETE":{"200":{"description":"OK"},"404":{"description":"Not Found"},
      "501":{"description":"Not Implemented"}},
      "GET":{"200":{"description":"OK"},
      "404":{"description":"Not Found"},"501":{"description":"Not Implemented"}
      },
      "PUT":{
      "200":{"description":"OK"},"201":{"description":"Created"},"501":{"description":"Not Implemented"}}
      },"auth":false},
      {"uri":"/leader",
      "methods":{
      "DELETE":{"200":{"description":"OK"},"404":{"description":"Not Found"},"501":{"description":"Not Implemented"}},
      "GET":{"200":{"description":"OK"},"404":{"description":"Not Found"},"501":{"description":"Not Implemented"}},
      "PUT":{"200":{"description":"OK"},"201":{"description":"Created"},"501":{"description":"Not Implemented"}}},
      "auth":false},
      {"uri":"/leaders",
      "methods":{
      "LIST":{"200":{"description":"OK"},"404":{"description":"Not Found"},"501":{"description":"Not Implemented"}}
      },
      "auth":false}]}}
  creationTimestamp: null
  name: leaders.management.vmware.org
spec:
  conversion:
    strategy: None
  group: management.vmware.org
  names:
    kind: Leader
    listKind: LeaderList
    plural: leaders
    shortNames:
    - leader
    singular: leader
  scope: Cluster
  versions:
  - name: v1
    schema:
      openAPIV3Schema:
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation
              of an object. Servers should convert recognized schemas to the latest
              internal value, and may reject unrecognized values.
              More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this
              object represents. Servers may infer this from the endpoint the client
              submits requests to. Cannot be updated. In CamelCase.
              More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            properties:
              designation:
                type: string
              employeeID:
                format: int32
                type: integer
              engManagersGvk:
                additionalProperties:
                  properties:
                    group:
                      type: string
                    kind:
                      type: string
                    name:
                      type: string
                  required:
                  - group
                  - kind
                  - name
                  type: object
                type: object
              hRGvk:
                properties:
                  group:
                    type: string
                  kind:
                    type: string
                  name:
                    type: string
                required:
                - group
                - kind
                - name
                type: object
              name:
                type: string
              roleGvk:
                properties:
                  group:
                    type: string
                  kind:
                    type: string
                  name:
                    type: string
                required:
                - group
                - kind
                - name
                type: object
            required:
            - designation
            - name
            - employeeID
            type: object
          status:
            properties:
              nexus:
                properties:
                  remoteGeneration:
                    format: int64
                    type: integer
                  sourceGeneration:
                    format: int64
                    type: integer
                required:
                - sourceGeneration
                - remoteGeneration
                type: object
              status:
                properties:
                  DaysLeftToEndOfVacations:
                    format: int32
                    type: integer
                  IsOnVacations:
                    type: boolean
                required:
                - IsOnVacations
                - DaysLeftToEndOfVacations
                type: object
            type: object
        type: object
    served: true
    storage: true
    subresources:
      status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: null
  storedVersions:
  - v1

`
