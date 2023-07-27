/*
Copyright 2023 The Dapr Authors
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

package secretstores

import (
	"context"

	contribMetadata "github.com/dapr/components-contrib/metadata"
	contribSecretStores "github.com/dapr/components-contrib/secretstores"

	"github.com/dapr-sandbox/components-go-sdk/internal"

	proto "github.com/dapr/dapr/pkg/proto/components/v1"

	"google.golang.org/grpc"
)

type secretStore struct {
	getInstance func(context.Context) SecretStore
}

func (s *secretStore) Init(ctx context.Context, initReq *proto.SecretStoreInitRequest) (*proto.SecretStoreInitResponse, error) {
	return &proto.SecretStoreInitResponse{}, s.getInstance(ctx).Init(ctx, contribSecretStores.Metadata{
		Base: contribMetadata.Base{Properties: initReq.Metadata.Properties},
	})
}

func (s *secretStore) Features(ctx context.Context, _ *proto.FeaturesRequest) (*proto.FeaturesResponse, error) {
	features := &proto.FeaturesResponse{
		Features: internal.Map(s.getInstance(ctx).Features(), func(f contribSecretStores.Feature) string {
			return string(f)
		}),
	}

	return features, nil
}

func toGetRequest(req *proto.GetSecretRequest) contribSecretStores.GetSecretRequest {
	return contribSecretStores.GetSecretRequest{
		Name:     req.Key,
		Metadata: req.Metadata,
	}
}

func (s *secretStore) Get(ctx context.Context, req *proto.GetSecretRequest) (*proto.GetSecretResponse, error) {
	resp, err := s.getInstance(ctx).GetSecret(ctx, toGetRequest(req))
	if err != nil {
		return nil, err
	}
	return &proto.GetSecretResponse{
		Data: resp.Data,
	}, nil
}

func toBulkGetRequest(req *proto.BulkGetSecretRequest) contribSecretStores.BulkGetSecretRequest {
	return contribSecretStores.BulkGetSecretRequest{
		Metadata: req.Metadata,
	}
}

func (s *secretStore) BulkGet(ctx context.Context, req *proto.BulkGetSecretRequest) (*proto.BulkGetSecretResponse, error) {
	resp, err := s.getInstance(ctx).BulkGetSecret(ctx, toBulkGetRequest(req))
	if err != nil {
		return &proto.BulkGetSecretResponse{}, err
	}
	items := make(map[string]*proto.SecretResponse, len(resp.Data))
	for k, v := range resp.Data {
		res := &proto.SecretResponse{
			Secrets: v,
		}
		items[k] = res
	}
	return &proto.BulkGetSecretResponse{
		Data: items,
	}, nil
}

func (s *secretStore) Ping(context.Context, *proto.PingRequest) (*proto.PingResponse, error) {
	return &proto.PingResponse{}, nil
}

func Register(server *grpc.Server, getInstance func(context.Context) SecretStore) {
	secretStore := &secretStore{
		getInstance: getInstance,
	}
	proto.RegisterSecretStoreServer(server, secretStore)
}
