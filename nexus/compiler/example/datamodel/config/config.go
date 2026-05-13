// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"net/http"

	"github.com/vmware-tanzu/graph-framework-for-microservices/compiler/example/datamodel/config/gns"
	servicegroup "github.com/vmware-tanzu/graph-framework-for-microservices/compiler/example/datamodel/config/gns/service-group"
	py "github.com/vmware-tanzu/graph-framework-for-microservices/compiler/example/datamodel/config/policy"
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
)

var nonNexusValue = 1
var nonValue int

var BarCustomCodesResponses = nexus.HTTPCodesResponse{
	http.StatusBadRequest: nexus.HTTPResponse{Description: "Bad Request"},
}

var BarCustomMethodsResponses = nexus.HTTPMethodsResponses{
	http.MethodPatch: BarCustomCodesResponses,
}

// nexus-graphql-query:root.GeneralGraphQLQuerySpec
type Config struct {
	nexus.Node
	GNS         gns.Gns                `nexus:"child"`
	DNS         gns.Dns                `nexus:"child"`
	VMPPolicies py.VMpolicy            `nexus:"child"`
	ACPPolicies py.AccessControlPolicy `nexus:"links"`
	Domain      Domain                 `nexus:"child"`
	// Examples for cross-package import.
	MyStr0 *gns.MyStr
	MyStr1 []gns.MyStr
	MyStr2 map[string]gns.MyStr

	XYZPort           gns.Port
	ABCHost           []gns.Host
	ClusterNamespaces []ClusterNamespace

	TestValMarkers TestValMarkers `json:"testValMarkers" yaml:"testValMarkers"`
	FooExample     FooTypeABC     `nexus:"children"`
	Instance       float32
	CuOption       string                        `json:"option_cu"`
	SvcGrpInfo     servicegroup.SvcGroupLinkInfo `nexus:"child"`
}

type FooTypeABC struct {
	nexus.Node
	FooA AMap
	FooB BArray
	FooC CInt   `nexus-graphql:"ignore:true"`
	FooD DFloat `nexus-graphql:"type:string"`
	FooE CInt   `json:"foo_e" nexus-graphql:"ignore:true"`
	FooF DFloat `json:"foo_f" yaml:"c_int" nexus-graphql:"type:string"`
}

type Domain struct {
	nexus.Node
	PointPort        *gns.Port
	PointString      *string
	PointInt         *int
	PointMap         *map[string]string
	PointSlice       *[]string
	SliceOfPoints    []*string
	SliceOfArrPoints []*BArray
	MapOfArrsPoints  map[string]*BArray
	PointStruct      *Cluster
}

type ClusterNamespace struct {
	Cluster   MatchCondition
	Namespace MatchCondition
}

type MatchCondition struct {
	Name string
	Type gns.Host
}

type Cluster struct {
	Name string
	MyID int
}

type AMap map[string]string

type BArray []string
type CInt uint8
type DFloat float32

type CrossPackageTester struct {
	Test gns.MyStr
}

type EmptyStructTest struct{}

type TestValMarkers struct {
	//nexus-validation: MaxLength=8, MinLength=2, Pattern=ab
	MyStr string `json:"myStr" yaml:"myStr"`

	//nexus-validation: Maximum=8, Minimum=2
	//nexus-validation: ExclusiveMaximum=true
	MyInt int `json:"myInt" yaml:"myInt"`

	//nexus-validation: MaxItems=3, MinItems=2
	//nexus-validation: UniqueItems=true
	MySlice []string `json:"mySlice" yaml:"mySlice"`
}

type SomeStruct struct{}

type StructWithEmbeddedField struct {
	SomeStruct
	gns.MyStr
}

// =============================================================================
// ExtensionRestAPI Examples - Telemetry Manager APIs
// =============================================================================

var MetricCategoriesAPI = nexus.ExtensionRestAPI{
	Uri: "/metrics/categories",
	BackendRef: nexus.ExtensionRestAPIBackend{
		Name:     "example-backend",
		PortName: "http",
	},
	OpenAPIPathSpec: `
get:
  tags:
    - Discovery
  summary: List available metric categories
  description: Returns a list of all available metric categories in the system
  operationId: listMetricCategories
  parameters:
    - name: start
      in: query
      required: true
      description: Start time for the query in RFC3339 format
      schema:
        type: string
        format: date-time
    - name: end
      in: query
      required: true
      description: End time for the query in RFC3339 format
      schema:
        type: string
        format: date-time
    - name: limit
      in: query
      required: false
      schema:
        type: integer
        minimum: 1
        maximum: 1000
        default: 100
    - name: offset
      in: query
      required: false
      schema:
        type: integer
        minimum: 0
        default: 0
  responses:
    "200":
      description: Successfully retrieved metric categories
    "400":
      description: Invalid request parameters
    "401":
      description: Missing or invalid tenant headers
    "408":
      description: Request timeout
    "429":
      description: Rate limit exceeded
    "500":
      description: Internal server error
`,
}

var MetricListAPI = nexus.ExtensionRestAPI{
	Uri: "/metrics/list",
	BackendRef: nexus.ExtensionRestAPIBackend{
		Name:     "example-backend",
		PortName: "http",
	},
	OpenAPIPathSpec: `
get:
  tags:
    - Discovery
  summary: List available metrics with metadata
  description: Returns a list of all available metrics in the system with their metadata
  operationId: listMetrics
  parameters:
    - name: start
      in: query
      required: true
      schema:
        type: string
        format: date-time
    - name: end
      in: query
      required: true
      schema:
        type: string
        format: date-time
    - name: categories
      in: query
      required: false
      description: Comma-separated list of metric categories to filter by
      schema:
        type: string
    - name: limit
      in: query
      required: false
      schema:
        type: integer
        default: 100
  responses:
    "200":
      description: Successfully retrieved metric names
    "400":
      description: Invalid request parameters
    "401":
      description: Missing or invalid tenant headers
    "500":
      description: Internal server error
`,
}

var MetricLabelsAPI = nexus.ExtensionRestAPI{
	Uri: "/metrics/labels",
	BackendRef: nexus.ExtensionRestAPIBackend{
		Name:     "example-backend",
		PortName: "http",
	},
	OpenAPIPathSpec: `
get:
  tags:
    - Discovery
  summary: List available metric label keys
  description: Returns a list of all label keys used in metrics
  operationId: listMetricLabels
  parameters:
    - name: start
      in: query
      required: true
      schema:
        type: string
        format: date-time
    - name: end
      in: query
      required: true
      schema:
        type: string
        format: date-time
    - name: categories
      in: query
      required: false
      schema:
        type: string
    - name: metric_names
      in: query
      required: false
      schema:
        type: string
  responses:
    "200":
      description: Successfully retrieved metric labels
    "400":
      description: Invalid request parameters
    "401":
      description: Missing or invalid tenant headers
    "500":
      description: Internal server error
`,
}

var GlobalMetricsAPI = nexus.ExtensionRestAPI{
	Uri: "/metrics",
	BackendRef: nexus.ExtensionRestAPIBackend{
		Name:     "example-backend",
		PortName: "http",
	},
	OpenAPIPathSpec: `
get:
  tags:
    - Metrics
  summary: Query metrics across all infrastructure
  description: Retrieve metrics data from all datacenters, clusters, and nodes
  operationId: getMetrics
  parameters:
    - name: start
      in: query
      required: true
      schema:
        type: string
        format: date-time
    - name: end
      in: query
      required: true
      schema:
        type: string
        format: date-time
    - name: categories
      in: query
      required: true
      description: Comma-separated list of metric categories
      schema:
        type: string
    - name: metric_names
      in: query
      required: true
      description: Comma-separated list of metric names
      schema:
        type: string
    - name: filter
      in: query
      required: false
      description: Filter by dimensions using key:value pairs
      schema:
        type: string
    - name: query_type
      in: query
      required: false
      schema:
        type: string
        enum: [aggregate, raw, trend, custom]
        default: aggregate
    - name: aggregation_interval
      in: query
      required: false
      schema:
        type: string
    - name: group_by
      in: query
      required: false
      schema:
        type: string
    - name: limit
      in: query
      required: false
      schema:
        type: integer
        default: 100
  responses:
    "200":
      description: Successfully retrieved analytics metrics
    "400":
      description: Invalid request parameters
    "401":
      description: Missing or invalid tenant headers
    "408":
      description: Request timeout
    "429":
      description: Rate limit exceeded
    "500":
      description: Internal server error
`,
}

var DatacenterMetricsAPI = nexus.ExtensionRestAPI{
	Uri: "/datacenter/{datacenter-id}/metrics",
	BackendRef: nexus.ExtensionRestAPIBackend{
		Name:     "example-backend",
		PortName: "http",
	},
	OpenAPIPathSpec: `
get:
  tags:
    - Metrics
  summary: Query datacenter-specific metrics
  description: Retrieve metrics for all clusters and nodes within a specific datacenter
  operationId: getDatacenterMetrics
  parameters:
    - name: datacenter-id
      in: path
      required: true
      description: Unique identifier for the datacenter
      schema:
        type: string
        pattern: '^[a-zA-Z0-9\-_]+$'
    - name: start
      in: query
      required: true
      schema:
        type: string
        format: date-time
    - name: end
      in: query
      required: true
      schema:
        type: string
        format: date-time
    - name: categories
      in: query
      required: true
      schema:
        type: string
    - name: metric_names
      in: query
      required: true
      schema:
        type: string
    - name: query_type
      in: query
      required: false
      schema:
        type: string
        enum: [aggregate, raw, trend, custom]
        default: aggregate
    - name: limit
      in: query
      required: false
      schema:
        type: integer
        default: 100
  responses:
    "200":
      description: Successfully retrieved metrics
    "400":
      description: Invalid request parameters
    "401":
      description: Missing or invalid tenant headers
    "404":
      description: Datacenter not found
    "408":
      description: Request timeout
    "500":
      description: Internal server error
`,
}

var ClusterMetricsAPI = nexus.ExtensionRestAPI{
	Uri: "/datacenter/{datacenter-id}/cluster/{cluster-id}/metrics",
	BackendRef: nexus.ExtensionRestAPIBackend{
		Name:     "example-backend",
		PortName: "http",
	},
	OpenAPIPathSpec: `
get:
  tags:
    - Metrics
  summary: Query cluster-specific metrics
  description: Retrieve metrics for all nodes within a specific cluster in a datacenter
  operationId: getClusterMetrics
  parameters:
    - name: datacenter-id
      in: path
      required: true
      description: Unique identifier for the datacenter
      schema:
        type: string
        pattern: '^[a-zA-Z0-9\-_]+$'
    - name: cluster-id
      in: path
      required: true
      description: Unique identifier for the cluster
      schema:
        type: string
        pattern: '^[a-zA-Z0-9\-_]+$'
    - name: start
      in: query
      required: true
      schema:
        type: string
        format: date-time
    - name: end
      in: query
      required: true
      schema:
        type: string
        format: date-time
    - name: categories
      in: query
      required: true
      schema:
        type: string
    - name: metric_names
      in: query
      required: true
      schema:
        type: string
    - name: query_type
      in: query
      required: false
      schema:
        type: string
        enum: [aggregate, raw, trend, custom]
        default: aggregate
    - name: limit
      in: query
      required: false
      schema:
        type: integer
        default: 100
  responses:
    "200":
      description: Successfully retrieved metrics
    "400":
      description: Invalid request parameters
    "401":
      description: Missing or invalid tenant headers
    "404":
      description: Datacenter or cluster not found
    "408":
      description: Request timeout
    "500":
      description: Internal server error
`,
}
