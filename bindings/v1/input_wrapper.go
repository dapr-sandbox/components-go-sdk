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

	"github.com/dapr/components-contrib/bindings"
	"github.com/dapr/components-contrib/metadata"
	proto "github.com/dapr/dapr/pkg/proto/components/v1"

	"github.com/dapr/kit/logger"

	"google.golang.org/grpc"
)

var inputLogger = logger.NewLogger("inputbinding-component")

type inputBinding struct {
	proto.UnimplementedInputBindingServer
	getInstance func(context.Context) InputBinding
}

func (in *inputBinding) Init(ctx context.Context, req *proto.InputBindingInitRequest) (*proto.InputBindingInitResponse, error) {
	return &proto.InputBindingInitResponse{}, in.getInstance(ctx).Init(bindings.Metadata{
		Base: metadata.Base{
			Properties: req.Metadata.Properties,
		},
	})
}

func (in *inputBinding) Read(stream proto.InputBinding_ReadServer) error {
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	handler, startAckLoop := streamReader(stream)

	err := in.getInstance(ctx).Read(ctx, handler)
	if err != nil {
		return err
	}

	return startAckLoop()
}

func (in *inputBinding) Ping(context.Context, *proto.PingRequest) (*proto.PingResponse, error) {
	return &proto.PingResponse{}, nil
}

// RegisterInput the inputbinding implementation for the component gRPC service.
func RegisterInput(server *grpc.Server, getInstance func(context.Context) InputBinding) {
	inputBinding := &inputBinding{
		getInstance: getInstance,
	}
	proto.RegisterInputBindingServer(server, inputBinding)
}
