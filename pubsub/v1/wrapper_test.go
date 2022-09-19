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
	"io"
	"sync/atomic"
	"testing"

	contribPubSub "github.com/dapr/components-contrib/pubsub"
	proto "github.com/dapr/dapr/pkg/proto/components/v1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRecvResp struct {
	msg *proto.PullMessagesRequest
	err error
}
type fakeStream struct {
	proto.PubSub_PullMessagesServer
	sendCalled   atomic.Int64
	onSendCalled func(*proto.PullMessagesResponse)
	sendErr      error
	recvCalled   atomic.Int64
	recvChan     chan *fakeRecvResp
	ctx          context.Context
}

func (f *fakeStream) Send(req *proto.PullMessagesResponse) error {
	f.sendCalled.Add(1)
	if f.onSendCalled != nil {
		f.onSendCalled(req)
	}
	return f.sendErr
}
func (f *fakeStream) Recv() (*proto.PullMessagesRequest, error) {
	f.recvCalled.Add(1)
	resp := <-f.recvChan
	return resp.msg, resp.err
}

func (f *fakeStream) Context() context.Context {
	if f.ctx == nil {
		return context.Background()
	}
	return f.ctx
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
	t.Run("pullmessages should return an error when can't receive first message", func(t *testing.T) {
		fakeErr := errors.New("fakeErr")
		ps := &pubsub{}
		recvChan := make(chan *fakeRecvResp, 1)
		recvChan <- &fakeRecvResp{
			err: fakeErr,
		}
		close(recvChan)
		stream := &fakeStream{
			recvChan: recvChan,
		}
		assert.Equal(t, fakeErr, ps.PullMessages(stream))
		assert.Equal(t, int64(1), stream.recvCalled.Load())
	})

	t.Run("pullmessages should return an error when first message doesn't contains a topic on it", func(t *testing.T) {
		ps := &pubsub{}
		recvChan := make(chan *fakeRecvResp, 1)
		recvChan <- &fakeRecvResp{
			msg: &proto.PullMessagesRequest{},
		}
		close(recvChan)
		stream := &fakeStream{
			recvChan: recvChan,
		}
		assert.Equal(t, ErrTopicNotSpecified, ps.PullMessages(stream))
		assert.Equal(t, int64(1), stream.recvCalled.Load())
	})
	t.Run("pullmessages should return an error when subscribe returns an error", func(t *testing.T) {
		const fakeTopic = "fake-topic"
		fakeSubsErr := errors.New("fake-subs-err")

		impl := &fakePubSubImpl{
			subscribeErr: fakeSubsErr,
		}
		ps := &pubsub{
			impl: impl,
		}
		recvChan := make(chan *fakeRecvResp, 1)
		recvChan <- &fakeRecvResp{
			msg: &proto.PullMessagesRequest{
				Topic: &proto.Topic{
					Name: fakeTopic,
				},
			},
		}
		close(recvChan)
		stream := &fakeStream{
			recvChan: recvChan,
		}
		assert.Equal(t, fakeSubsErr, ps.PullMessages(stream))
		assert.Equal(t, int64(1), stream.recvCalled.Load())
		assert.Equal(t, int64(1), impl.subscribeCalled.Load())
	})
	t.Run("pullmessages should callback handler when new messages arrive", func(t *testing.T) {
		const fakeTopic = "fake-topic"

		subsChan := make(chan *contribPubSub.NewMessage, 1)
		subsChan <- &contribPubSub.NewMessage{
			Topic: fakeTopic,
		}
		close(subsChan)
		handleCount := atomic.Int64{}
		impl := &fakePubSubImpl{
			subscribeChan: subsChan,
			onHandlerResp: func(_ error) {
				handleCount.Add(1)
			},
			subscribeCtx: context.Background(),
		}
		ps := &pubsub{
			impl: impl,
		}
		recvChan := make(chan *fakeRecvResp, 3)
		recvChan <- &fakeRecvResp{
			msg: &proto.PullMessagesRequest{
				Topic: &proto.Topic{
					Name: fakeTopic,
				},
			},
		}
		msgIDChan := make(chan string, 1)
		go func() {
			recvChan <- &fakeRecvResp{
				msg: &proto.PullMessagesRequest{
					AckMessageId: <-msgIDChan,
				},
			}
			recvChan <- &fakeRecvResp{
				err: io.EOF,
			} // endstream
			close(recvChan)
		}()

		stream := &fakeStream{
			recvChan: recvChan,
			onSendCalled: func(pmr *proto.PullMessagesResponse) {
				msgIDChan <- pmr.Id
				assert.Equal(t, pmr.TopicName, fakeTopic)
			},
		}
		assert.Nil(t, ps.PullMessages(stream))
		assert.Equal(t, int64(3), stream.recvCalled.Load()) // recv topic, recv message, recv eof
		assert.Equal(t, int64(1), stream.sendCalled.Load())
		assert.Equal(t, int64(1), impl.subscribeCalled.Load())
		assert.Equal(t, int64(0), handleCount.Load())
	})
}

func TestPubSub(t *testing.T) {
	t.Run("features should list pubsubfeatures", func(t *testing.T) {
		const fakeFeature = "fake-feature"
		impl := &fakePubSubImpl{
			featuresResp: []contribPubSub.Feature{fakeFeature},
		}
		ps := &pubsub{
			impl: impl,
		}

		resp, err := ps.Features(context.Background(), &proto.FeaturesRequest{})
		require.NoError(t, err)
		assert.Equal(t, int64(1), impl.featuresCalled.Load())
		assert.Len(t, resp.Feature, 1)
	})

	t.Run("publish should call impl publish", func(t *testing.T) {
		impl := &fakePubSubImpl{}
		ps := &pubsub{
			impl: impl,
		}
		_, err := ps.Publish(context.Background(), &proto.PublishRequest{})
		require.NoError(t, err)
		assert.Equal(t, int64(1), impl.publishCalled.Load())
	})
}
