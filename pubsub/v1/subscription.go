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

	"github.com/dapr-sandbox/components-go-sdk/internal"
	contribPubSub "github.com/dapr/components-contrib/pubsub"
	proto "github.com/dapr/dapr/pkg/proto/components/v1"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	ErrConsumerTimeout         = errors.New("there is no consumer available for this subscription")
	ErrSendOnUnsubscribedTopic = errors.New("unsubscribe was called by the consumer")
	ErrAckTimeout              = errors.New("ack has timed out")
)

type outstandingMessage struct {
	message *proto.Message
	ack     chan error
}
type subscription struct {
	messagesChan chan *outstandingMessage
	ctx          context.Context
	cancel       context.CancelFunc
}

// send message to the consumer and wait for its acknowledgement.
func (s *subscription) send(ctx context.Context, msg *contribPubSub.NewMessage) error {
	msgID := uuid.New().String()
	message := &proto.Message{
		Id:          msgID,
		Data:        msg.Data,
		Topic:       msg.Topic,
		Metadata:    msg.Metadata,
		ContentType: internal.ZeroIfNil(msg.ContentType),
	}
	ackChan := make(chan error, 1)
	pendingAckMsg := &outstandingMessage{
		message: message,
		ack:     ackChan,
	}
	select {
	case <-ctx.Done():
		return ErrConsumerTimeout
	case <-s.ctx.Done():
		return ErrSendOnUnsubscribedTopic // the topic was unsubscribed
	case s.messagesChan <- pendingAckMsg: // the message was transmitted
		select {
		case <-ctx.Done():
			return ErrAckTimeout
		case err := <-ackChan:
			return err
		}
	}
}

type subsManager struct {
	subscriptions map[string]*subscription
	mu            *sync.RWMutex
}

// newSubscription creates a new subscription.
func (s *subsManager) newSubscription() (string, *subscription) {
	s.mu.Lock()
	defer s.mu.Unlock()
	subscriptionID := uuid.New().String()
	ctx, cancel := context.WithCancel(context.Background())
	subscription := &subscription{
		messagesChan: make(chan *outstandingMessage),
		ctx:          ctx,
		cancel:       cancel,
	}

	s.subscriptions[subscriptionID] = subscription
	return subscriptionID, subscription
}

// getSubscription returns the subscription by its ID.
func (s *subsManager) getSubscription(subsID string) (*subscription, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	subs, ok := s.subscriptions[subsID]
	return subs, ok
}

// unsubscribe closes the channel and cleanup the subscriptions, if subsID does not exists it is a no-op.
func (s *subsManager) unsubscribe(subsID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	subscription, ok := s.subscriptions[subsID]
	if !ok {
		return
	}
	subscription.cancel() // avoid send on closed channel
	close(subscription.messagesChan)
	delete(s.subscriptions, subsID)
}
