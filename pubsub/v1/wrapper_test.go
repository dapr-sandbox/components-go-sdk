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
	"sync/atomic"
	"testing"

	contribPubSub "github.com/dapr/components-contrib/pubsub"
	proto "github.com/dapr/dapr/pkg/proto/components/v1"
)

type fakeStream struct {
	proto.PubSub_PullMessagesServer
	sendCalled   atomic.Int64
	onSendCalled func(*proto.PullMessagesResponse)
	sendErrChan  chan error
}

func (f *fakeStream) Send(req *proto.PullMessagesResponse) error {
	f.sendCalled.Add(1)
	return <-f.sendErrChan

}

type fakePubSubImpl struct {
	PubSub
	initCalled        atomic.Int64
	onInitCalled      func(contribPubSub.Metadata)
	initErr           error
	featuresCalled    atomic.Int64
	featuresResp      []contribPubSub.Feature
	publishCalled     atomic.Int64
	onPublishCalled   func(*contribPubSub.PublishRequest)
	publishErr        error
	subscribeCalled   atomic.Int64
	subscribeErr      error
	onSubscribeCalled func(contribPubSub.SubscribeRequest)
	subscribeCtx      context.Context
	subscribeChan     chan *contribPubSub.NewMessage
	onHandlerResp     func(error)
}

func (f *fakePubSubImpl) Init(metadata contribPubSub.Metadata) error {
	f.initCalled.Add(1)
	if f.onInitCalled != nil {
		f.onInitCalled(metadata)
	}
	return f.initErr
}
func (f *fakePubSubImpl) Features() []contribPubSub.Feature {
	f.featuresCalled.Add(1)
	return f.featuresResp
}
func (f *fakePubSubImpl) Publish(req *contribPubSub.PublishRequest) error {
	f.publishCalled.Add(1)
	if f.onPublishCalled != nil {
		f.onPublishCalled(req)
	}
	return f.publishErr
}
func (f *fakePubSubImpl) Subscribe(ctx context.Context, req contribPubSub.SubscribeRequest, handler contribPubSub.Handler) error {
	f.subscribeCalled.Add(1)
	if f.onSubscribeCalled != nil {
		f.onSubscribeCalled(req)
	}
	if f.subscribeChan != nil {
		go func() {
			for msg := range f.subscribeChan {
				err := handler(f.subscribeCtx, msg)
				if f.onHandlerResp != nil {
					f.onHandlerResp(err)
				}
			}
		}()
	}
	return f.subscribeErr
}

func TestPubSubPullMessages(t *testing.T) {
	t.Run("pullmessages should return an error when can't receive first message", func(t *testing.T) {})
	t.Run("pullmessages should return an error when first message doesn't contains a topic on it", func(t *testing.T) {})
	t.Run("pullmessages should return an error when subscribe returns an error", func(t *testing.T) {})
	t.Run("pullmessages should callback handler when new messages arrive", func(t *testing.T) {})
}
