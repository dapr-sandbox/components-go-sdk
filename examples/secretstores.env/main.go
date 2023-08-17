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

package main

import (
	dapr "github.com/dapr-sandbox/components-go-sdk"
	"github.com/dapr-sandbox/components-go-sdk/secretstores/v1"
	im "github.com/dapr/components-contrib/secretstores/local/env"
	"github.com/dapr/kit/logger"
)

var log = logger.NewLogger("env-secret-store-pluggable")

func main() {
	dapr.Register("local.env-pluggable", dapr.WithSecretStore(func() secretstores.SecretStore {
		return im.NewEnvSecretStore(log)
	}))
	dapr.MustRun()
}
