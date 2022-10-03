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

package dapr

import (
	"context"
	"sync"

	"google.golang.org/grpc/metadata"
)

const metadataInstanceID = "x-component-instance"

// mux returns a function that creates and store new instances based on `x-component-instance` metadata header.
// when no component instance is provided so a default instance is used instead.
func mux[TComponent any](new func() TComponent) func(context.Context) TComponent {
	instances := sync.Map{}
	firstLoad := sync.Mutex{}
	return func(ctx context.Context) TComponent {
		instanceID := "#default__instance#"
		metadata, ok := metadata.FromIncomingContext(ctx)
		if ok {
			instanceIDs := metadata.Get(metadataInstanceID)
			if len(instanceIDs) != 0 {
				instanceID = instanceIDs[0]
			}
		}
		svcLogger.Infof("received request for instance: %s\n", instanceID)
		instance, ok := instances.Load(instanceID)
		if !ok {
			firstLoad.Lock()
			defer firstLoad.Unlock()
			instance, ok = instances.Load(instanceID) // double check lock
			if ok {
				return instance.(TComponent)
			}
			instance = new()
			instances.Store(instanceID, instance)
		}
		return instance.(TComponent)
	}
}
