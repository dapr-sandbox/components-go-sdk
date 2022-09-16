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
	"sync"
	"time"

	"github.com/dapr-sandbox/components-go-sdk/internal"
	contribPubSub "github.com/dapr/components-contrib/pubsub"
	proto "github.com/dapr/dapr/pkg/proto/components/v1"
	"github.com/pkg/errors"
)

var ErrAckTimeout = errors.New("ack has timed out")

// threadSafeStream wraps a grpc stream with locks to permit send and recv in multiples goroutines.
type threadSafeStream interface {
	send(*proto.PullMessageResponse) error
	recv() (*proto.PullMessagesRequest, error)
}

// grpcThreadSafeStream wraps a grpcStream for thread safe operations.
// As per documentation is unsafe to call recv OR send in multiples goroutines
// https://github.com/grpc/grpc-go/blob/master/Documentation/concurrency.md#streams
// it is safe to call recv AND send in different goroutines.
type grpcThreadSafeStream struct {
	recvLock *sync.Mutex
	sendLock *sync.Mutex
	stream   proto.PubSub_PullMessagesServer
}

func (s *grpcThreadSafeStream) send(msg *proto.PullMessageResponse) error {
	s.sendLock.Lock()
	defer s.sendLock.Unlock()
	return s.stream.Send(msg)
}

func (s *grpcThreadSafeStream) recv() (*proto.PullMessagesRequest, error) {
	s.recvLock.Lock()
	defer s.recvLock.Unlock()
	return s.stream.Recv()
}

// ackLoop starts a receive loop for messages ack
func ackLoop(ctx context.Context, manager *acknowledgementManager, stream threadSafeStream) {
	for {
		if ctx.Err() != nil {
			return
		}

		ack, err := stream.recv()
		if err == io.EOF {
			return
		}

		if err != nil {
			// FIXME
			// should we continue without sleep ?
			// should we stop and cancel everything?
			pubsubLogger.Errorf("error %v when trying to receive ack, sleeping 5 seconds", err)
			time.Sleep(time.Second * 5)
			continue
		}

		var ackError error

		if ack.AckError != nil {
			ackError = errors.New(ack.AckError.Message)
		}
		manager.ack(ack.AckMessageId, ackError)
	}
}

// handleFor creates a message handler for the given stream.
// it starts an ack loop in background as a separate goroutine
func handlerFor(stream proto.PubSub_PullMessagesServer) (handler func(ctx context.Context, msg *contribPubSub.NewMessage) error) {
	streamCtx := stream.Context()
	tfStream := &grpcThreadSafeStream{
		stream:   stream,
		recvLock: &sync.Mutex{},
		sendLock: &sync.Mutex{},
	}
	ackManager := &acknowledgementManager{
		pendingAcks: map[string]chan error{},
		mu:          &sync.RWMutex{},
	}
	go ackLoop(streamCtx, ackManager, tfStream)
	return func(ctx context.Context, contribMsg *contribPubSub.NewMessage) error {
		msgID, pendingAck, cleanup := ackManager.get()
		defer cleanup()

		msg := &proto.PullMessageResponse{
			Data:        contribMsg.Data,
			Topic:       contribMsg.Topic,
			Metadata:    contribMsg.Metadata,
			ContentType: internal.ZeroIfNil(contribMsg.ContentType),
			Id:          msgID,
		}

		// in case of message can't be sent it does not mean that the sidecar didn't receive the message
		// it only means that the component wasn't able to receive the response back
		// it could leads in messages being acknowledged without even being pending first
		// we should ignore this since it will be probably retried by the underlying component.
		err := tfStream.send(msg)
		if err != nil {
			return errors.Wrapf(err, "error when sending message %s to consumer on topic %s", msg.Id, msg.Topic)
		}

		select {
		case err := <-pendingAck:
			return err
		case <-ctx.Done():
			return ErrAckTimeout
		}
	}
}
