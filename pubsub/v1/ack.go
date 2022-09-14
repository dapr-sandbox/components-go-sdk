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

	"github.com/pkg/errors"
)

var ErrMsgNotFound = errors.New("message not found")

// acknowledgementManager control the messages acknowledgement from the server.
type acknowledgementManager struct {
	pendingAcks map[string]chan error
	mu          *sync.RWMutex
}

// wait blocks until the message get ack'ed.
func (m *acknowledgementManager) wait(messageID string) error {
	m.mu.Lock()
	ackChan := make(chan error, 1)
	m.pendingAcks[messageID] = ackChan
	m.mu.Unlock()
	return <-ackChan
}

// ack acknowledge a message
func (m *acknowledgementManager) ack(messageID string, err error) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.pendingAcks[messageID]
	if !ok {
		return ErrMsgNotFound
	}

	defer delete(m.pendingAcks, messageID)
	c <- err
	close(c)
	return nil
}
