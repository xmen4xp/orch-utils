## Tenancy Manager Overview

The **Tenancy Manager** is a core component responsible for monitoring multi-tenancy config object creation and generating corresponding runtime objects for these events. It ensures the correct provisioning, tracking, and status reporting of org and project creation/deletion.

The **Tenancy Manager** subscribes to config events related to **organization (org), project, orgActiveWatcher,** and **projectActiveWatcher** creation/deletion. During startup, it registers **Org Watchers** and **Project Watchers**.  

## Key Responsibilities  

### Org/Project Creation Scenario 

- When an org or project is created, the corresponding runtime entity is created.  
- It iterates through all registered Org Watchers and verifies the creation of Org Active Watchers. Similarly, it checks for the creation of Project Active Watchers corresponding to registered Project Watchers.  
- Once all required Active Watchers are successfully created, the Tenancy Manager marks the config status as Complete.  
- If at least one Active Watcher is missing or not in a successful state, the configuration status is marked as In Progress.  
- In case of any error during Org Active Watcher or Project Active Watcher processing, the status is marked as Error.  
- If the Active Watchers are not created within a defined time interval, the org status is set to Timeout.  

### Org/Project Deletion Scenario 

- When an org or project is deleted, the corresponding runtime entity is deleted.  
- The Tenancy Manager verifies that all associated Org Active Watchers and Project Active Watchers are deleted.
- If at least one Active Watcher is present, the configuration status is marked as In Progress.
- If any errors occur during the deletion process, the status is updated to Error.  
- If the deletion does not complete within a defined time interval, the status is marked as Timeout.  
 
### Status Reporting  

The Tenancy Manager is responsible for reporting the status of org and project creation/deletion back to the user. This status is communicated through the "status" section of the Org** and Project objects in the data model.  

The status is represented using well-defined enums that are consistently used across all the components to ensure clear communication.

| **Enum**                          | **Description** |  
|------------------------------------|-----------------------------------------------------------------|  
| **STATUS_INDICATION_ERROR**        | Error state. Indicates that the last request was not completed successfully. |  
| **STATUS_INDICATION_IN_PROGRESS**  | In progress. Indicates that the last request is still being processed. |  
| **STATUS_INDICATION_IDLE**         | Steady state. Indicates that the last request was successfully completed, and the system is idle, waiting for future requests. |  

## Get Started

Install Tenancy Manager.

```
helm install -n orch-iam --create-namespace charts/tenancy-manager
```

Another way to try out Tenancy API Mapping is by using the Open Edge Platform Deployment.

## Develop

- Tenancy-Manager is developed in the **Go** language and is built as a Docker image, through a `Dockerfile` which is in `tenancy-manager` folder. The CI integration for this repository will publish the container image to the Edge Orchestrator Release Service OCI registry upon merge to the `main` branch.

- Tenancy-Manager has a corresponding Helm chart in `charts/tenancy-manager` folder. The CI integration for this repository will publish this Helm charts to the Edge Orchestrator Release Service OCI registry upon merge to `main` branch.

- Tenancy-Manager is deployed to the Edge Orchestrator using this Helm chart, whose lifecycle is in turn managed by Argo CD.

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

1. Edit and build the code with `make go-build` command

2. Run linters with `make lint` command

3. Run the unit tests with `make test` command

4. Build the code with `make build` command

## Contribute

We welcome contributions from the community! To contribute, please open a pull request to have your changes reviewed and merged into the main. We encourage you to add appropriate unit tests and e2e tests if your contribution introduces a new feature.

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
- `build` to build the tenancy-manager Docker container
- `release` to publish the built tenancy-manager Docker container to pre-defined docker container registry. This registry is set in a env variable (TENANCY_MANAGER_DOCKER_IMAGE_OEP) in tenancy-manager/Makefile
