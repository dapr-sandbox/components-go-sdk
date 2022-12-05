# Dapr pluggable components Go SDK

[Pluggable components](https://docs.dapr.io/concepts/components-concept/#built-in-and-pluggable-components) are Dapr components that resides outside Dapr binary and are dynamically registered in runtime.

The SDK provides a better interface to create pluggable components without worrying about underlying communication protocols and connection resiliency.

## Building blocks Interfaces

All building blocks interfaces for this sdk follows the same interfaces provided by the [built-in](https://github.com/dapr/components-contrib) components.

See concrete examples of them:

- [State Store](https://github.com/dapr/components-contrib/tree/master/state)
- [Bindings](https://github.com/dapr/components-contrib/tree/master/bindings)
- [Pub/Sub](https://github.com/dapr/components-contrib/tree/master/pubsub)

## Implementing your State Store

Implement the [State Store interface](https://github.com/dapr/components-contrib/blob/master/state/store.go#L23).

## Implementing your Input/Output Binding

Implement the [Input](https://github.com/dapr/components-contrib/blob/master/bindings/input_binding.go#L24) or/and [Output](https://github.com/dapr/components-contrib/blob/master/bindings/output_binding.go#L24) interface.

## Implementing your Pub/Sub

Implement the [Pub/Sub](https://github.com/dapr/components-contrib/blob/master/pubsub/pubsub.go#L24) interface.

## Registering a pluggable component

Once you have your component implemented, now in order to get your component discovered and [registered](https://docs.dapr.io/operations/components/pluggable-components/pluggable-components-registration/) by Dapr runtime you must register it using the sdk.

Let's say you want to name your state store component as `my-component`, so you have to do the following:

```golang

package main

import (
	dapr "github.com/dapr-sandbox/components-go-sdk"
	"github.com/dapr-sandbox/components-go-sdk/state/v1"
)


func main() {
	dapr.Register("my-component", dapr.WithStateStore(func() state.Store {
		return &MyStateStoreComponent{}
	}))
	dapr.MustRun()
}

```

That's all!

## Starting the component daemon

Component daemon (or server) is a term used for pluggable components processes that runs alongside the Dapr runtime.

You can start your component daemon without any Dapr runtime connecting to it for testing purposes. When running in this mode, you have to make gRPC calls to see your component working.

Run your component using `go run` as you do normally.

## Running examples

Start by running `./run.sh` inside `/examples` folder. It will start the daprd runtime with pluggable components version + default in memory state store implementation from components-contrib, use `./run.sh state.redis` to run redis instead.

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

Optionally you can also add a `docker-compose.dependencies.yml` file and specify container dependencies that will be used when starting your app.
