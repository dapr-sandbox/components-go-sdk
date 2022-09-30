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

	"github.com/dapr-sandbox/components-go-sdk/internal"
	"github.com/dapr/kit/logger"
	"google.golang.org/grpc"

	contribBindings "github.com/dapr/components-contrib/bindings"
	"github.com/dapr/components-contrib/metadata"
	proto "github.com/dapr/dapr/pkg/proto/components/v1"
)

var outputLogger = logger.NewLogger("outputbinding-component")

type outputBinding struct {
	proto.UnimplementedOutputBindingServer
	getInstance func(context.Context) OutputBinding
}

func (out *outputBinding) Init(ctx context.Context, req *proto.OutputBindingInitRequest) (*proto.OutputBindingInitResponse, error) {
	return &proto.OutputBindingInitResponse{}, out.getInstance(ctx).Init(contribBindings.Metadata{
		Base: metadata.Base{
			Properties: req.Metadata.Properties,
		},
	})
}

func (out *outputBinding) Invoke(ctx context.Context, req *proto.InvokeRequest) (*proto.InvokeResponse, error) {
	resp, err := out.getInstance(ctx).Invoke(ctx, &contribBindings.InvokeRequest{
		Data:      req.Data,
		Metadata:  req.Metadata,
		Operation: contribBindings.OperationKind(req.Operation),
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return &proto.InvokeResponse{}, nil
	}
	return &proto.InvokeResponse{
		Data:        resp.Data,
		Metadata:    resp.Metadata,
		ContentType: internal.ZeroValueIfNil(resp.ContentType),
	}, nil
}

func (out *outputBinding) ListOperations(ctx context.Context, _ *proto.ListOperationsRequest) (*proto.ListOperationsResponse, error) {
	return &proto.ListOperationsResponse{
		Operations: internal.Map(out.getInstance(ctx).Operations(), func(op contribBindings.OperationKind) string {
			return string(op)
		}),
	}, nil
}

func (out *outputBinding) Ping(context.Context, *proto.PingRequest) (*proto.PingResponse, error) {
	return &proto.PingResponse{}, nil
}

// RegisterOutput the outputbinding implementation for the component gRPC service.
func RegisterOutput(server *grpc.Server, getInstance func(context.Context) OutputBinding) {
	outputBinding := &outputBinding{
		getInstance: getInstance,
	}
	proto.RegisterOutputBindingServer(server, outputBinding)
}
