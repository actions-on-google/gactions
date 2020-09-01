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

package yamlutils

import (
	"testing"

	"github.com/protolambda/messagediff"
	"gopkg.in/yaml.v2"
)

func TestUnmarshalYamlToJSONRecover(t *testing.T) {
	originalUnmarshal := unmarshal
	unmarshal = func([]byte, interface{}) error {
		panic("This is a panic")
	}
	defer func() { unmarshal = originalUnmarshal }()
	got, err := UnmarshalYAMLToMap([]byte("ignored_yaml"))
	if got != nil {
		t.Errorf("got %v, want err", got)
	}
	if "panic caught: invalid yaml file" != err.Error() {
		t.Errorf("got %v, want \"invalid yaml file\"", err)
	}
}

func TestUnmarshalYamlToMap(t *testing.T) {
	in, err := yaml.Marshal(map[string]interface{}{
		"snake_case": 3,
		"foo_bar": map[string]interface{}{
			"the_answer": 42,
			"foo":        "remains the same",
		},
	})
	if err != nil {
		t.Fatalf("Can not parse map into yaml: %v", err)
	}
	got, err := UnmarshalYAMLToMap(in)
	want := map[string]interface{}{
		"snake_case": 3,
		"foo_bar": map[string]interface{}{
			"the_answer": 42,
			"foo":        "remains the same",
		},
	}
	if err != nil {
		t.Errorf("UnmarshalYAMLToMap produced an error: got %v, want %v", err, nil)
	}
	diff, equal := messagediff.DeepDiff(want, got)
	if !equal {
		t.Errorf("UnmarshalYAMLToMap returned an incorrect value; diff (want -> got)\n%s", diff)
	}
}

func TestDOSYAML(t *testing.T) {
	dos := `
a: &a ["lol","lol","lol","lol","lol","lol","lol","lol","lol"]
b: &b [*a,*a,*a,*a,*a,*a,*a,*a,*a]
c: &c [*b,*b,*b,*b,*b,*b,*b,*b,*b]
d: &d [*c,*c,*c,*c,*c,*c,*c,*c,*c]
e: &e [*d,*d,*d,*d,*d,*d,*d,*d,*d]
f: &f [*e,*e,*e,*e,*e,*e,*e,*e,*e]
g: &g [*f,*f,*f,*f,*f,*f,*f,*f,*f]
h: &h [*g,*g,*g,*g,*g,*g,*g,*g,*g]
i: &i [*h,*h,*h,*h,*h,*h,*h,*h,*h]
`
	if b, err := UnmarshalYAMLToMap([]byte(dos)); err == nil {
		t.Errorf("DOS YAML successfully parsed into build %v.", b)
	}
}
