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

package bindings

import (
	"context"
	"io"
	"time"

	"github.com/dapr-sandbox/components-go-sdk/internal"
	contribBindings "github.com/dapr/components-contrib/bindings"
	proto "github.com/dapr/dapr/pkg/proto/components/v1"
	"github.com/pkg/errors"
)

type handleResponse struct {
	data []byte
	err  error
}

var ErrAckTimeout = errors.New("ack has timed out")

// ackLoop starts an active ack loop reciving acks from client.
func ackLoop(streamCtx context.Context, tfStream internal.ThreadSafeStream[proto.ReadResponse, proto.ReadRequest], ackManager *internal.AcknowledgementManager[*handleResponse]) error {
	for {
		if streamCtx.Err() != nil {
			return streamCtx.Err()
		}
		ack, err := tfStream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			// FIXME
			// should we continue without sleep ?
			// should we stop and cancel everything?
			inputLogger.Errorf("error %v when trying to receive ack, sleeping 5 seconds", err)
			time.Sleep(time.Second * 5)
			continue
		}

		var ackError error

		if ack.ResponseError != nil {
			ackError = errors.New(ack.ResponseError.Message)
		}
		if err := ackManager.Ack(ack.MessageId, &handleResponse{
			data: ack.ResponseData,
			err:  ackError,
		}); err != nil {
			inputLogger.Warnf("error %v when trying to notify ack", err)
		}
	}
}

// handler build a pubsub handler using the given threadsafe stream and the ack manager.
func handler(tfStream internal.ThreadSafeStream[proto.ReadResponse, proto.ReadRequest], ackManager *internal.AcknowledgementManager[*handleResponse]) contribBindings.Handler {
	return func(ctx context.Context, contribMsg *contribBindings.ReadResponse) ([]byte, error) {
		msgID, pendingAck, cleanup := ackManager.Get()
		defer cleanup()

		msg := &proto.ReadResponse{
			Data:        contribMsg.Data,
			Metadata:    contribMsg.Metadata,
			ContentType: internal.ZeroValueIfNil(contribMsg.ContentType),
			MessageId:   msgID,
		}

		// in case of message can't be sent it does not mean that the sidecar didn't receive the message
		// it only means that the component wasn't able to receive the response back
		// it could leads in messages being acknowledged without even being pending first
		// we should ignore this since it will be probably retried by the underlying component.
		err := tfStream.Send(msg)
		if err != nil {
			return nil, errors.Wrapf(err, "error when sending message %s", msg.MessageId)
		}

		select {
		case resp := <-pendingAck:
			return resp.data, resp.err
		case <-ctx.Done():
			return nil, ErrAckTimeout
		}
	}
}

// streamReader creates a message handler for the given stream.
func streamReader(stream proto.InputBinding_ReadServer) (bindingsHandler contribBindings.Handler, acknLoop func() error) {
	tfStream := internal.NewGRPCThreadSafeStream[proto.ReadResponse, proto.ReadRequest](stream)
	ackManager := internal.NewAckManager[*handleResponse]()
	return handler(tfStream, ackManager), func() error {
		return ackLoop(stream.Context(), tfStream, ackManager)
	}
}
