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

	"github.com/dapr-sandbox/components-go-sdk/internal"
	contribPubSub "github.com/dapr/components-contrib/pubsub"
	proto "github.com/dapr/dapr/pkg/proto/components/v1"
	"github.com/pkg/errors"
)

var ErrAckTimeout = errors.New("ack has timed out")

// ackLoop starts an active ack loop reciving acks from client.
func ackLoop(streamCtx context.Context, tfStream internal.ThreadSafeStream[proto.PullMessagesResponse, proto.PullMessagesRequest], ackManager *internal.AcknowledgementManager[error]) error {
	for {
		if streamCtx.Err() != nil {
			return streamCtx.Err()
		}
		ack, err := tfStream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		var ackError error

		if ack.AckError != nil {
			ackError = errors.New(ack.AckError.Message)
		}
		if err := ackManager.Ack(ack.AckMessageId, ackError); err != nil {
			pubsubLogger.Warnf("error %v when trying to notify ack", err)
		}
	}
}

// handler build a pubsub handler using the given threadsafe stream and the ack manager.
func handler(tfStream internal.ThreadSafeStream[proto.PullMessagesResponse, proto.PullMessagesRequest], ackManager *internal.AcknowledgementManager[error]) contribPubSub.Handler {
	return func(ctx context.Context, contribMsg *contribPubSub.NewMessage) error {
		msgID, pendingAck, cleanup := ackManager.Get()
		defer cleanup()

		msg := &proto.PullMessagesResponse{
			Data:        contribMsg.Data,
			TopicName:   contribMsg.Topic,
			Metadata:    contribMsg.Metadata,
			ContentType: internal.ZeroValueIfNil(contribMsg.ContentType),
			Id:          msgID,
		}

		// in case of message can't be sent it does not mean that the sidecar didn't receive the message
		// it only means that the component wasn't able to receive the response back
		// it could leads in messages being acknowledged without even being pending first
		// we should ignore this since it will be probably retried by the underlying component.
		err := tfStream.Send(msg)
		if err != nil {
			return errors.Wrapf(err, "error when sending message %s to consumer on topic %s", msg.Id, msg.TopicName)
		}

		select {
		case err := <-pendingAck:
			return err
		case <-ctx.Done():
			return ErrAckTimeout
		}
	}
}

// pullFor creates a message handler for the given stream.
func pullFor(stream proto.PubSub_PullMessagesServer) (pubsubHandler contribPubSub.Handler, acknLoop func() error) {
	tfStream := internal.NewGRPCThreadSafeStream[proto.PullMessagesResponse, proto.PullMessagesRequest](stream)
	ackManager := internal.NewAckManager[error]()
	return handler(tfStream, ackManager), func() error {
		return ackLoop(stream.Context(), tfStream, ackManager)
	}
}
