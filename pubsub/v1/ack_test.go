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

package pubsub

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAckManager(t *testing.T) {
	t.Run("ack manager should add pending when ack get called", func(t *testing.T) {
		manager := &acknowledgementManager{
			pendingAcks: map[string]chan error{},
			mu:          &sync.RWMutex{},
		}
		assert.Empty(t, manager.pendingAcks)
		manager.get()
		assert.NotEmpty(t, manager.pendingAcks)
	})
	t.Run("ack manager should return a bufferized channel when ack get called", func(t *testing.T) {
		manager := &acknowledgementManager{
			pendingAcks: map[string]chan error{},
			mu:          &sync.RWMutex{},
		}
		_, c, _ := manager.get()
		assert.Equal(t, 1, cap(c))
	})
	t.Run("cleanup should delete pending ack and close pending channel", func(t *testing.T) {
		manager := &acknowledgementManager{
			pendingAcks: map[string]chan error{},
			mu:          &sync.RWMutex{},
		}
		assert.Empty(t, manager.pendingAcks)
		_, c, cleanup := manager.get()
		assert.NotEmpty(t, manager.pendingAcks)
		cleanup()
		assert.Empty(t, manager.pendingAcks)
		err, ok := <-c
		assert.Nil(t, err)
		assert.False(t, ok)
	})
	t.Run("ack-ing a message that doesn't exists should return an error", func(t *testing.T) {
		manager := newAckManager()
		assert.NotNil(t, manager.ack("fake-id", nil))
	})
	t.Run("ack-ing a message that exists should return not return error", func(t *testing.T) {
		manager := newAckManager()
		msgID, _, _ := manager.get()
		assert.Nil(t, manager.ack(msgID, nil))
	})
	t.Run("duplicated messages should have timeout when no consumer available", func(t *testing.T) {
		const fakeMessageID = "fakeMessageID"
		unbufferizedChannel := make(chan error)
		timeVal := time.Time{}
		c := make(chan time.Time, 1)
		c <- timeVal
		close(c)
		manager := &acknowledgementManager{
			pendingAcks: map[string]chan error{
				fakeMessageID: unbufferizedChannel,
			},
			ackTimeoutFunc: func() <-chan time.Time {
				return c
			},
			mu: &sync.RWMutex{},
		}

		assert.Len(t, c, 1)
		manager.ack(fakeMessageID, nil)
		assert.Len(t, c, 0)
	})
}
