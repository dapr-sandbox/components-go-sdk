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

	proto "github.com/dapr/dapr/pkg/proto/components/v1"
	"github.com/google/uuid"
)

type subsManager struct {
	subscriptions map[string]*proto.SubscribeRequest
	mu            *sync.RWMutex
}

// newSubscription creates a new subscription.
func (s *subsManager) newSubscription(req *proto.SubscribeRequest) (id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	subscriptionID := uuid.New().String()
	s.subscriptions[subscriptionID] = req
	return subscriptionID
}

// getSubscription returns the subscription by its ID.
func (s *subsManager) getSubscription(subsID string) (*proto.SubscribeRequest, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	subs, ok := s.subscriptions[subsID]
	return subs, ok
}
