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
)

var pubsubLogger = logger.NewLogger("dapr-pubsub")

var defaultPubSub = &pubsub{
	ackManager: &acknowledgementManager{
		pendingAcks: map[string]chan error{},
		mu:          &sync.RWMutex{},
	},
}

type pubsub struct {
	proto.UnimplementedPubSubServer
	impl       PubSub
	ackManager *acknowledgementManager
}

func (s *pubsub) AckMessage(stream proto.PubSub_AckMessageServer) error {
	for {
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

		if err := s.ackManager.ack(ack.MessageId, ackErr); err != nil {
			pubsubLogger.Errorf("error %v when ack'ing message %s", err, ack.MessageId)
		}
	}
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

func (s *pubsub) Subscribe(req *proto.SubscribeRequest, stream proto.PubSub_SubscribeServer) error {
	return s.impl.Subscribe(stream.Context(), contribPubSub.SubscribeRequest{
		Topic:    req.Topic,
		Metadata: req.Metadata,
	}, func(ctx context.Context, msg *contribPubSub.NewMessage) error {
		msgID, awaiter, dispose := s.ackManager.getAwaiter()
		defer dispose()

		err := stream.Send(&proto.Message{
			Id:          msgID,
			Data:        msg.Data,
			Topic:       msg.Topic,
			Metadata:    msg.Metadata,
			ContentType: internal.ZeroIfNil(msg.ContentType),
		})
		if err != nil {
			return errors.Wrapf(err, "could not handle the message for topic %s", msg.Topic)
		}
		return awaiter(ctx)
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
