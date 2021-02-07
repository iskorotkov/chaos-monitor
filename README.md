# Chaos Monitor

Chaos Monitor is used to detect pod failures in a Kubernetes cluster.

## Features

- filter pod update events to find pod failures
- ignore pods from ignore list
- ignore deployments from ignore list
- ignore nodes from ignore list

## Env vars

- `APP_NS` - target namespace to monitor
- `APP_LABEL` - pod label to use to determine pod deployment
- `DURATION` - time to run
- `IGNORED_PODS` - list of ignored pods (delimited by `;`)
- `IGNORED_DEPLOYMENTS` - list of ignored deployments (delimited by `;`)
- `IGNORED_NODES` - list of ignored nodes (delimited by `;`)

## Development

To build project:

```shell
go build ./...
```

To run tests:

```shell
go test ./...
```
