# Stagerunner

A pipeline execution service.

Stagerunner is providing a REST API server and a client to manage pipelines and pipeline runs.

## API

The API is defined in [http/api.go](http/api.go). The following endpoints are available:

- `GET /pipelines`: List all pipelines
- `POST /pipelines`: Create a pipeline
- `GET /pipelines/{id}`: Get a pipeline
- `PUT /pipelines/{id}`: Update a pipeline
- `DELETE /pipelines/{id}`: Delete a pipeline
- `POST /pipelines/{id}/trigger`: Trigger a pipeline run
- `GET /runs`: List all pipeline runs
- `GET /runs/{run_id}`: Get a pipeline run

You can use curl or the CLI client to interact with the API server.

### Example curl requests
```
# create a pipeline
curl -X POST http://localhost:8080/pipelines -H "Authorization: some-token" -d '{"name": "test1", "repository": "repo1", "stages": {"run_stage": {"command": "some command"}, "build_stage": {"dockerfile_path": "Dockerfile"}, "deploy_stage": {"cluster_name": "staging_eks_cluster", "manifest_path": "k8s/"}}}'

# get a pipeline
curl -X GET http://localhost:8080/pipelines/9cab004d-07c4-4637-a999-a96ddaddbfe6 -H "Authorization: some-token"
```

The API server is storing pipelines and pipeline runs in memory. A simple concurrent execution engine is "executing" (just printing logs) the pipelines with a configurable number of concurrent workers. A failure probability is configurable to simulate failure handling.

The following assumptions are made:

- Users can only add the following stages to a pipeline: `run`, `build`, `deploy`
  - `run` stage: can contain an arbitrary command to run tests, linting, etc.
  - `build` stage: needs to contain a Dockerfile path to build a docker image
  - `deploy` stage: needs to contain a cluster name and a manifest path to deploy to a kubernetes cluster
- We just require a token to authenticate requests to the API server for demonstration purposes. No fancy auth or RBAC is implemented.
- Only one pipeline run can be executing for a pipeline at a time. Other runs for the same pipeline are queued up.

## Design

I tried to split the code into different packages and files to separate concerns. The interfaces and types are defined to be composable to make alternative implementations and testing easy. A few example middlewares have been provided, but I didn't find time to all of their features.

- `domain`: contains the domain logic, like the store, pipeline and pipeline run types and interfaces, and the executor
- `store`: contains an in-memory implementation of the store interface
- `http`: contains the REST API server and client
- `cmd`: contains the CLI implementation for starting the server and running client commands

## Building

```
go build -o stagerunner cmd/*.go
```

## Running

The cli is providing a server and client. Use the `-h` flag to see the available commands and flags.

```
./stagerunner -h
```

To start the server, run:

```
./stagerunner server
```

In another terminal, you can run the client to create a pipelines and trigger pipeline runs:

```
# create a pipeline
./stagerunner client --token "secret" create \
'{"name": "test1", "repository": "repo1", "stages": {"run_stage": {"command": "some command"}, "build_stage": {"dockerfile_path": "Dockerfile"}, "deploy_stage": {"cluster_name": "staging_eks_cluster", "manifest_path": "k8s/"}}}'

Pipeline created. ID: e2c90447-03e4-45a5-a41f-650394c5d2d1

# trigger a pipeline run
./stagerunner client --token "secret" trigger e2c90447-03e4-45a5-a41f-650394c5d2d1 main

Pipeline triggered. Run ID: 9cab004d-07c4-4637-a999-a96ddaddbfe6

# get run status
./stagerunner client --token "secret" get-run 9cab004d-07c4-4637-a999-a96ddaddbfe6

Run: ID: 9cab004d-07c4-4637-a999-a96ddaddbfe6
  PipelineID: e2c90447-03e4-45a5-a41f-650394c5d2d1
  GitRef: main
  Status: success
  CreatedAt: 2025-01-02 02:34:34.929705 +0100 CET
  UpdatedAt: 2025-01-02 02:34:50.006764 +0100 CET
  [...]
```

For convenience I provided a Makefile to run the server and some example client commands:

```
# run the tests
make test

# build the binary
make build

# start the server
make server

# run some example client commands (in another terminal)
make demo
```