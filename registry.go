/*
Copyright 2022 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dapr

import (
	"errors"

	"github.com/dapr-sandbox/components-go-sdk/bindings/v1"
	"github.com/dapr-sandbox/components-go-sdk/pubsub/v1"
	"github.com/dapr-sandbox/components-go-sdk/state/v1"
	"github.com/dapr/kit/logger"
	"google.golang.org/grpc"
)

var (
	svcLogger              = logger.NewLogger("dapr-component")
	ErrNoneComponentsFound = errors.New("at least one component service should be defined")
)

type componentsOpts struct {
	useGrpcServer []func(*grpc.Server)
}

var (
	factories = make(map[string]*componentsOpts)
)

type option = func(*componentsOpts)

// WithPubSub adds pubsub factory for the component.
func WithPubSub(factory func() pubsub.PubSub) option {
	return func(cf *componentsOpts) {
		cf.useGrpcServer = append(cf.useGrpcServer, func(s *grpc.Server) {
			pubsub.Register(s, mux(factory))
		})
	}
}

// WithStateStore adds statestore factory for the component.
func WithStateStore(factory func() state.Store) option {
	return func(cf *componentsOpts) {
		cf.useGrpcServer = append(cf.useGrpcServer, func(s *grpc.Server) {
			state.Register(s, mux(factory))
		})
	}
}

// WithInputBinding adds inputbinding factory for the component.
func WithInputBinding(factory func() bindings.InputBinding) option {
	return func(cf *componentsOpts) {
		cf.useGrpcServer = append(cf.useGrpcServer, func(s *grpc.Server) {
			bindings.RegisterInput(s, mux(factory))
		})
	}
}

// WithOutputBinding adds outputbinding factory for the component.
func WithOutputBinding(factory func() bindings.OutputBinding) option {
	return func(cf *componentsOpts) {
		cf.useGrpcServer = append(cf.useGrpcServer, func(s *grpc.Server) {
			bindings.RegisterOutput(s, mux(factory))
		})
	}
}

// validate check options are valid.
// if none component was specified so it will return an error.
func (c *componentsOpts) validate() error {
	if c.useGrpcServer == nil || len(c.useGrpcServer) == 0 {
		return ErrNoneComponentsFound
	}
	return nil
}

// apply applies the options to the given grpcServer.
func (c *componentsOpts) apply(s *grpc.Server) error {
	if err := c.validate(); err != nil {
		return err
	}

	for _, useGrpcServer := range c.useGrpcServer {
		useGrpcServer(s)
	}

	return nil
}

// Register a component with the given name.
func Register(name string, opts ...option) {
	cmpFactories := &componentsOpts{}

	for _, opt := range opts {
		opt(cmpFactories)
	}

	factories[name] = cmpFactories
}
