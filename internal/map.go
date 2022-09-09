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

// Map apply the given function to all list elements.
func Map[From any, To any](from []From, mapper func(From) To) []To {
	res := make([]To, len(from))

	for idx, value := range from {
		res[idx] = mapper(value)
	}
	return res
}

// MapValues transform map values by applying mapper func.
func MapValuesErr[From any, To any](from map[string]From, mapper func(From) (To, error)) (map[string]To, error) {
	res := make(map[string]To, len(from))

	for key, value := range from {
		mapped, err := mapper(value)
		if err != nil {
			return nil, err
		}
		res[key] = mapped
	}
	return res, nil
}

// IfNotNilP apply the mapper func if the value is not nil returns nil otherwise.
func IfNotNilP[From any, To any](value *From, mapper func(*From) To) *To {
	if value != nil {
		value := mapper(value)
		return &value
	}
	return nil
}

// IfNotNil apply the mapper func if the value is not nil returns zero value otherwise.
func IfNotNil[From any, To any](value *From, mapper func(*From) To) To {
	v := IfNotNilP(value, mapper)
	if v == nil {
		var zero To
		return zero
	}
	return *v
}
