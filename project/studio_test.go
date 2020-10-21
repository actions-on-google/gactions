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

package studio

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/actions-on-google/gactions/api/testutils"
	"github.com/actions-on-google/gactions/project"
	"github.com/google/go-cmp/cmp"
)

type MockStudio struct {
	root         string
	files        map[string][]byte
	clientSecret []byte
	projectID    string
}

func (p MockStudio) ProjectID() string {
	return p.projectID
}

func NewMock(root string) MockStudio {
	m := MockStudio{}
	m.root = root
	m.files = map[string][]byte{}
	for k, v := range configFiles {
		m.files[k] = v
	}
	for k, v := range dataFiles {
		m.files[k] = v
	}
	// Add extra files that should be ignored by the CLI
	m.files[".git"] = []byte("...")
	folders := []string{
		"verticals",
		"actions",
		"custom",
		"custom/global",
		"custom/intents",
		"custom/prompts",
		"custom/scenes",
		"custom/types",
		"webhooks/",
	}
	for _, v := range folders {
		m.files[path.Join(v, ".DS_Store")] = []byte("...")
	}
	m.files["webhooks/webhook1/.git"] = []byte("...")
	return m
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

var configFiles = map[string][]byte{
	"verticals/character_alarm.yaml":           []byte("name: foo"),
	"actions/actions.yaml":                     []byte("intent: bar"),
	"manifest.yaml":                            []byte("version: 1"),
	"custom/global/actions.intent.CANCEL.yaml": []byte("transitionToScene: actions.scene.END_CONVERSATION"),
	"custom/intents/help.yaml":                 []byte("phrase: hello"),
	"custom/intents/ru/help.yaml":              []byte("phrase: hello"),
	"custom/prompts/foo.yaml":                  []byte("prompt: yes"),
	"custom/prompts/ru/foo.yaml":               []byte("prompt: yes"),
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
}

var dataFiles = map[string][]byte{
	"resources/images/a.png":         []byte("abc123"),
	"resources/audio/b.mp3":          []byte("cde456"),
	"resources/audio/c.wav":          []byte("mno234"),
	"webhooks/webhook1/index.js":     []byte("exports.hello = functions.https.onRequest(app);"),
	"webhooks/webhook1/package.json": []byte("{}"),
	"resources/animations/d.flr":     []byte("fgh789"),
}

func TestAlreadySetup(t *testing.T) {
	proj := New([]byte{}, ".")
	tests := []struct {
		dirExists   bool
		dirEmpty    bool
		wantIsSetup bool
	}{
		{
			dirExists:   true,
			dirEmpty:    false,
			wantIsSetup: true,
		},
		{
			dirExists:   true,
			dirEmpty:    true,
			wantIsSetup: false,
		},
		{
			dirExists:   false,
			dirEmpty:    true,
			wantIsSetup: false,
		},
		{
			dirExists:   false,
			dirEmpty:    false,
			wantIsSetup: false,
		},
	}
	for _, tc := range tests {
		var dirName string
		if tc.dirExists {
			var err error
			dirName, err = ioutil.TempDir(testutils.TestTmpDir, "actions-sdk-cli-project-folder")
			if err != nil {
				t.Errorf("Can't create temporary directory under %q: %v", testutils.TestTmpDir, err)
			}
			defer os.RemoveAll(dirName)
			if !tc.dirEmpty {
				tempFile, err := ioutil.TempFile(dirName, "actions-sdk-*.yaml")
				fmt.Printf("tempFile = %v\n", tempFile.Name())
				if err != nil {
					t.Fatalf("can not create tempfile. got %v", err)
				}
				defer tempFile.Close()
			}
		}
		if isSetup := proj.AlreadySetup(dirName); isSetup != tc.wantIsSetup {
			t.Errorf("AlreadySetup returned %v, expected %v, when project directory exists (%v) and is empty (%v)", isSetup, tc.wantIsSetup, tc.dirExists, tc.dirEmpty)
		}
	}
}

func TestFilesWhenDirectoryManifestPresent(t *testing.T) {
	dirName, err := ioutil.TempDir(testutils.TestTmpDir, "actions-sdk-cli-project-folder")
	if err != nil {
		t.Fatalf("Can't create temporary directory under %q: %v", testutils.TestTmpDir, err)
	}
	proj := New([]byte("secret"), dirName)
	defer os.RemoveAll(dirName)
	// first file
	err = ioutil.WriteFile(filepath.Join(dirName, "manifest.yaml"), []byte("hello"), 0666)
	if err != nil {
		t.Fatalf("Can't write a file under %q: %v", dirName, err)
	}
	// second file
	err = ioutil.WriteFile(filepath.Join(dirName, "second-file.yaml"), []byte("world"), 0666)
	if err != nil {
		t.Fatalf("Can't create a file under %q: %v", dirName, err)
	}
	got, err := proj.Files()
	if err != nil {
		t.Errorf("Files got %v, want %v\n", err, nil)
	}
	gotNorm := make(map[string][]byte)
	// strip parent paths to eliminate undeterminism
	for k, v := range got {
		gotNorm[filepath.Base(k)] = v
	}
	want := map[string][]byte{
		"manifest.yaml":    []byte("hello"),
		"second-file.yaml": []byte("world"),
	}
	if !cmp.Equal(gotNorm, want) {
		t.Errorf("Files returned incorrect files, got %v, want %v", got, want)
	}
}

func TestClientSecretJSON(t *testing.T) {
	dirName, err := ioutil.TempDir(testutils.TestTmpDir, "actions-sdk-cli-project-folder")
	if err != nil {
		t.Fatalf("Can't create temporary directory under %q: %v", testutils.TestTmpDir, err)
	}
	defer os.RemoveAll(dirName)
	want := "{client_id: 123456789}"
	err = ioutil.WriteFile(filepath.Join(dirName, "test-client-secret.json"), []byte(want), 0666)
	if err != nil {
		t.Fatalf("Can't create a file under %q: %v", dirName, err)
	}
	proj := Studio{clientSecretJSON: []byte(want)}
	got, err := proj.ClientSecretJSON()
	if err != nil {
		t.Errorf("ClientSecretJSON got %v, want %v", err, nil)
	}
	if string(got) != want {
		t.Errorf("ClientSecretJSON returned incorrect result, got %v, want %v", string(got), want)
	}
}

func TestConfigFiles(t *testing.T) {
	p := NewMock(".")
	want := configFiles
	files, _ := p.Files()
	got := ConfigFiles(files)
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("ConfigFiles returned %v, want %v, diff %v", got, want, diff)
	}
}

func TestDataFiles(t *testing.T) {
	p := NewMock(".")
	want := map[string][]byte{}
	// Server expects Cloud Functions to have the filePath stripped
	// (i.e. webhooks/myfunction/index.js -> ./index.js)
	for k, v := range dataFiles {
		if !strings.Contains(k, "resources/") {
			want[path.Base(k)] = v
		} else {
			want[k] = v
		}
	}
	p.files["webhooks/myfunction/node_modules/foo/foo.js"] = []byte("console.log('hello world');")
	got, err := DataFiles(p)
	if err != nil {
		t.Errorf("DataFiles got %v, want %v", err, nil)
	}
	if zipped, ok := got["webhooks/webhook1.zip"]; !ok {
		t.Errorf("DataFiles didn't include webhook1.zip into a map of data files: data files = %v", got)
	} else {
		r, err := zip.NewReader(bytes.NewReader(zipped), int64(len(zipped)))
		if err != nil {
			t.Fatalf("can not create a zip.NewReader: got %v", err)
		}
		for _, f := range r.File {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("can not open %v: got %v", f.Name, err)
			}
			b, err := ioutil.ReadAll(rc)
			if err != nil {
				t.Fatalf("can not read from %v: got %v", f.Name, err)
			}
			rc.Close()
			got[f.Name] = b
		}
		delete(got, "webhooks/webhook1.zip")
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("DataFiles returned %v, want %v, diff %v", got, want, diff)
	}
}

func TestAddInlineWebhooksReturnsErrorWithInvalidWebhookYaml(t *testing.T) {
	p := NewMock(".")
	p.files["webhooks/malformed_webhook.yaml"] = []byte(
		`
external_endpoint:
   base_url: https://google.com
  endpoint_api_version: 1
`)

	err := addInlineWebhooks(map[string][]byte{}, p.files, "")
	if err == nil || !strings.Contains(err.Error(), "malformed_webhook.yaml has incorrect syntax") {
		t.Errorf("Expected error not thrown")
	}
}

func TestProjectIDFound(t *testing.T) {
	want := "my_project123"
	files := map[string][]byte{
		"settings/settings.yaml": []byte(fmt.Sprintf("projectId: %v", want)),
	}
	proj := MockStudio{files: files}
	got, err := ProjectID(proj)
	if err != nil {
		t.Errorf("ProjectID returned %v, want %v", err, nil)
	}
	if got != want {
		t.Errorf("ProjectID returned %v, want %v", got, want)
	}
}

func TestProjectIDSNotFound(t *testing.T) {
	files := map[string][]byte{
		"manifest.yaml": []byte("version: 1"),
	}
	proj := MockStudio{files: files}
	_, err := ProjectID(proj)
	if err == nil {
		t.Errorf("When settings.yaml is absent, ProjectID returned %v, want %v", err, errors.New("can't find a projectId: settings.yaml not found"))
	}
	files = map[string][]byte{
		"settings.yaml": []byte("display_name: foo"),
	}
	proj = MockStudio{files: files}
	_, err = ProjectID(proj)
	if err == nil {
		t.Errorf("When settings.yaml doesn't contain projectId field, ProjectID returned %v, want %v", err, errors.New("projectId is not present in the settings file"))
	}
}

func TestUnixPath(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{
			in:   "/google/assistant/aog/sdk/",
			want: "/google/assistant/aog/sdk/",
		},
		{
			in:   "\\google\\assistant\\aog\\sdk",
			want: "/google/assistant/aog/sdk",
		},
		{
			in:   "foo/",
			want: "foo/",
		},
		{
			in:   "foo\\",
			want: "foo/",
		},
		{
			in:   "dir\\to\\foo bar",
			want: "dir/to/foo bar",
		},
	}
	for _, tc := range tests {
		if got := winToUnix(tc.in); got != tc.want {
			t.Errorf("unixPath returned %v, want %v", got, tc.want)
		}
	}
}

func TestSetProjectID(t *testing.T) {
	tests := []struct {
		settings []byte
		flag     string
		want     string
	}{
		{ // Case 1.
			settings: nil,
			flag:     "",
			want:     "",
		},
		{ // Case 2.
			settings: []byte("projectId: placeholder_project"),
			flag:     "",
			want:     "placeholder_project",
		},
		{ // Case 3.
			settings: []byte("projectId: hello-world"),
			flag:     "",
			want:     "hello-world",
		},
		{ // Case 4.
			settings: nil,
			flag:     "foobar",
			want:     "foobar",
		},
		{ // Case 5.
			settings: []byte("projectId: placeholder_project"),
			flag:     "hello-world",
			want:     "hello-world",
		},
		{
			settings: []byte("projectId: hello-world"),
			flag:     "foobar",
			want:     "foobar",
		},
	}
	for _, tc := range tests {
		t.Run("foo", func(t *testing.T) {
			dirName, err := ioutil.TempDir(testutils.TestTmpDir, "actions-sdk-cli-project-folder")
			if err != nil {
				t.Fatalf("Can't create temporary directory under %q: %v", testutils.TestTmpDir, err)
			}
			defer func() {
				if err := os.RemoveAll(dirName); err != nil {
					t.Fatalf("Can't remove temp directory: %v", err)
				}
			}()
			if tc.settings != nil {
				fp := filepath.Join(dirName, "settings", "settings.yaml")
				if err := os.MkdirAll(filepath.Dir(fp), 0750); err != nil {
					t.Fatalf("Can't create settings directory: %v", err)
				}
				if err := ioutil.WriteFile(fp, tc.settings, 0640); err != nil {
					t.Fatalf("Can't create settings file: %v", err)
				}
			}
			studio := New([]byte{}, dirName)
			if err := (&studio).SetProjectID(tc.flag); err != nil && tc.settings != nil {
				t.Errorf("SetProjectID returned %v, want %v", err, nil)
			}
			if studio.projectID != tc.want {
				t.Errorf("Project ID is %v after calling SetProjectID, but want %v", studio.projectID, tc.want)
			}
		})
	}
}

func cloudFuncZip(t *testing.T) []byte {
	t.Helper()
	files := map[string][]byte{}
	for k, v := range dataFiles {
		if strings.Contains(k, ".js") {
			files[k] = v
		}
	}
	b, err := zipFiles(files)
	if err != nil {
		t.Fatalf("Can not zip %v: %v", files, err)
	}
	return b
}

func TestWriteToDiskToNonEmptyDir(t *testing.T) {
	tests := []struct {
		user  string
		force bool
		name  string
	}{
		{
			user:  "yes",
			force: false,
			name:  "User says yes",
		},
		{
			user:  "no",
			name:  "User says no",
			force: false,
		},
		{
			user:  "",
			force: true,
			name:  "Force is true",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dirName, err := ioutil.TempDir(testutils.TestTmpDir, "actions-sdk-cli-project-folder")
			if err != nil {
				t.Fatalf("Can't create temporary directory under %q: %v", testutils.TestTmpDir, err)
			}
			defer os.RemoveAll(dirName)
			proj := NewMock(dirName)
			og := askYesNo
			askYesNo = func(msg string) (string, error) {
				return tc.user, nil
			}
			defer func() {
				askYesNo = og
			}()
			if err := ioutil.WriteFile(filepath.Join(dirName, "manifest.yaml"), []byte("version:2.0"), 0640); err != nil {
				t.Fatalf("Can't write %v: %v", filepath.Join(dirName, "manifest.yaml"), err)
			}
			if err := WriteToDisk(proj, "manifest.yaml", "", []byte("version:1.0"), tc.force); err != nil {
				t.Errorf("WriteToDisk returned %v, want %v", err, nil)
			}
			if tc.user == "yes" || tc.force {
				if !exists(filepath.Join(dirName, "manifest.yaml")) {
					t.Errorf("WriteToDisk didn't write %v to disk", filepath.Join(dirName, "manifest.yaml"))
				}
				b, err := ioutil.ReadFile(filepath.Join(dirName, "manifest.yaml"))
				if err != nil {
					t.Errorf("Failed to read %v: %v", filepath.Join(dirName, "manifest.yaml"), err)
				}
				if len(b) == 0 {
					t.Errorf("WriteToDisk wrote empty file %v", filepath.Join(dirName, "manifest.yaml"))
				}
			}
		})
	}
}

func TestWriteToDiskToEmptyDir(t *testing.T) {
	tests := []struct {
		path        string
		contentType string
		payload     []byte
		wantFiles   []string
		name        string
	}{
		{
			path:        "webhooks/webhook1.zip",
			contentType: "application/zip;zip_type=cloud_function",
			payload:     cloudFuncZip(t),
			wantFiles:   []string{"webhooks/webhook1/index.js", "webhooks/webhook1/package.json"},
			name:        "Webhook.zip",
		},
		{
			path:        "settings/en/settings.yaml",
			contentType: "",
			payload:     []byte("projectId: hello-world"),
			wantFiles:   []string{"settings/en/settings.yaml"},
			name:        "Settings",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dirName, err := ioutil.TempDir(testutils.TestTmpDir, "actions-sdk-cli-project-folder")
			if err != nil {
				t.Fatalf("Can't create temporary directory under %q: %v", testutils.TestTmpDir, err)
			}
			defer os.RemoveAll(dirName)
			proj := NewMock(dirName)
			if err := WriteToDisk(proj, tc.path, tc.contentType, tc.payload, false); err != nil {
				t.Errorf("WriteToDisk got %v, want %v", err, nil)
			}
			for _, f := range tc.wantFiles {
				fp := path.Join(dirName, f)
				fp = filepath.FromSlash(fp)
				if !exists(fp) {
					t.Errorf("WriteToDisk didn't write %v to disk", fp)
				}
				b, err := ioutil.ReadFile(fp)
				if err != nil {
					t.Errorf("Failed to read %v: %v", fp, err)
				}
				if len(b) == 0 {
					t.Errorf("WriteToDisk created an empty file %v, want not empty", fp)
				}
			}
		})
	}
}

func TestFindProjectRoot(t *testing.T) {
	tests := []struct {
		names []string
		err   error
		cwd   string
		name  string
	}{
		{
			names: []string{
				"manifest.yaml",
				filepath.Join("settings", "settings.yaml"),
				filepath.Join("mywebhook", "index.js"),
			},
			err:  nil,
			cwd:  "",
			name: "manifest found and cwd is .",
		},
		{
			names: []string{
				filepath.Join("settings", "settings.yaml"),
				filepath.Join("verticals", "foo.yaml"),
			},
			err:  errors.New("manifest.yaml not found"),
			cwd:  "",
			name: "manifest not found",
		},
		{
			names: []string{
				filepath.Join("settings", "settings.yaml"),
				filepath.Join("verticals", "foo.yaml"),
				"manifest.yaml",
			},
			cwd:  "settings",
			err:  nil,
			name: "manifest found and cwd is settings",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dirName, err := ioutil.TempDir(testutils.TestTmpDir, "actions-sdk-cli-project-folder")
			if err != nil {
				t.Fatalf("Can't create temporary directory under %q: %v", testutils.TestTmpDir, err)
			}
			defer os.RemoveAll(dirName)
			for _, f := range tc.names {
				if err := os.MkdirAll(filepath.Join(dirName, filepath.Dir(f)), 0777); err != nil {
					t.Errorf("Can't create a directory %v, got %v", filepath.Join(dirName, filepath.Dir(f)), err)
				}
				if err := ioutil.WriteFile(filepath.Join(dirName, f), []byte("hello"), 0666); err != nil {
					t.Fatalf("Can't write a file under %q: %v", dirName, err)
				}
			}
			wkdir := dirName
			if tc.cwd != "" {
				wkdir = filepath.Join(wkdir, tc.cwd)
			}
			if err := os.Chdir(wkdir); err != nil {
				t.Errorf("Could not cd into %v: %v", wkdir, err)
			}
			got, err := FindProjectRoot()
			if tc.err == nil {
				if got != dirName {
					t.Errorf("findProjectRoot found %v as root, but should get %v", dirName, got)
				}
			} else {
				if err == nil {
					t.Errorf("findProjectRoot got %v, want %v", err, tc.err)
				}
			}
		})
	}
}
