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

package internal

import (
	"sync"
)

// GRPCServerStream is the interface for server streaming.
type GRPCServerStream[TSend any, TRecv any] interface {
	Send(*TSend) error
	Recv() (*TRecv, error)
}

// ThreadSafeStream wraps a grpc stream with locks to permit send and recv in multiples goroutines.
type ThreadSafeStream[TSend any, TRecv any] interface {
	GRPCServerStream[TSend, TRecv]
}

// GRPCThreadSafeStream wraps a grpcStream for thread safe operations.
// As per documentation is unsafe to call recv OR send in multiples goroutines
// https://github.com/grpc/grpc-go/blob/master/Documentation/concurrency.md#streams
// it is safe to call recv AND send in different goroutines.
type GRPCThreadSafeStream[TSend any, TRecv any, TStream GRPCServerStream[TSend, TRecv]] struct {
	recvLock *sync.Mutex
	sendLock *sync.Mutex
	stream   TStream
}

func NewGRPCThreadSafeStream[TSend any, TRecv any](stream GRPCServerStream[TSend, TRecv]) ThreadSafeStream[TSend, TRecv] {
	return &GRPCThreadSafeStream[TSend, TRecv, GRPCServerStream[TSend, TRecv]]{
		recvLock: &sync.Mutex{},
		sendLock: &sync.Mutex{},
		stream:   stream,
	}
}
func (s *GRPCThreadSafeStream[TSend, TRecv, GRPCServerStream]) Send(msg *TSend) error {
	s.sendLock.Lock()
	defer s.sendLock.Unlock()
	return s.stream.Send(msg)
}

func (s *GRPCThreadSafeStream[TSend, TRecv, GRPCServerStream]) Recv() (*TRecv, error) {
	s.recvLock.Lock()
	defer s.recvLock.Unlock()
	return s.stream.Recv()
}
