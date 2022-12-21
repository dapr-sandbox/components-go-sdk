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

package internal

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// AcknowledgementManager control the messages acknowledgement from the server.
type AcknowledgementManager[TAckResult any] struct {
	pendingAcks    map[string]chan TAckResult
	mu             *sync.RWMutex
	ackTimeoutFunc func() <-chan time.Time
}

func NewAckManager[TAckResult any]() *AcknowledgementManager[TAckResult] {
	return &AcknowledgementManager[TAckResult]{
		pendingAcks: map[string]chan TAckResult{},
		mu:          &sync.RWMutex{},
		ackTimeoutFunc: func() <-chan time.Time {
			return time.After(time.Second)
		},
	}
}

// Pending returns a copy of pending acks list
func (m *AcknowledgementManager[TAckResult]) Pending() []chan TAckResult {
	m.mu.RLock()
	defer m.mu.RUnlock()
	pendings := make([]chan TAckResult, len(m.pendingAcks))
	idx := 0
	for _, ack := range m.pendingAcks {
		pendings[idx] = ack
		idx++
	}
	return pendings
}

// Get generate a new messageID for Get acks and returns the ack chan.
func (m *AcknowledgementManager[TAckResult]) Get() (messageID string, ackChan chan TAckResult, cleanup func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	msgID := uuid.New().String()

	ackChan = make(chan TAckResult, 1)
	m.pendingAcks[msgID] = ackChan

	return msgID, ackChan, func() {
		m.mu.Lock()
		delete(m.pendingAcks, msgID)
		close(ackChan)
		m.mu.Unlock()
	}
}

// Ack acknowledge a message
func (m *AcknowledgementManager[TAckResult]) Ack(messageID string, result TAckResult) error {
	m.mu.RLock()
	c, ok := m.pendingAcks[messageID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("message %s not found or not specified", messageID)
	}

	select {
	// wait time for outstanding acks
	// that should be instantaneous as the channel is bufferized size 1
	// if this operation takes longer than the waitFunc (defaults to 1s)
	// it probably means that no consumer is waiting for the message ack
	// or it could be duplicated ack for the same message.
	case <-m.ackTimeoutFunc():
		return nil
	case c <- result:
		return nil
	}
}
