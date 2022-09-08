# Dapr components go SDK POC

This repository is a POC of a SDK for [Dapr gRPC Components (a.k.a pluggable components)](https://github.com/dapr/dapr/issues/4925)

## Running examples

Start by running `./run.sh` inside `/examples` folder. It will start the daprd runtime with pluggable components version + default in memory state store implementation from components-contrib, use `./run.sh redis` to run redis instead.

> Run `ARGS=--no-cache ./run.sh` if you want to rebuild the docker image.

See it working:

```sh
curl -X POST -H "Content-Type: application/json" -d '[{ "key": "name", "value": "Bruce Wayne", "metadata": { "ttlInSeconds": "60"}}]' http://localhost:3500/v1.0/state/default
```

```sh
curl http://localhost:3500/v1.0/state/default/name
```

## Implementing a Pluggable State Store component

To create your own implementation:

1. Create a new folder under `/examples`
2. Implement a stateStore using the sdk
3. create a `component.yml` (copying from other sources and changing the component-specific metadata)
4. Run `./run.sh your_folder_goes_here`

This will build your component and bootstrap the dapr runtime with the default options.

## Getting started

Creating a new component is nothing more than implement a [StateStore](https://github.com/dapr/components-contrib/blob/master/state/store.go#L23) interface and Run the dapr component server.

```golang
package main

import (
	dapr "github.com/dapr-sandbox/components-go-sdk"
)

func main() {
	dapr.MustRun(dapr.UseStateStore(MyComponentGoesHere{}))
}
```
