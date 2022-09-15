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

	"github.com/dapr-sandbox/components-go-sdk/internal"
	contribMetadata "github.com/dapr/components-contrib/metadata"
	contribPubSub "github.com/dapr/components-contrib/pubsub"
	proto "github.com/dapr/dapr/pkg/proto/components/v1"
	"github.com/dapr/kit/logger"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var pubsubLogger = logger.NewLogger("pubsub-component")

var defaultPubSub = &pubsub{
	subsManager: &subsManager{
		subscriptions: make(map[string]*subscription),
		mu:            &sync.RWMutex{},
	},
}

type pubsub struct {
	proto.UnimplementedPubSubServer
	impl        PubSub
	subsManager *subsManager
}

const metadataSubscriptionID = "subscription-id"

var ErrSubscriptionNotFound = status.Errorf(codes.NotFound, "subscription not found or not specified")

func sendLoop(stream proto.PubSub_PullMessagesServer, subscription *subscription) {
	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-subscription.messagesChan:
			if !ok { // closed chan
				return
			}

			err := stream.Send(msg)
			if err == io.EOF { // no more messages
				return
			}

			if err != nil {
				pubsubLogger.Errorf("error when sending message %s to consumer on topic %s", msg.Id, msg.Topic)
			}
		}
	}
}

func ackLoop(stream proto.PubSub_PullMessagesServer, subscription *subscription) error {
	ctx := stream.Context()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		ack, err := stream.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		var ackErr error

		if ack.Error != nil {
			ackErr = errors.New(ack.Error.Message)
		}

		if err := subscription.ackManager.ack(ack.MessageId, ackErr); err != nil {
			pubsubLogger.Errorf("error %v when ack'ing message %s", err, ack.MessageId)
		}
	}
}

// Establishes a stream with the server, which sends messages down to the
// client. The client streams acknowledgements back to the server. The server
// will close the stream and return the status on any error. In case of closed
// connection, the client should re-establish the stream.
// The first trailing metadata MUST contain a `subscription-id: X` to select
// the subscription that should be used for the streaming pull.
func (s *pubsub) PullMessages(stream proto.PubSub_PullMessagesServer) error {
	streamCtx := stream.Context()
	metadata, ok := metadata.FromIncomingContext(streamCtx)
	if !ok {
		return ErrSubscriptionNotFound
	}

	subscriptionIDs := metadata.Get(metadataSubscriptionID)
	if len(subscriptionIDs) == 0 {
		return ErrSubscriptionNotFound
	}

	subscription, ok := s.subsManager.getSubscription(subscriptionIDs[0])

	if !ok || subscription == nil {
		return ErrSubscriptionNotFound
	}

	// consume messages goroutine
	go sendLoop(stream, subscription)

	// ack loop
	return ackLoop(stream, subscription)
}

func (s *pubsub) Unsubscribe(_ context.Context, req *proto.UnsubscribeRequest) (*proto.UnsubscribeResponse, error) {
	s.subsManager.unsubscribe(req.SubscriptionId)
	return &proto.UnsubscribeResponse{}, nil
}

func (s *pubsub) Init(_ context.Context, initReq *proto.PubSubInitRequest) (*proto.PubSubInitResponse, error) {
	return &proto.PubSubInitResponse{}, s.impl.Init(contribPubSub.Metadata{
		Base: contribMetadata.Base{Properties: initReq.Metadata.Properties},
	})
}

func (s *pubsub) Features(context.Context, *proto.FeaturesRequest) (*proto.FeaturesResponse, error) {
	features := &proto.FeaturesResponse{
		Feature: internal.Map(s.impl.Features(), func(f contribPubSub.Feature) string {
			return string(f)
		}),
	}

	return features, nil
}

func (s *pubsub) Publish(_ context.Context, req *proto.PublishRequest) (*proto.PublishResponse, error) {
	return &proto.PublishResponse{}, s.impl.Publish(&contribPubSub.PublishRequest{
		Data:        req.Data,
		PubsubName:  req.PubsubName,
		Topic:       req.Topic,
		Metadata:    req.Metadata,
		ContentType: &req.ContentType,
	})
}

func (s *pubsub) Subscribe(ctx context.Context, req *proto.SubscribeRequest) (*proto.SubscribeResponse, error) {
	id, subscription := s.subsManager.newSubscription()
	return &proto.SubscribeResponse{
			SubscriptionId: id,
		}, s.impl.Subscribe(subscription.ctx, contribPubSub.SubscribeRequest{
			Topic:    req.Topic,
			Metadata: req.Metadata,
		}, subscription.send)
}

func (s *pubsub) Ping(context.Context, *proto.PingRequest) (*proto.PingResponse, error) {
	return &proto.PingResponse{}, nil
}

// Register the pubsub implementation for the component gRPC service.
func Register(server *grpc.Server, ps PubSub) {
	defaultPubSub.impl = ps
	proto.RegisterPubSubServer(server, defaultPubSub)
}
