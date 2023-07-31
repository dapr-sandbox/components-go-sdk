# Running Examples on KinD

Prerequisites:

- kind
- dapr cli
- kubectl

## Setup local registry and kind

```shell
./setup_kind.sh
```

## Run any component on kubernetes

> Note: Some components may require a dependency like redis/kafka and others. So you need to install it first.

```shell
./run.sh state.memory
```

## See it working

Forward local 3500 port to remote daprd port and make requests:

```sh
curl -X POST -H "Content-Type: application/json" -d '[{ "key": "name", "value": "Bruce Wayne", "metadata": { "ttlInSeconds": "60"}}]' http://localhost:3500/v1.0/state/default
```

```sh
curl --silent http://localhost:3500/v1.0/state/default/name |base64 --decode
```
