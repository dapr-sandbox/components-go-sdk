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

	"github.com/dapr-sandbox/components-go-sdk/internal"
	contribMetadata "github.com/dapr/components-contrib/metadata"
	contribPubSub "github.com/dapr/components-contrib/pubsub"
	proto "github.com/dapr/dapr/pkg/proto/components/v1"
	"github.com/dapr/kit/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var pubsubLogger = logger.NewLogger("pubsub-component")

var defaultPubSub = &pubsub{}

var (
	ErrTopicNotSpecified = status.Errorf(codes.InvalidArgument, "topic should be fulfilled in the very first message")
)

type pubsub struct {
	proto.UnimplementedPubSubServer
	impl PubSub
}

// Establishes a stream with the server, which sends messages down to the
// client. The client streams acknowledgements back to the server. The server
// will close the stream and return the status on any error. In case of closed
// connection, the client should re-establish the stream.
// The first message MUST contain a `subscription` on it
// that should be used for the entire streaming pull.
func (s *pubsub) PullMessages(stream proto.PubSub_PullMessagesServer) error {
	fstMessage, err := stream.Recv()
	if err != nil {
		return err
	}

	subscription := fstMessage.GetTopic()

	if subscription == nil {
		return ErrTopicNotSpecified
	}

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	handler, startAckLoop := pullFor(stream)

	err = s.impl.Subscribe(ctx, contribPubSub.SubscribeRequest{
		Topic:    subscription.Name,
		Metadata: subscription.Metadata,
	}, handler)

	if err != nil {
		return err
	}

	return startAckLoop()
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

func (s *pubsub) Ping(context.Context, *proto.PingRequest) (*proto.PingResponse, error) {
	return &proto.PingResponse{}, nil
}

// Register the pubsub implementation for the component gRPC service.
func Register(server *grpc.Server, ps PubSub) {
	defaultPubSub.impl = ps
	proto.RegisterPubSubServer(server, defaultPubSub)
}
