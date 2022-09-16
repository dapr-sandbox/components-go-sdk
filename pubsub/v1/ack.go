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
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// acknowledgementManager control the messages acknowledgement from the server.
type acknowledgementManager struct {
	pendingAcks map[string]chan error
	mu          *sync.RWMutex
}

// get generate a new messageID for get acks and returns the ack chan.
func (m *acknowledgementManager) get() (messageID string, ackChan chan error, cleanup func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	msgID := uuid.New().String()

	ackChan = make(chan error, 1)
	m.pendingAcks[msgID] = ackChan

	return msgID, ackChan, func() {
		m.mu.Lock()
		delete(m.pendingAcks, msgID)
		close(ackChan)
		m.mu.Unlock()
	}
}

// ack acknowledge a message
func (m *acknowledgementManager) ack(messageID string, err error) error {
	m.mu.RLock()
	c, ok := m.pendingAcks[messageID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("message %s not found or not specified", messageID)
	}

	select {
	// wait time for outstanding acks
	// acks should be instantaneous as the channel is bufferized size 1
	// if this operation takes longer than 1 second
	// it probably means that no consumer is waiting for the message ack
	// or it could be duplicated ack for the same message.
	case <-time.After(time.Second):
		return nil
	case c <- err:
		return nil
	}
}
