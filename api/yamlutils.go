// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package yamlutils provides utility methods to convert Yaml files to SDK protos.
package yamlutils

import (
	"errors"
	"fmt"
	"time"

	"gopkg.in/yaml.v2"
)

var unmarshal = yaml.Unmarshal

// UnmarshalYAMLToMap unmarshalls Yaml file into a map[string]interface{} that can be decoded into JSON.
// The implementation has been copied over with slight modifications from a standard template.
// The function returns a JSON representation instead of the Proto as it's done in the referenced file.
func UnmarshalYAMLToMap(data []byte) (map[string]interface{}, error) {
	errCh := make(chan error)
	ch := make(chan map[string]interface{})
	go func() {
		// The yaml library can panic.
		// Add a recover() here to handle this gracefully.
		defer func() {
			if r := recover(); r != nil {
				errCh <- errors.New("panic caught: invalid yaml file")
			}
		}()
		var m map[string]interface{}
		if err := unmarshal(data, &m); err != nil {
			errCh <- err
			return
		}
		ch <- m
	}()

	var m map[string]interface{}
	select {
	case err := <-errCh:
		return nil, err
	case m = <-ch:
		break
	case <-time.After(10 * time.Second):
		return nil, errors.New("unmarshal took too long")
	}
	// fix is guaranteed to modify m to make it the right type.
	return fix(m).(map[string]interface{}), nil
}

// YAML unmarshalling produces a map[string]interface{} where the value might
// be a map[interface{}]interface{}, or a []interface{} where values might be a
// map[interface{}]interface{}, which json.Marshal does not support.
//
// So we have to go through the map and change any map[interface{}]interface{}
// we find into a map[string]interface{}, which JSON decoding supports.
//
// In order to make it compatible with our use case, the keys in the JSON object
// are converted from snake_case to camelCase.
func fix(in interface{}) interface{} {
	switch in.(type) {
	case map[interface{}]interface{}:
		// Create a new map[string]interface{} and fill it with fixed keys.
		cp := map[string]interface{}{}
		for k, v := range in.(map[interface{}]interface{}) {
			cp[fmt.Sprintf("%s", k)] = v
		}
		// Now fix the map[string]interface{} to fix the values.
		return fix(cp)
	case map[string]interface{}:
		// Fix each value in the map.
		sm := in.(map[string]interface{})
		cp := map[string]interface{}{}
		for k, v := range sm {
			cp[k] = fix(v)
		}
		return cp
	case []interface{}:
		// Fix each element in the slice.
		s := in.([]interface{})
		for i, v := range s {
			s[i] = fix(v)
		}
		return s
	default:
		// Value doesn't need to be fixed. If this is not a supported type, JSON
		// encoding will fail.
		return in
	}
}
