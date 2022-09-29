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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestPool(t *testing.T) {
	t.Run("pool should use default instance when metadata is not present", func(t *testing.T) {
		called := 0
		fakeFactory := func() int {
			called++
			return 0
		}
		factory := poolFor(fakeFactory)
		factory(context.TODO())
		assert.Equal(t, 1, called)
		factory(context.TODO())
		assert.Equal(t, 1, called)
	})
	t.Run("pool should use default instance when metadata is present but no instance id is provided", func(t *testing.T) {
		called := 0
		fakeFactory := func() int {
			called++
			return 0
		}
		factory := poolFor(fakeFactory)

		factory(metadata.NewIncomingContext(context.TODO(), metadata.Pairs("a", "b")))
		assert.Equal(t, 1, called)
		factory(metadata.NewIncomingContext(context.TODO(), metadata.Pairs("a", "b")))
		assert.Equal(t, 1, called)
	})
	t.Run("pool should reuse created instances when instance id is provided", func(t *testing.T) {
		called := 0
		fakeFactory := func() int {
			called++
			return 0
		}
		factory := poolFor(fakeFactory)

		factory(metadata.NewIncomingContext(context.TODO(), metadata.Pairs(metadataInstanceID, "x")))
		assert.Equal(t, 1, called)
		factory(metadata.NewIncomingContext(context.TODO(), metadata.Pairs(metadataInstanceID, "x")))
		assert.Equal(t, 1, called)
	})
}
