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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceRun(t *testing.T) {
	t.Run("run should return an error when socket was not specified", func(t *testing.T) {
		assert.NotNil(t, Run())
	})

	t.Run("run should return an when no component was specified", func(t *testing.T) {
		t.Setenv(unixSocketPathEnvVar, "/tmp/fake.sock")
		assert.NotNil(t, Run())
	})
}
