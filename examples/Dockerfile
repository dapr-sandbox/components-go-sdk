# syntax=docker/dockerfile:1

## Build
FROM golang:1.20-buster AS build

ARG COMPONENT=state.memory

WORKDIR /sdk

## Copying sdk files
COPY internal internal
COPY *.go ./
COPY state state
COPY bindings bindings
COPY pubsub pubsub
COPY go.mod ./
COPY go.sum ./
RUN go mod download

WORKDIR /app

COPY examples/${COMPONENT}/go.mod ./
COPY examples/${COMPONENT}/go.sum ./
RUN go mod edit -replace github.com/dapr-sandbox/components-go-sdk=/sdk/
RUN go mod download

COPY examples/${COMPONENT}/*.go ./

RUN CGO_ENABLED=0 go build -o /component

## Deploy
FROM gcr.io/distroless/static-debian11

WORKDIR /

COPY --from=build /component /component

ENTRYPOINT ["/component"]