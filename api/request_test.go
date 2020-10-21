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

package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/protolambda/messagediff"
	"gopkg.in/yaml.v2"
)

func TestWritePreview(t *testing.T) {
	projectID := "project-123"
	sandbox := true
	want := map[string]interface{}{
		"parent": fmt.Sprintf("projects/%v", projectID),
		"previewSettings": map[string]interface{}{
			"sandbox": sandbox,
		},
	}
	got := WritePreview(projectID, sandbox)
	diff, equal := messagediff.DeepDiff(want, got)
	if !equal {
		t.Errorf("WritePreview returned an incorrect value; diff (want -> got)\n%s", diff)
	}
}

func TestWriteDraft(t *testing.T) {
	projectID := "project-123"
	want := map[string]interface{}{
		"parent": fmt.Sprintf("projects/%v", projectID),
	}
	got := WriteDraft(projectID)
	diff, equal := messagediff.DeepDiff(want, got)
	if !equal {
		t.Errorf("WritePreview returned an incorrect value; diff (want -> got)\n%s", diff)
	}
}

func TestCreateVersion(t *testing.T) {
	projectID := "project-123"
	releaseChannel := "prod"
	want := map[string]interface{}{
		"parent":          fmt.Sprintf("projects/%v", projectID),
		"release_channel": releaseChannel,
	}
	got := CreateVersion(projectID, releaseChannel)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("WriteVersion incorrectly populated the request: diff (-want, +got)\n%s", diff)
	}
}

func TestReadVersion(t *testing.T) {
	projectID := "project-123"
	versionID := "2"
	want := map[string]interface{}{
		"name": fmt.Sprintf("projects/%v/versions/%v", projectID, versionID),
	}
	got := ReadVersion(projectID, versionID)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ReadVersion returned an incorrect value: diff (-want, +got)\n%s", diff)
	}
}

func TestListReleaseChannels(t *testing.T) {
	projectID := "project-123"
	want := map[string]interface{}{
		"parent": fmt.Sprintf("projects/%v", projectID),
	}
	got := ListReleaseChannels(projectID)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ListReleaseChannels returned an incorrect value: diff (-want, +got)\n%s", diff)
	}
}

func TestAddConfigFiles(t *testing.T) {
	tests := []struct {
		files map[string][]byte
		want  map[string]interface{}
		err   error
	}{
		{
			files: map[string][]byte{
				"verticals/CharacterAlarms.yaml":           []byte("foo: bar"),
				"actions/actions.yaml":                     []byte("intent_name: alarm"),
				"manifest.yaml":                            []byte("version: 1.0"),
				"settings/settings.yaml":                   []byte("display_name: alarm"),
				"settings/zh-TW/settings.yaml":             []byte("developer_email: foo@foo.com"),
				"custom/global/actions.intent.CANCEL.yaml": []byte("transitionToScene: actions.scene.END_CONVERSATION"),
				"custom/intents/help.yaml":                 []byte("phrase: hello"),
				"custom/intents/ru/help.yaml":              []byte("phrase: hello"),
				"custom/prompts/foo.yaml":                  []byte("prompt: \"yes\""),
				"custom/prompts/ru/foo.yaml":               []byte("prompt: \"yes\""),
				"custom/scenes/a.yaml":                     []byte("name: a"),
				"custom/types/b.yaml":                      []byte("type: b"),
				"custom/types/ru/b.yaml":                   []byte("type: b"),
				"webhooks/webhook1.yaml": []byte(
					`
inlineCloudFunction:
  execute_function: hello
`),
				"webhooks/webhook2.yaml": []byte(
					`
external_endpoint:
  base_url: https://google.com
  http_headers:
    content-type: application/json
  endpoint_api_version: 1
`),
				"resources/strings/bundle.yaml": []byte(
					`
x: "777"
y: "777"
greeting: "hello world"
`),
			},
			want: map[string]interface{}{
				"configFiles": map[string][]interface{}{
					"configFiles": {
						map[string]interface{}{
							"filePath":         "verticals/CharacterAlarms.yaml",
							"verticalSettings": map[string]interface{}{"foo": "bar"},
						},
						map[string]interface{}{
							"filePath": "actions/actions.yaml",
							"actions":  map[string]interface{}{"intent_name": "alarm"},
						},
						map[string]interface{}{
							"filePath": "manifest.yaml",
							"manifest": map[string]interface{}{"version": 1.0},
						},
						map[string]interface{}{
							"filePath": "settings/settings.yaml",
							"settings": map[string]interface{}{"display_name": "alarm"},
						},
						map[string]interface{}{
							"filePath": "settings/zh-TW/settings.yaml",
							"settings": map[string]interface{}{"developer_email": "foo@foo.com"},
						},
						map[string]interface{}{
							"filePath":          "custom/global/actions.intent.CANCEL.yaml",
							"globalIntentEvent": map[string]interface{}{"transitionToScene": "actions.scene.END_CONVERSATION"},
						},
						map[string]interface{}{
							"filePath": "custom/intents/help.yaml",
							"intent":   map[string]interface{}{"phrase": "hello"},
						},
						map[string]interface{}{
							"filePath": "custom/intents/ru/help.yaml",
							"intent":   map[string]interface{}{"phrase": "hello"},
						},
						map[string]interface{}{
							"filePath":     "custom/prompts/foo.yaml",
							"staticPrompt": map[string]interface{}{"prompt": "yes"},
						},
						map[string]interface{}{
							"filePath":     "custom/prompts/ru/foo.yaml",
							"staticPrompt": map[string]interface{}{"prompt": "yes"},
						},
						map[string]interface{}{
							"filePath": "custom/scenes/a.yaml",
							"scene":    map[string]interface{}{"name": "a"},
						},
						map[string]interface{}{
							"filePath": "custom/types/b.yaml",
							"type":     map[string]interface{}{"type": "b"},
						},
						map[string]interface{}{
							"filePath": "custom/types/ru/b.yaml",
							"type":     map[string]interface{}{"type": "b"},
						},
						map[string]interface{}{
							"filePath": "webhooks/webhook1.yaml",
							"webhook": map[string]interface{}{
								"inlineCloudFunction": map[string]interface{}{
									"execute_function": "hello",
								},
							},
						},
						map[string]interface{}{
							"filePath": "webhooks/webhook2.yaml",
							"webhook": map[string]interface{}{
								"external_endpoint": map[string]interface{}{
									"base_url": "https://google.com",
									"http_headers": map[string]interface{}{
										"content-type": "application/json",
									},
									"endpoint_api_version": 1,
								},
							},
						},
						map[string]interface{}{
							"filePath": "resources/strings/bundle.yaml",
							"resourceBundle": map[string]interface{}{
								"x":        "777",
								"y":        "777",
								"greeting": "hello world",
							},
						},
					},
				},
			},
			err: nil,
		},
		{
			files: map[string][]byte{},
			want: map[string]interface{}{
				"configFiles": map[string][]interface{}{},
			},
			err: nil,
		},
		{
			files: map[string][]byte{
				"manifest.yaml": []byte("version: 1.0"),
				"extrafile":     []byte("key: should raise an error"),
			},
			want: map[string]interface{}{},
			err:  errors.New("failed to add extrafile to a request"),
		},
	}
	for _, tc := range tests {
		req := map[string]interface{}{}
		err := addConfigFiles(req, tc.files, ".")
		if err != nil {
			if tc.err == nil {
				t.Errorf("AddConfigFiles returned %v, want %v, input %v", err, tc.err, tc.files)
			}
		}
		if tc.err == nil {
			wantCfgs, ok := tc.want["configFiles"].(map[string][]interface{})
			if !ok {
				t.Errorf("Failed to convert to type: tc.want[\"configFiles\"] is incorrect type")
			}
			fs, ok := req["files"].(map[string]interface{})
			if !ok {
				t.Errorf("Failed type conversion: expected files inside of the request to be of type map[string]interface{}")
			}
			reqCfgs, ok := fs["configFiles"].(map[string][]interface{})
			if !ok {
				t.Errorf("Failed type conversion: expected configFiles inside of the request to be of type map[string][]interface{}")
			}
			if diff := cmp.Diff(wantCfgs["configFiles"], reqCfgs["configFiles"], cmpopts.SortSlices(func(l, r interface{}) bool {
				lmp, ok := l.(map[string]interface{})
				if !ok {
					t.Errorf("can not convert %v to map[string]interface{}", l)
				}
				rmp, ok := r.(map[string]interface{})
				if !ok {
					t.Errorf("can not convert %v to map[string]interface{}", r)
				}
				return lmp["filePath"].(string) < rmp["filePath"].(string)
			})); diff != "" {
				t.Errorf("AddConfigFiles didn't add the config files to a request correctly: diff (-want, +got)\n%s", diff)
			}
		}
	}
}

func TestAddDataFiles(t *testing.T) {
	tests := []struct {
		files map[string][]byte
		want  map[string]interface{}
		err   error
	}{
		{
			files: map[string][]byte{
				"audio1.mp3":     []byte("abc123"),
				"image1.jpg":     []byte("abc123"),
				"audio2.wav":     []byte("abc123"),
				"animation1.flr": []byte("xyz789"),
			},
			want: map[string]interface{}{
				"files": map[string]interface{}{
					"dataFiles": map[string][]interface{}{
						"dataFiles": {
							map[string]interface{}{
								"filePath":    "audio1.mp3",
								"payload":     []byte("abc123"),
								"contentType": mime.TypeByExtension(filepath.Ext("audio1.mp3")),
							},
							map[string]interface{}{
								"filePath":    "image1.jpg",
								"payload":     []byte("abc123"),
								"contentType": mime.TypeByExtension(filepath.Ext("image1.jpg")),
							},
							map[string]interface{}{
								"filePath":    "audio2.wav",
								"payload":     []byte("abc123"),
								"contentType": mime.TypeByExtension(filepath.Ext("audio2.wav")),
							},
							map[string]interface{}{
								"filePath":    "animation1.flr",
								"payload":     []byte("xyz789"),
								"contentType": "x-world/x-vrml",
							},
						},
					},
				},
			},
			err: nil,
		},
		{
			files: map[string][]byte{
				"audio1.xyz": []byte("abc123"),
			},
			want: map[string]interface{}{},
			err:  nil,
		},
		{
			files: map[string][]byte{
				"webhooks/webhook1.zip": []byte("===abc==="),
			},
			want: map[string]interface{}{
				"files": map[string]interface{}{
					"dataFiles": map[string][]interface{}{
						"dataFiles": {
							map[string]interface{}{
								"filePath":    "webhooks/webhook1.zip",
								"payload":     []byte("===abc==="),
								"contentType": "application/zip;zip_type=cloud_function",
							},
						},
					},
				},
			},
			err: nil,
		},
	}
	for _, tc := range tests {
		req := map[string]interface{}{}
		if err := addDataFiles(req, tc.files, "."); err != nil {
			if tc.err == nil {
				t.Errorf("addDataFiles returned %v, want %v, input (files: %v)", err, tc.err, tc.files)
			}
		}
		type dataFile struct {
			Filepath    string `json:"filePath"`
			Payload     []byte `json:"payload"`
			ContentType string `json:"contentType"`
		}
		type reqFmt struct {
			Files struct {
				DataFiles struct {
					DataFiles []dataFile `json:"dataFiles"`
				} `json:"dataFiles"`
			} `json:"files"`
		}
		b, err := json.Marshal(tc.want)
		if err != nil {
			t.Errorf("Failed to marshal %v into JSON: %v", tc.want, err)
		}
		r := &reqFmt{}
		if err := json.Unmarshal(b, r); err != nil {
			t.Errorf("Failed to unmarshal into a struct: %v", err)
		}

		b, err = json.Marshal(req)
		if err != nil {
			t.Errorf("Failed to marshal %v into JSON: %v", req, err)
		}
		r2 := &reqFmt{}
		if err := json.Unmarshal(b, r2); err != nil {
			t.Errorf("Failed to unmarshal into a struct: %v", err)
		}
		if diff := cmp.Diff(r.Files.DataFiles.DataFiles, r2.Files.DataFiles.DataFiles, cmpopts.SortSlices(func(l, r dataFile) bool {
			return l.Filepath < r.Filepath
		})); diff != "" {
			t.Errorf("addDataFiles incorrectly populated the request: diff (-want, +got)\n%s", diff)
		}
	}
}

func TestNewStreamer(t *testing.T) {
	cfgs := map[string][]byte{
		"actions/actions.yaml":             []byte("42"),
		"settings/en/settings.yaml":        []byte("displayName: foo"),
		"custom/intents/intent1.yaml":      []byte("name: intent123"),
		"settings/settings.yaml":           []byte("projectID: 123"),
		"manifest.yaml":                    []byte("version: 1.0"),
		"resources/strings/bundle.yaml":    []byte("a: foo"),
		"resources/strings/en/bundle.yaml": []byte("a: foo b: bar"),
	}
	dfs := map[string][]byte{
		"resources/images/image1.png": []byte("abc"),
		"resources/images/image3.png": []byte("abcdefghi"),
		"resources/images/image2.png": []byte("abcdef"),
	}
	makeRequest := func() map[string]interface{} {
		return nil
	}
	root := "."
	chunkSize := 1024
	s := NewStreamer(cfgs, dfs, makeRequest, root, chunkSize)

	// This is in correct sorted order
	wantCfgnames := []string{"settings/settings.yaml", "manifest.yaml", "settings/en/settings.yaml",
		"actions/actions.yaml", "resources/strings/bundle.yaml", "resources/strings/en/bundle.yaml", "custom/intents/intent1.yaml"}
	// Check that first three elements are settings, manifest files and the rest are sorted according to their size.
	if diff := cmp.Diff(wantCfgnames[:3], s.configFilenames[:3], cmpopts.SortSlices(strLess)); diff != "" {
		t.Errorf("NewStreamer didn't have settings and manifest in the beginning of configFilenames: diff (-want, +got)\n%s", diff)
	}
	if diff := cmp.Diff(wantCfgnames[3:], s.configFilenames[3:]); diff != "" {
		t.Errorf("NewStreamer didn't have rest of config files sorted correctly: diff (-want, +got)\n%s", diff)
	}

	wantDfnames := []string{"resources/images/image1.png", "resources/images/image2.png", "resources/images/image3.png"}
	if diff := cmp.Diff(wantDfnames, s.dataFilenames); diff != "" {
		t.Errorf("NewStreamer didn't have rest of config files sorted correctly: diff (-want, +got)\n%s", diff)
	}
}

func TestMoveToFront(t *testing.T) {
	tests := []struct {
		a    []string
		ps   []int
		want []string
	}{
		{
			a:    []string{"settings/settings.yaml", "settings/en/settings.yaml", "manifest.yaml"},
			ps:   []int{0, 1, 2},
			want: []string{"settings/settings.yaml", "settings/en/settings.yaml", "manifest.yaml"},
		},
		{
			a:    []string{"settings/settings.yaml", "custom/intents/intent.yaml", "settings/en/settings.yaml", "manifest.yaml", "actions/actions.yaml"},
			ps:   []int{0, 2, 3},
			want: []string{"settings/settings.yaml", "settings/en/settings.yaml", "manifest.yaml"},
		},
	}
	for _, tc := range tests {
		moveToFront(tc.a, tc.ps)
		if diff := cmp.Diff(tc.a[:len(tc.ps)], tc.want, cmpopts.SortSlices(strLess)); diff != "" {
			t.Errorf("moveToFront didn't produce correct result: diff (-want, +got)\n%s", diff)
		}
	}
}

var strLess = func(s1, s2 string) bool { return s1 < s2 }

func parseReq(t *testing.T, req map[string]interface{}) []string {
	t.Helper()
	type configFileReq struct {
		Files struct {
			ConfigFiles struct {
				ConfigFiles []struct {
					FilePath string `json:"filePath"`
				} `json:"configFiles"`
			} `json:"configFiles"`
		} `json:"files"`
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Errorf("Failed to marshal request into JSON: %v", err)
	}
	r := configFileReq{}
	if err = json.Unmarshal(b, &r); err != nil {
		t.Errorf("Failed to unmarshal JSON into a map: %v", err)
	}
	res := []string{}
	for _, v := range r.Files.ConfigFiles.ConfigFiles {
		res = append(res, v.FilePath)
	}
	return res
}

func TestNextWithTwoFiles(t *testing.T) {
	cfgs := map[string][]byte{
		"settings/settings.yaml": []byte(`projectId: hello-world`),
		"manifest.yaml":          []byte(`version: 1.0`),
	}
	// Add a file that is equal to the "sum" of the rest of the confg files,
	// so that it will be easier to split files.
	yml := map[string]interface{}{
		"version":   "1.0",
		"projectId": "hello-world",
	}
	out, err := yaml.Marshal(yml)
	if err != nil {
		t.Fatalf("Failed to marshall %v into YAML: %v", yml, err)
	}
	cfgs["custom/intents/intent1.yaml"] = out
	dfs := map[string][]byte{}
	mkreq := func() map[string]interface{} {
		return map[string]interface{}{}
	}
	// Sets chunkSize to the sum of the first two request. Thus,
	// streamer is guaranteed to return two requests.
	s := NewStreamer(cfgs, dfs, mkreq, ".", len(out))
	req1, err := s.Next()
	if err != nil {
		t.Errorf("SDKStreamer.Next failed to return the 1st request: %v", err)
	}
	want1 := []string{"settings/settings.yaml", "manifest.yaml"}
	got1 := parseReq(t, req1)
	if diff := cmp.Diff(want1, got1, cmpopts.SortSlices(strLess)); diff != "" {
		t.Errorf("SDKStreamer.Next returned an incorrect request: diff (-want, +got)\n%s", diff)
	}
	if hasNext := s.HasNext(); !hasNext {
		t.Errorf("HasNext returned %v, but want %v", hasNext, true)
	}
	req2, err := s.Next()
	if err != nil {
		t.Errorf("SDKStreamer.Next failed to return the 1st request: %v", err)
	}
	want2 := []string{"custom/intents/intent1.yaml"}
	got2 := parseReq(t, req2)
	if diff := cmp.Diff(want2, got2, cmpopts.SortSlices(strLess)); diff != "" {
		t.Errorf("SDKStreamer.Next returned an incorrect request: diff (-want, +got)\n%s", diff)
	}
	if hasNext := s.HasNext(); hasNext {
		t.Errorf("HasNext returned %v, but want %v", hasNext, false)
	}
}

func TestNextWhenChunkSizeTooSmall(t *testing.T) {
	cfgs := map[string][]byte{
		"settings/settings.yaml": []byte(`projectId: hello-world`),
		"manifest.yaml":          []byte(`version: 1.0`),
	}
	yml := map[string]interface{}{
		"version":   "1.0",
		"projectId": "hello-world",
	}
	out, err := yaml.Marshal(yml)
	if err != nil {
		t.Fatalf("Failed to marshall %v into YAML: %v", yml, err)
	}
	cfgs["custom/intents/intent1.yaml"] = out
	dfs := map[string][]byte{}
	mkreq := func() map[string]interface{} {
		return map[string]interface{}{}
	}
	s := NewStreamer(cfgs, dfs, mkreq, ".", 1)
	req1, err := s.Next()
	if err == nil {
		t.Errorf("SDKStreamer.Next returned %v, but needs an error: %v", req1, err)
	}
}
