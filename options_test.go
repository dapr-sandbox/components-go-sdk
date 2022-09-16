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
	"testing"

	"github.com/dapr-sandbox/components-go-sdk/pubsub/v1"
	"github.com/dapr-sandbox/components-go-sdk/state/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type fakePubSub struct {
	pubsub.PubSub
}

type fakeStateStore struct {
	state.Store
}

func TestOptions(t *testing.T) {
	t.Run("validate should return an error when none backed component is specified", func(t *testing.T) {
		opts := &componentOpts{}
		assert.NotNil(t, opts.validate())
	})

	t.Run("validate should not return an error when at least one component is specified", func(t *testing.T) {
		opts := &componentOpts{
			useGrpcServer: []func(*grpc.Server){
				func(*grpc.Server) {},
			},
		}
		assert.Nil(t, opts.validate())
	})

	t.Run("apply should return an error if validate returns an error", func(t *testing.T) {
		opts := &componentOpts{}
		assert.NotNil(t, opts.apply(&grpc.Server{}))
	})

	t.Run("usePubSub should add a new useGrpcServer callback", func(t *testing.T) {
		opts := &componentOpts{}
		opt := UsePubSub(&fakePubSub{})
		opt(opts)
		assert.Len(t, opts.useGrpcServer, 1)
	})

	t.Run("useStateStore should add a new useGrpcServer callback", func(t *testing.T) {
		opts := &componentOpts{}
		opt := UseStateStore(&fakeStateStore{})
		opt(opts)
		assert.Len(t, opts.useGrpcServer, 1)
	})
}
