<!---
 Copyright (C) 2025 Intel Corporation

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing,
 software distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions
 and limitations under the License.

 SPDX-License-Identifier: Apache-2.0
-->

# Nexus API Gateway

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Overview

Nexus API Gateway is a cloud native application on the Edge Orchestrator. It is a Multi-tenant API Gateway and is one 
of the core and central functions of the Centralized IAM.

The API Gateway provides following functionality:

- It provides an unified API endpoint for all the APIs for accessing Open Edge Platform Services.
- It contains a tenancy aware authentication layer, which authenticates each of the API Request messages.
- It contains a tenancy and hierarchy aware authorization layer for authorization for all the API Request messages.
- It has API remapping module which provides isolation between external and internal service APIs.
- It has a declarative policy management layer and APIs.
- It has an extensible common interception point for all APIs in Open Edge Platform.

The Nexus API Gateway exposes a HTTP endpoint that serves the hierarchical Open Edge Platform APIs. The user request 
terminates at the API gateway, where authentication and authorization related processing is done before 
the request is proxied to the appropriate backend service.

Nexus API Gateway is responsible for 5 primary functions:

- `HTTP Server` - This exposes endpoints to serve the hierarchical Open Edge Platform Service APIs. In addition to those, it also provides following endpoints:
  - `/openapi` - This endpiont returns a unified OpenAPI spec for all Open Edge Platform user-facing APIs.
  - `/swagger` - This will render a swagger page, providing a user friendly page to interact with the Open Edge Platform system using APIs.
- `AuthN` - It is an authentication plugin layer. This layer authenticates the user of an API request. In a detailed way, it does following:
  - Validates the JWT token presented as part of the API request.
  - Extracts following data from the API request:
    - Org/Tenant of the user from JWT
    - Active project associated with the request. This is inferred from the API Request URL.
    - All Projects that the user is associated with. This is from JWT.
    - Claims associated with the user. This is from the JWT.
  - Returns the result of Authentication to the API Gateway
- `AuthZ` - It is an authorization plugin layer. This layer authorizes the user of an API request. authorization layer will constitute a centralized policy decision and enforcement point in Open Edge Platform. This layer is independent of the PDP/PEP implemented in the Open Edge Platform services. It has broadly following 2 parts:
  - `Policy specification` - This comprises of definition of roles and rules 
  - `Policy enforcement` - This layer deals with the enforcement logic of policies that should be applied to the API request by an user. The incoming requests can be broadly classfied into 3 categories/types
    - IAM/SI Admin Persona
    - Org Admin Persona
    - Open Edge Platform User Persona
- `API Remapping` - API remapping plugin provides a URI rewrite scheme, that maps external facing URIs with internal representation of those corresponding APIs with following benefits:
  - Exposes a multi-tenant, hierarchical API to the external users, that is structured to compliant with industry standard API guidelines like Google AIP’s, AWS API etc.
  - Avoids the need for backend services to migrate the served APIs to roll this out to the user. Rather, the remapping feature provides a path where the user gets the benefit of the newer API right away, without requiring the backend to be changed as a pre-requisite.
  The APIs to be mapped are specified as configuration by the admin, at install/upgrade time, through API remapping K8s CRD. The expectation is that, these CR’s will only be created/updated at the time of install or upgrade. Otherwise these mappings are not expected to be changed at runtime.
- `Proxy to Open Edge Platform Services` - Upon completion of processing of all plugins, the request is now ready to be proxied to the appropriate Open Edge Platform backend service. The response from the backend will be copied back to the original request, back to the user.

## Get Started

Install Nexus API Gateway.

```
helm install -n orch-iam --create-namespace charts/nexus-api-gw
```

Another way to try out Nexus API Gateway is by using the Open Edge Platform deployment. 

## Develop

Nexus API Gateway is developed in the **Go** language and is built as a Docker image, through a `Dockerfile` 
which is in `nexus-api-gateway` folder. The CI integration for this repository will publish the container
image to the Edge Orchestrator Release Service OCI registry upon merge to the `main` branch.

Nexus API Gateway has a corresponding Helm chart in `charts/nexus-api-gw` folder. The CI integration for 
this repository will publish this Helm charts to the Edge Orchestrator Release Service OCI registry upon 
merge to `main` branch. Nexus API Gateway is deployed to the Edge Orchestrator using this Helm chart, 
whose lifecycle is in turn managed by Argo CD (see [Foundational Platform]).

Instructions on how to build, install and test.

### Prerequisites

This code requires the following tools to be installed on your development machine:

- [Go\* programming language](https://go.dev) - check [Makefile](./Makefile) on usage
- [golangci-lint](https://github.com/golangci/golangci-lint) - check [Makefile](./Makefile)  on usage
- [hadolint](https://github.com/hadolint/hadolint) - check [Makefile](./Makefile)  on usage
- [yamllint](https://github.com/adrienverge/yamllint) - check [Makefile](./Makefile)  on usage
- [reuse](https://github.com/fsfe/reuse-tool) - check [Makefile](./Makefile)  on usage
- Python\* programming language version 3.10 or later
- [gocover-cobertura](github.com/boumenot/gocover-cobertura) - check [Makefile](./Makefile)  on usage
- [Docker](https://docs.docker.com/engine/install/) to build containers
- [Helm](https://helm.sh/docs/intro/install/) for install helm charts for end-to-end tests

### Build, Scan and Test

The basic workflow to make changes to the code, verify those changes, and create a Github pull request (PR) is:

0. Edit and build the code with `make go-build` command

1. Run linters with `make lint` command

2. Run the unit tests with `make test` command

3. Build the code with `make build` command

## Contribute

We welcome contributions from the community! To contribute, please open a pull request to have your changes reviewed
and merged into the main. We encourage you to add appropriate unit tests and e2e tests if your contribution introduces
a new feature. See the [CONTRIBUTING.md](../CONTRIBUTING.md) file for more information.

Additionally, ensure the following commands are successful:

```shell
make test
make lint
make license
make build
```
You can use `help` to see a list of makefile targets. The following is a list of makefile targets that support developer activities:

- `fmt` to run go fmt against golang source files
- `vet` to run go vet against golang source files
- `lint` to run a list of linting targets
- `license` to run a scan for the License headers in source code files
- `hadolint` to lint dockerfile(s) using hadolint
- `go-lint` to lint golang source files with golangci-lint
- `yamllint` to lint yaml files using yamllint
- `go-tidy` to update the Go dependencies and regenerate the go.sum file
- `go-build` to build the go source code files
- `test` to run the unit test
- `coverage` to run the unit test coverage
- `build` to build the nexus API gateway Docker container
- `release` to publish the built nexus API gateway Docker container to pre-defined docker container registry. This registry is set in a env variable (API_GW_DOCKER_IMAGE_OEP) in nexus-api-gateway/Makefile

## Community and Support

To learn more about the project, its community, and governance, visit the [Edge Orchestrator Community](https://github.com/open-edge-platform).
For support, start with [Troubleshooting](https://github.com/open-edge-platform) or [contact us](https://github.com/open-edge-platform).

## License

Nexus API Gateway is licensed under Apache 2.0.

[Nexus API Gateway]: https://github.com/open-edge-platform/orch-utils/nexus-api-gateway
[Foundational Platform]: https://literate-adventure-7vjeyem.pages.github.io/developer_guide/foundational_platform/foundational_platform_main.html

