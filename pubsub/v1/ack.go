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
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	ErrAckTimeout  = errors.New("ack has timed out")
	ErrMsgNotFound = errors.New("message not found")
)

// acknowledgementManager control the messages acknowledgement from the server.
type acknowledgementManager struct {
	pendingAcks map[string]chan error
	mu          *sync.RWMutex
}

// getAwaiter adds the messageID for pending acks.
// and returns the message awaiter and a discard function that should be called as soon as the message was ack'ed.
func (m *acknowledgementManager) getAwaiter() (messageID string, await func(context.Context) error, discard func()) {
	msgID := uuid.New().String()

	m.mu.Lock()
	defer m.mu.Unlock()
	ackChan := make(chan error, 1)
	m.pendingAcks[messageID] = ackChan

	awaiter := func(ctx context.Context) error {
		select {
		case err := <-ackChan:
			return err
		case <-ctx.Done():
			return ErrAckTimeout
		}
	}

	disposer := func() {
		m.mu.Lock()
		close(ackChan)
		delete(m.pendingAcks, messageID)
		m.mu.Unlock()
	}

	return msgID, awaiter, disposer
}

// ack acknowledge a message
func (m *acknowledgementManager) ack(messageID string, err error) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	c, ok := m.pendingAcks[messageID]
	if !ok {
		return ErrMsgNotFound
	}

	c <- err
	return nil
}
