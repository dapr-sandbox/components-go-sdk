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

	"github.com/dapr/kit/logger"

	"github.com/dapr-sandbox/components-go-sdk/state/v1"
	"google.golang.org/grpc"
)

var (
	svcLogger              = logger.NewLogger("dapr-component")
	ErrNoneComponentsFound = errors.New("at least one component service should be defined")
)

type componentOpts struct {
	useGrpcServer []func(*grpc.Server)
}

type Option = func(*componentOpts)

// UseStateStore sets the component state store implementation.
func UseStateStore(stateStore state.Store) Option {
	return func(co *componentOpts) {
		co.useGrpcServer = append(co.useGrpcServer, func(s *grpc.Server) {
			svcLogger.Info("dapr state store was registered")
			state.Register(s, stateStore)
		})
	}
}

// validate check options are valid.
// if none component was specified so it will return an error.
func (c *componentOpts) validate() error {
	if c.useGrpcServer == nil || len(c.useGrpcServer) == 0 {
		return ErrNoneComponentsFound
	}
	return nil
}

// apply applies the options to the given grpcServer.
func (c *componentOpts) apply(s *grpc.Server) error {
	if err := c.validate(); err != nil {
		return err
	}

	for _, useGrpcServer := range c.useGrpcServer {
		useGrpcServer(s)
	}

	return nil
}
