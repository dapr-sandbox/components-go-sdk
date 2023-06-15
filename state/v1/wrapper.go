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

package state

import (
	"context"
	"encoding/json"

	contribMetadata "github.com/dapr/components-contrib/metadata"
	contribState "github.com/dapr/components-contrib/state"
	contribQuery "github.com/dapr/components-contrib/state/query"

	"github.com/dapr-sandbox/components-go-sdk/internal"

	proto "github.com/dapr/dapr/pkg/proto/components/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	consistencyEventual   = "eventual"
	consistencyStrong     = "strong"
	concurrencyLastWrite  = "last-write"
	concurrencyFirstWrite = "first-write"
)

type store struct {
	getInstance func(context.Context) Store
}

//nolint:nosnakecase
var consistencyModels = map[proto.StateOptions_StateConsistency]string{
	proto.StateOptions_CONSISTENCY_EVENTUAL:    consistencyEventual,
	proto.StateOptions_CONSISTENCY_STRONG:      consistencyStrong,
	proto.StateOptions_CONSISTENCY_UNSPECIFIED: "",
}

//nolint:nosnakecase
func toConsistency(consistency proto.StateOptions_StateConsistency) string {
	c, ok := consistencyModels[consistency]
	if !ok {
		return ""
	}
	return c
}

//nolint:nosnakecase
var concurrencyModels = map[proto.StateOptions_StateConcurrency]string{
	proto.StateOptions_CONCURRENCY_FIRST_WRITE: concurrencyFirstWrite,
	proto.StateOptions_CONCURRENCY_LAST_WRITE:  concurrencyLastWrite,
	proto.StateOptions_CONCURRENCY_UNSPECIFIED: "",
}

//nolint:nosnakecase
func toConcurrency(concurrency proto.StateOptions_StateConcurrency) string {
	c, ok := concurrencyModels[concurrency]
	if !ok {
		return ""
	}
	return c
}

func (s *store) Init(ctx context.Context, initReq *proto.InitRequest) (*proto.InitResponse, error) {
	return &proto.InitResponse{}, s.getInstance(ctx).Init(ctx, contribState.Metadata{
		Base: contribMetadata.Base{Properties: initReq.Metadata.Properties},
	})
}

func (s *store) Features(ctx context.Context, _ *proto.FeaturesRequest) (*proto.FeaturesResponse, error) {
	features := &proto.FeaturesResponse{
		Features: internal.Map(s.getInstance(ctx).Features(), func(f contribState.Feature) string {
			return string(f)
		}),
	}

	return features, nil
}

func toDeleteRequest(req *proto.DeleteRequest) *contribState.DeleteRequest {
	return &contribState.DeleteRequest{
		Key: req.Key,
		ETag: internal.IfNotNilP(req.Etag, func(f *proto.Etag) string {
			return f.Value
		}),
		Metadata: req.Metadata,
		Options: internal.IfNotNil(req.Options, func(f *proto.StateOptions) contribState.DeleteStateOption {
			return contribState.DeleteStateOption{
				Concurrency: toConcurrency(f.Concurrency),
				Consistency: toConsistency(f.Consistency),
			}
		}),
	}
}

func (s *store) Delete(ctx context.Context, req *proto.DeleteRequest) (*proto.DeleteResponse, error) {
	return &proto.DeleteResponse{}, s.getInstance(ctx).Delete(ctx, toDeleteRequest(req))
}

func toGetRequest(req *proto.GetRequest) *contribState.GetRequest {
	return &contribState.GetRequest{
		Key:      req.Key,
		Metadata: req.Metadata,
		Options: contribState.GetStateOption{
			Consistency: req.Consistency.String(),
		},
	}
}

func fromGetResponse(res *contribState.GetResponse) *proto.GetResponse {
	return &proto.GetResponse{
		Data: res.Data,
		Etag: internal.IfNotNil(res.ETag, func(etagValue *string) *proto.Etag {
			return &proto.Etag{
				Value: *etagValue,
			}
		}),
		ContentType: internal.IfNotNil(res.ContentType, func(f *string) string {
			return *f
		}),
		Metadata: res.Metadata,
	}
}

func (s *store) Get(ctx context.Context, req *proto.GetRequest) (*proto.GetResponse, error) {
	resp, err := s.getInstance(ctx).Get(ctx, toGetRequest(req))
	return internal.IfNotNil(resp, fromGetResponse), err
}

// dataParser is used to parse content by its content type
var dataParser = map[string]func([]byte) (any, error){
	"application/json": func(b []byte) (any, error) {
		var result any
		return result, json.Unmarshal(b, &result)
	},
}

func toSetRequest(req *proto.SetRequest) *contribState.SetRequest {
	var value any = req.Value
	var contentType *string
	if req.ContentType != "" {
		contentType = &req.ContentType
	}
	if ct, ok := req.Metadata["contentType"]; ok {
		contentType = &ct
	}

	if contentType != nil {
		if parser, ok := dataParser[*contentType]; ok {
			v, _ := parser(req.Value)
			value = v
		}
	}

	return &contribState.SetRequest{
		Key:   req.Key,
		Value: value,
		ETag: internal.IfNotNilP(req.Etag, func(f *proto.Etag) string {
			return f.Value
		}),
		ContentType: contentType,
		Metadata:    req.Metadata,
		Options: internal.IfNotNil(req.Options, func(f *proto.StateOptions) contribState.SetStateOption {
			return contribState.SetStateOption{
				Concurrency: toConcurrency(f.Concurrency),
				Consistency: toConsistency(f.Consistency),
			}
		}),
	}
}

func (s *store) Set(ctx context.Context, req *proto.SetRequest) (*proto.SetResponse, error) {
	return &proto.SetResponse{}, s.getInstance(ctx).Set(ctx, toSetRequest(req))
}

func (s *store) Ping(context.Context, *proto.PingRequest) (*proto.PingResponse, error) {
	return &proto.PingResponse{}, nil
}

// TODO: The default value was added  contribState.BulkStoreOpts{Parallelism: 1}
func (s *store) BulkDelete(ctx context.Context, req *proto.BulkDeleteRequest) (*proto.BulkDeleteResponse, error) {
	return &proto.BulkDeleteResponse{}, s.getInstance(ctx).BulkDelete(ctx, internal.Map(req.Items, func(delReq *proto.DeleteRequest) contribState.DeleteRequest {
		return *toDeleteRequest(delReq)
	}), contribState.BulkStoreOpts{Parallelism: 1})
}

func fromBulkGetResponse(item contribState.BulkGetResponse) *proto.BulkStateItem {
	return &proto.BulkStateItem{
		Key:  item.Key,
		Data: item.Data,
		Etag: internal.IfNotNil(item.ETag, func(etagValue *string) *proto.Etag {
			return &proto.Etag{
				Value: *etagValue,
			}
		}),
		Error:    item.Error,
		Metadata: item.Metadata,
		ContentType: internal.IfNotNil(item.ContentType, func(f *string) string {
			return *f
		}),
	}
}

func (s *store) BulkGet(ctx context.Context, req *proto.BulkGetRequest) (*proto.BulkGetResponse, error) {
	items, err := s.getInstance(ctx).BulkGet(ctx, internal.Map(req.Items, func(getReq *proto.GetRequest) contribState.GetRequest {
		return *toGetRequest(getReq)
	}), contribState.BulkGetOpts{Parallelism: 1})
	return &proto.BulkGetResponse{
		Items: internal.Map(items, fromBulkGetResponse),
	}, err
}

func (s *store) BulkSet(ctx context.Context, req *proto.BulkSetRequest) (*proto.BulkSetResponse, error) {
	return &proto.BulkSetResponse{}, s.getInstance(ctx).BulkSet(ctx, internal.Map(req.Items, func(setReq *proto.SetRequest) contribState.SetRequest {
		return *toSetRequest(setReq)
	}), contribState.BulkStoreOpts{Parallelism: 1})
}

func toTransactionalStateOperation(op *proto.TransactionalStateOperation) contribState.TransactionalStateOperation {
	if opDelete := op.GetDelete(); opDelete != nil {
		return *toDeleteRequest(opDelete)
	} else {
		return *toSetRequest(op.GetSet())
	}
}

func (s *store) Transact(ctx context.Context, req *proto.TransactionalStateRequest) (*proto.TransactionalStateResponse, error) {
	transactional, ok := s.getInstance(ctx).(contribState.TransactionalStore)

	if transactional == nil || !ok {
		return nil, status.Errorf(codes.Unimplemented, "method Transact not implemented")
	}

	return &proto.TransactionalStateResponse{}, transactional.Multi(ctx, &contribState.TransactionalStateRequest{
		Operations: internal.Map(req.Operations, toTransactionalStateOperation),
		Metadata:   req.Metadata,
	})
}

func (s *store) Query(ctx context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
	querier, ok := s.getInstance(ctx).(contribState.Querier)
	if querier == nil || !ok {
		return nil, status.Errorf(codes.Unimplemented, "method Query not implemented")
	}

	filters, err := internal.MapValuesErr(req.Query.Filter, func(f *anypb.Any) (any, error) {
		var v any
		return v, json.Unmarshal(f.Value, &v)
	})
	if err != nil {
		return nil, err
	}

	query := contribQuery.Query{
		QueryFields: contribQuery.QueryFields{
			Filters: filters,
			Sort: internal.Map(req.Query.Sort, func(s *proto.Sorting) contribQuery.Sorting {
				return contribQuery.Sorting{
					Key:   s.Key,
					Order: s.Order.String(),
				}
			}),
			Page: contribQuery.Pagination{
				Limit: int(req.Query.Pagination.Limit),
				Token: req.Query.Pagination.Token,
			},
		},
	}

	// marshal and unmarshal query is necessary since the filters are built when unmarshalling the query value.
	// TODO expose buildFilters function.
	btsQuery, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	// FIXME no idea why its necessary to unmarshal it into a map before trying to unmarshalling to the query object.
	var dict map[string]any

	if err = json.Unmarshal(btsQuery, &dict); err != nil {
		return nil, err
	}

	bts, err := json.Marshal(dict)
	if err != nil {
		return nil, err
	}

	var nq contribQuery.Query
	if err = json.Unmarshal(bts, &nq); err != nil {
		return nil, err
	}

	resp, err := querier.Query(ctx, &contribState.QueryRequest{
		Query:    nq,
		Metadata: req.Metadata,
	})
	if err != nil {
		return nil, err
	}

	return &proto.QueryResponse{
		Items: internal.Map(resp.Results, func(item contribState.QueryItem) *proto.QueryItem {
			return &proto.QueryItem{
				Key:  item.Key,
				Data: item.Data,
				Etag: internal.IfNotNil(item.ETag, func(etagValue *string) *proto.Etag {
					return &proto.Etag{
						Value: *etagValue,
					}
				}),
				Error: item.Error,
				ContentType: internal.IfNotNil(item.ContentType, func(f *string) string {
					return *f
				}),
			}
		}),
		Token:    resp.Token,
		Metadata: resp.Metadata,
	}, nil
}

func Register(server *grpc.Server, getInstance func(context.Context) Store) {
	store := &store{
		getInstance: getInstance,
	}
	proto.RegisterStateStoreServer(server, store)
	proto.RegisterTransactionalStateStoreServer(server, store)
	proto.RegisterQueriableStateStoreServer(server, store)
}
