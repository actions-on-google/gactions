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

package sdk

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/actions-on-google/gactions/api/request"
	"github.com/actions-on-google/gactions/api/testutils"
	"github.com/actions-on-google/gactions/api/yamlutils"
	"github.com/actions-on-google/gactions/project"
	"github.com/actions-on-google/gactions/project/studio"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func buildPathToProjectFiles() string {
	return filepath.Join("api", "examples", "account_linking_gsi")
}

type MockStudio struct {
	files        map[string][]byte
	clientSecret []byte
	root         string
	projectID    string
}

func NewMock(files map[string][]byte) MockStudio {
	m := MockStudio{}
	m.files = files
	m.projectID = "placeholder_project"
	return m
}

func (m MockStudio) ProjectID() string {
	return m.projectID
}

func (MockStudio) Download(sample project.SampleProject, dest string) error {
	return nil
}

func (MockStudio) AlreadySetup(pathToWorkDir string) bool {
	return false
}

func (p MockStudio) Files() (map[string][]byte, error) {
	return p.files, nil
}

func (MockStudio) ClientSecretJSON() ([]byte, error) {
	return []byte{}, nil
}

func (p MockStudio) ProjectRoot() string {
	return p.root
}

type myReader struct {
	r   io.Reader
	lat time.Duration
}

func (mr myReader) Read(p []byte) (n int, err error) {
	time.Sleep(mr.lat)
	return mr.r.Read(p)
}

func TestReadBodyWithTimeout(t *testing.T) {
	var got, want []byte
	var err error
	var r myReader

	r = myReader{r: strings.NewReader("hello"), lat: time.Duration(200) * time.Millisecond}
	// Timeout for 5 seconds to reduce flakiness.
	got, err = readBodyWithTimeout(r, time.Duration(5)*time.Second)
	want = []byte("hello")
	if err != nil {
		t.Errorf("readBodyWithTimeout returned %v, want %v", err, nil)
	}
	if string(got) != string(want) {
		t.Errorf("readBodyWithTimeout got %v, want %v", string(got), string(want))
	}

	// slow case
	r = myReader{r: strings.NewReader("hello"), lat: time.Duration(3) * time.Second}
	got, err = readBodyWithTimeout(r, time.Duration(1)*time.Second)
	want = []byte("")
	if err != nil {
		t.Errorf("readBodyWithTimeout returned %v, want %v", err, nil)
	}
	if string(got) != string(want) {
		t.Errorf("readBodyWithTimeout got %v, want %v", string(got), string(want))
	}
}

func TestPostprocessJSONResponse(t *testing.T) {
	tests := []struct {
		in        *http.Response
		shouldErr bool
	}{
		{
			in: &http.Response{
				StatusCode: 200,
				Body: ioutil.NopCloser(bytes.NewReader([]byte(
					`{
						  "validationResults": {
							  "results":[
								  {
									  "validationMesssage": "Your app doesn't have the correct size for the logo."
								  }
								]
							}
						}`,
				))),
			},
			shouldErr: false,
		},
		{
			in: &http.Response{
				StatusCode: 500,
				Body: ioutil.NopCloser(bytes.NewReader([]byte(
					`{
						  "error": {
						  "code": 500,
						  "message": "Internal error encountered",
						  "status": "INTERNAL",
						  "details": [
							 	{
									"@type": "type.googleapis.com/google.rpc.DebugInfo",
									"detail": "Should not be shown to user."
								}
							 ]
						 }
					 }`,
				))),
			},
			shouldErr: true,
		},
		{
			in: &http.Response{
				StatusCode: 400,
				Body: ioutil.NopCloser(bytes.NewReader([]byte(
					`{}`,
				))),
			},
			shouldErr: true,
		},
	}
	for _, tc := range tests {
		errCh := make(chan error)
		go postprocessJSONResponse(tc.in, errCh, func(body []byte) error {
			// TODO: Ideally would like to check that this function gets called.
			// Need a way to cleanly implement it.
			return nil
		})
		got := <-errCh
		if tc.shouldErr && got == nil {
			t.Errorf("postprocessJSONResponse returned incorrect result: got %v, want an error", got)
		}
	}
}

func unmarshal(t *testing.T, p string) map[string]interface{} {
	t.Helper()
	b := testutils.ReadFileOrDie(p)
	m, err := yamlutils.UnmarshalYAMLToMap(b)
	if err != nil {
		t.Fatalf("unmarshal: can not parse settins yaml into proto: %v", err)
	}
	return m
}

func TestSendFilesToServerJSON(t *testing.T) {
	if runtime.GOOS == "windows" {
		// This test does not work on Windows, as the "actions/actions.yaml"
		// and other files cannot be found.
		// The error specifically is:
		// Cannot open file C:\...\_bazel_kbuilder\jzegmkbf\execroot\__main__\bazel-out\x64_windows-fastbuild\bin\api\sdk_test_\sdk_test.exe.runfiles\__main__92api\examples\account_linking_gsi\actions\actions.yaml
		// Exit early.
		return
	}
	tests := []struct {
		projFiles                 map[string][]byte
		wantRequests              []map[string]interface{}
		wantErrorMessageToContain string
	}{
		{
			projFiles: map[string][]byte{
				"actions/actions.yaml":               testutils.ReadFileOrDie(filepath.Join(buildPathToProjectFiles(), "actions", "actions.yaml")),
				"manifest.yaml":                      testutils.ReadFileOrDie(filepath.Join(buildPathToProjectFiles(), "manifest.yaml")),
				"settings/settings.yaml":             testutils.ReadFileOrDie(filepath.Join(buildPathToProjectFiles(), "settings", "settings.yaml")),
				"resources/audio/confirmation_01.mp3": testutils.ReadFileOrDie(filepath.Join(buildPathToProjectFiles(), "resources", "audio", "confirmation_01.mp3")),
				"resources/images/smallLogo.jpg":     testutils.ReadFileOrDie(filepath.Join(buildPathToProjectFiles(), "resources", "images", "smallLogo.jpg")),
				"settings/zh-TW/settings.yaml":       testutils.ReadFileOrDie(filepath.Join(buildPathToProjectFiles(), "settings", "zh-TW", "settings.yaml")),
				"resources/images/zh-TW/smallLogo.jpg": testutils.ReadFileOrDie(filepath.Join(buildPathToProjectFiles(), "resources", "images", "zh-TW", "smallLogo.jpg")),
				"webhooks/webhook.yaml":              testutils.ReadFileOrDie(filepath.Join(buildPathToProjectFiles(), "webhooks", "webhook.yaml")),
				"settings/accountLinkingSecret.yaml": []byte(strings.Join([]string{"encryptedClientSecret: bar", "encryptionKeyVersion: 1"}, "\n")),
			},
			wantRequests: []map[string]interface{}{
				map[string]interface{}{
					"parent": "projects/placeholder_project",
					"files": map[string]interface{}{
						"configFiles": map[string]interface{}{
							"configFiles": []map[string]interface{}{
								map[string]interface{}{
									"filePath": "actions/actions.yaml",
									"actions": unmarshal(t, filepath.Join(buildPathToProjectFiles(), "actions", "actions.yaml")),
								},
								map[string]interface{}{
									"filePath": "manifest.yaml",
									"manifest": unmarshal(t, path.Join(buildPathToProjectFiles(), "manifest.yaml")),
								},
								map[string]interface{}{
									"filePath": "settings/settings.yaml",
									"settings": unmarshal(t, path.Join(buildPathToProjectFiles(), "settings", "settings.yaml")),
								},
								map[string]interface{}{
									"filePath": "settings/zh-TW/settings.yaml",
									"settings": unmarshal(t, path.Join(buildPathToProjectFiles(), "settings", "zh-TW", "settings.yaml")),
								},
								map[string]interface{}{
									"filePath": "webhooks/webhook.yaml",
									"webhook":  unmarshal(t, path.Join(buildPathToProjectFiles(), "webhooks", "webhook.yaml")),
								},
								map[string]interface{}{
									"filePath": "settings/accountLinkingSecret.yaml",
									"accountLinkingSecret": map[string]interface{}{
										"encryptedClientSecret": "bar",
										"encryptionKeyVersion":  1,
									},
								},
							},
						},
					},
				},
				map[string]interface{}{
					"parent": "projects/placeholder_project",
					"files": map[string]interface{}{
						"dataFiles": map[string]interface{}{
							"dataFiles": []map[string]interface{}{
								map[string]interface{}{
									"filePath":    "resources/images/smallLogo.jpg",
									"contentType": "image/jpeg",
									"payload":     testutils.ReadFileOrDie(path.Join(buildPathToProjectFiles(), "resources", "images", "smallLogo.jpg")),
								},
								map[string]interface{}{
									"filePath":    "resources/audio/confirmation_01.mp3",
									"contentType": "audio/mpeg",
									"payload":     testutils.ReadFileOrDie(path.Join(buildPathToProjectFiles(), "resources", "audio", "confirmation_01.mp3")),
								},
								map[string]interface{}{
									"filePath":    "resources/images/zh-TW/smallLogo.jpg",
									"contentType": "image/jpeg",
									"payload":     testutils.ReadFileOrDie(path.Join(buildPathToProjectFiles(), "resources", "images", "zh-TW", "smallLogo.jpg")),
								},
							},
						},
					},
				},
			},
			wantErrorMessageToContain: "",
		},
		{
			projFiles:                 map[string][]byte{},
			wantRequests:              nil,
			wantErrorMessageToContain: "configuration files for your Action were not found",
		},
	}
	for _, tc := range tests {
		p := NewMock(tc.projFiles)
		r, w := io.Pipe()
		ch := make(chan []byte)
		errCh := make(chan error)
		go func() {
			b, err := ioutil.ReadAll(r)
			ch <- b
			errCh <- err
		}()
		err := sendFilesToServerJSON(p, w, func() map[string]interface{} {
			// TODO: Parametrize this to enable testing of various requests.
			// This will remove need for request tests in request_test.
			return request.WriteDraft("placeholder_project")
		})
		gotBytes := <-ch
		if err := <-errCh; err != nil {
			t.Errorf("Unable to read from pipe: got %v, input %v", err, tc.projFiles)
		}
		if tc.wantRequests != nil {
			wantBytes, err := json.Marshal(tc.wantRequests)
			if err != nil {
				t.Errorf("Could not marshall into JSON: got %v", err)
			}
			var got []map[string]interface{}
			if err := json.Unmarshal(gotBytes, &got); err != nil {
				t.Errorf("Could not unmarshall to JSON: got %v", err)
			}
			// Checks request were sent in alphabetical order of filenames.
			var fps []string
			for _, v := range got {
				if fp, ok := v["filePath"]; ok {
					fps = append(fps, fp.(string))
				}
			}
			if ok := sort.StringsAreSorted(fps); !ok {
				t.Errorf("Expected requests to be in alphabetical order, but got %v\n", fps)
			}
			var want []map[string]interface{}
			if err := json.Unmarshal(wantBytes, &want); err != nil {
				t.Errorf("Could not unmarshall to JSON: got %v", err)
			}
			if diff := cmp.Diff(want, got, cmpopts.SortSlices(func(l, r interface{}) bool {
				lb, err := json.Marshal(l)
				if err != nil {
					t.Errorf("can not marshal %v to JSON: %v", lb, err)
				}
				rb, err := json.Marshal(r)
				if err != nil {
					t.Errorf("can not marshal %v to JSON: %v", rb, err)
				}
				return string(lb) < string(rb)
			})); diff != "" {
				t.Errorf("sendFilesToServerJSON didn't send correct files: diff (-want, +got)\n%s", diff)
			}
		} else {
			if !strings.Contains(err.Error(), tc.wantErrorMessageToContain) {
				t.Errorf("sendFilesToServerJSON got %v, but want the error to have %v\n", err, tc.wantErrorMessageToContain)
			}
		}
	}
}

func TestProcWritePreviewResponse(t *testing.T) {
	tests := []struct {
		in      []byte
		wantURL string
	}{
		{
			in: []byte(
				`
{
 "simulatorUrl": "https://google.com"
}`,
			),
			wantURL: "https://google.com",
		},
		{
			in: []byte(
				`
{
	"simulatorUrl": "https://google.com",
	"validationResults": {
		"results": [
			{
				"validationMessage": "Your app must have a 32x32 logo"
			}
		]
	}
}`,
			),
			wantURL: "https://google.com",
		},
		{
			in:      []byte("{}"),
			wantURL: "",
		},
		{
			in: []byte(
				`
{
	"simulatorUrl": "https://google.com",
	"validationResults": {
		"results": [
			{}
		]
	}
}`,
			),
			wantURL: "https://google.com",
		},
	}
	for _, tc := range tests {
		gotURL, err := procWritePreviewResponse(tc.in)
		if err != nil {
			t.Errorf("procWritePreviewResponse returned %v, but want %v, input %v", err, nil, tc.in)
		}
		if tc.wantURL != gotURL {
			t.Errorf("procWritePreviewResponse didn't set the right value of the simulator URL: got %v, want %v, input %v", gotURL, tc.wantURL, tc.in)
		}
	}
}

func TestProcWriteDraftResponse(t *testing.T) {
	tests := []struct {
		body string
	}{
		{
			body: `
{
	"name": "foo/bar",
	"validationResults": {
		"results": [
			{
				"validationMessage": "Your app must have a 32x32 logo"
			}
		]
	}
}
`,
		},
		{
			body: `
{
	"name": "foo/bar",
	"validationResults": {
		"results": [
			{}
		]
	}
}
`,
		},
	}
	for _, tc := range tests {
		if err := procWriteDraftResponse([]byte(tc.body)); err != nil {
			t.Errorf("procWriteDraftResponse returned %v, but want %v", err, nil)
		}
	}
}

func TestErrorMessage(t *testing.T) {
	tests := []struct {
		code    int
		message string
		details []map[string]interface{}
		want    string
	}{
		{
			code:    500,
			message: "Internal error occurred",
			details: []map[string]interface{}{
				map[string]interface{}{
					"@type":  "type.googleapis.com/google.rpc.DebugInfo",
					"detail": "[ORIGINAL ERROR]",
				},
			},
			want: strings.Join([]string{
				"{",
				"  \"error\": {",
				"    \"code\": 500,",
				"    \"message\": \"Internal error occurred\"",
				"  }",
				"}",
			}, "\n"),
		},
		{
			code:    400,
			message: "Invalid Argument",
			details: []map[string]interface{}{
				map[string]interface{}{
					"@type":  "type.googleapis.com/google.rpc.InvalidArgument",
					"detail": "[ORIGINAL ERROR]",
				},
			},
			want: strings.Join([]string{
				`{`,
				`  "error": {`,
				`    "code": 400,`,
				`    "message": "Invalid Argument",`,
				`    "details": [`,
				`      {`,
				`        "@type": "type.googleapis.com/google.rpc.InvalidArgument",`,
				`        "detail": "[ORIGINAL ERROR]"`,
				`      }`,
				`    ]`,
				`  }`,
				`}`,
			}, "\n"),
		},
		{
			code:    400,
			message: "Failed precondition",
			details: []map[string]interface{}{
				map[string]interface{}{
					"@type":  "type.googleapis.com/google.rpc.FailedPrecondition",
					"detail": "[ORIGINAL ERROR]",
				},
			},
			want: strings.Join([]string{
				`{`,
				`  "error": {`,
				`    "code": 400,`,
				`    "message": "Failed precondition",`,
				`    "details": [`,
				`      {`,
				`        "@type": "type.googleapis.com/google.rpc.FailedPrecondition",`,
				`        "detail": "[ORIGINAL ERROR]"`,
				`      }`,
				`    ]`,
				`  }`,
				`}`,
			}, "\n"),
		},
	}
	for _, tc := range tests {
		in := &PublicError{}
		in.Error.Code = tc.code
		in.Error.Message = tc.message
		in.Error.Details = tc.details
		got := errorMessage(in)
		if got != tc.want {
			t.Errorf("errorMessages got %v, want %v", got, tc.want)
		}
	}
}

func TestReceiveStream(t *testing.T) {
	tests := []struct {
		body      string
		wantFiles []string
		name      string
	}{
		{
			name: "only settings",
			body: strings.Join([]string{
				`[`,
				`  {`,
				`    "files": {`,
				`      "configFiles": {`,
				`        "configFiles":`,
				`          [`,
				`            {`,
				`	             "filePath": "settings/settings.yaml",`,
				`	             "settings": {`,
				`		             "actionsForFamilyUpdated": true,`,
				`		             "category": "GAMES_AND_TRIVIA",`,
				`		             "defaultLocale": "en",`,
				`		             "localizedSettings": {`,
				`					         "developerEmail": "dschrute@gmail.com",`,
				`				           "developerName": "Dwight Schrute",`,
				` 	       			   "displayName": "Mike Simple Question",`,
				`       				   "fullDescription": "Test Full Description",`,
				`				           "sampleInvocations": [`,
				`					           "Talk to Mike Simple Question"`,
				`				           ],`,
				`				           "smallLogoImage": "$resources.images.square"`,
				`	                },`,
				`	               "projectId": "placeholder_project"`,
				`              }`,
				`            }`,
				`          ]`,
				`        }`,
				`     }`,
				`  }`,
				`]`}, "\n"),
			wantFiles: []string{"settings/settings.yaml"},
		},
		{
			name: "configFiles and dataFiles",
			body: strings.Join([]string{
				`[`,
				`  {`,
				`    "files": {`,
				`      "configFiles": {`,
				`        "configFiles": [`,
				`          {`,
				`            "filePath": "settings/settings.yaml",`,
				`	           "settings": {`,
				`     		     "category": "GAMES_AND_TRIVIA"`,
				`            }`,
				`          },`,
				`          {`,
				`	           "filePath": "custom/global/actions.intent.MAIN.yaml",`,
				`	           "globalIntentEvent": {`,
				`		           "handler": {`,
				`  		           "staticPrompt": {`,
				`			           "candidates": [`,
				`   			         {`,
				`				             "promptResponse": {`,
				`					             "firstSimple": {`,
				`    					             "variants": [`,
				` 						                 {`,
				`					                     "speech": "$resources.strings.WELCOME",`,
				`						                   "text": "$resources.strings.WELCOME"`,
				`        						           }`,
				`					                  ]`,
				`					                }`,
				`				                }`,
				`    				         }`,
				`			             ]`,
				` 			         }`,
				`		            },`,
				`	              "transitionToScene": "questionpage"`,
				`	           }`,
				`          },`,
				`          {`,
				`	           "filePath": "settings/es/settings.yaml",`,
				`	           "settings": {`,
				`		           "localizedSettings": {`,
				`			           "displayName": "Mike Pregunta simple",`,
				`			           "fullDescription": "Descripción completa de la muestra"`,
				`		           }`,
				`	           }`,
				`          }`,
				`        ]`,
				`      }`,
				`    }`,
				`  },`,
				`  {`,
				`    "files": {`,
				`      "dataFiles": {`,
				`        "dataFiles": [`,
				`          {`,
				`	           "filePath": "resources/images/foo.png",`,
				`            "contentType": "images/png",`,
				`            "payload": ""`,
				`          }`,
				`        ]`,
				`      }`,
				`    }`,
				`  }`,
				`]`}, "\n"),
			wantFiles: []string{"resources/images/foo.png", "settings/es/settings.yaml", "custom/global/actions.intent.MAIN.yaml"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup directory where receiveStream will write files to.
			dirName, err := ioutil.TempDir(testutils.TestTmpDir, "actions-sdk-cli-project-folder")
			if err != nil {
				t.Fatalf("Can't create temporary directory under %q: %v", testutils.TestTmpDir, err)
			}
			defer func() {
				if err := os.RemoveAll(dirName); err != nil {
					t.Fatalf("Can't remove temp directory: %v", err)
				}
			}()
			proj := studio.New([]byte("secret"), dirName)
			seen := map[string]bool{}
			if err := receiveStream(proj, strings.NewReader(tc.body), false, seen); err != nil {
				t.Errorf("receiveStream returned %v, but expected to return %v", err, nil)
			}
			for _, v := range tc.wantFiles {
				osPath := filepath.FromSlash(v)
				// TODO: Verify the content of the written file
				_, err := ioutil.ReadFile(filepath.Join(proj.ProjectRoot(), osPath))
				if err != nil {
					t.Errorf("receiveStream expected to write file to disk, but got %v", err)
				}
				if !seen[v] {
					t.Errorf("receiveStream expected to mark file as seen, but did not")
				}
			}
		})
	}
}

func TestFindExtra(t *testing.T) {
	tests := []struct {
		a    map[string][]byte
		b    map[string]bool
		want []string
	}{
		{
			a: map[string][]byte{
				"settings/settings.yaml":           []byte("abc"),
				"manifest.yaml":                    []byte("abc"),
				"resources/strings/en/bundle.yaml": []byte("abc"),
			},
			b: map[string]bool{
				"settings/settings.yaml":           true,
				"manifest.yaml":                    true,
				"resources/strings/en/bundle.yaml": true,
			},
			want: nil,
		},
		{
			a: map[string][]byte{
				"settings/settings.yaml": []byte("abc"),
				"manifest.yaml":          []byte("abc"),
			},
			b: map[string]bool{
				"settings/settings.yaml":           true,
				"manifest.yaml":                    true,
				"resources/strings/en/bundle.yaml": true,
			},
			want: nil,
		},
		{
			a: map[string][]byte{
				"settings/settings.yaml":           []byte("abc"),
				"manifest.yaml":                    []byte("abc"),
				"resources/strings/en/bundle.yaml": []byte("abc"),
			},
			b: map[string]bool{
				"settings/settings.yaml": true,
				"manifest.yaml":          true,
			},
			want: []string{"resources/strings/en/bundle.yaml"},
		},
	}
	for _, tc := range tests {
		got := findExtra(tc.a, tc.b)
		sort.Strings(got)
		sort.Strings(tc.want)
		if diff := cmp.Diff(tc.want, got); diff != "" {
			t.Errorf("findExtra didn't return correct result: diff (-want, +got)\n%s", diff)
		}
	}
}
