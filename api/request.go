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

// Package request represents an SDK request.
package request

import (
	"encoding/base64"
	"fmt"
	"mime"
	"path"
	"path/filepath"
	"sort"

	"github.com/actions-on-google/gactions/api/yamlutils"
	"github.com/actions-on-google/gactions/log"
	"github.com/actions-on-google/gactions/project/studio"
)

const (
	// MaxChunkSizeBytes specifies the max size limit on JSON payload for a single request/response in the stream.
	// It's enforced by server.
	MaxChunkSizeBytes = 10 * 1024 * 1024
	// Padding accounts for bytes from surrounding JSON fields coming from the request schema;
	// 512 Kb buffer is a very generous upper-bound and I don't think we'll hit it in practice.
	Padding = 512 * 1024
)

// EncryptSecret returns a map representing a EncryptSecret request populated with clientSecret field.
func EncryptSecret(secret string) map[string]interface{} {
	return map[string]interface{}{
		"clientSecret": secret,
	}
}

// DecryptSecret returns a map representing a DecryptSecret request populated with encryptedClientSecret field.
func DecryptSecret(secret string) map[string]interface{} {
	return map[string]interface{}{
		"encryptedClientSecret": secret,
	}
}

// ReadDraft returns a map representing a ReadDraft request populated with name field.
func ReadDraft(name, keyVersion string) map[string]interface{} {
	req := map[string]interface{}{
		"name": fmt.Sprintf("projects/%v/draft", name),
	}
	if keyVersion != "" {
		req["clientSecretEncryptionKeyVersion"] = keyVersion
	}
	return req
}

// WriteDraft returns a map representing a WriteDraft request populated with name field.
func WriteDraft(name string) map[string]interface{} {
	return map[string]interface{}{
		"parent": fmt.Sprintf("projects/%v", name),
	}
}

// WritePreview returns a map representing a WriteDraft request populated with name and sandbox fields.
func WritePreview(name string, sandbox bool) map[string]interface{} {
	v := map[string]interface{}{}
	v["parent"] = fmt.Sprintf("projects/%v", name)
	v["previewSettings"] = map[string]interface{}{
		"sandbox": sandbox,
	}
	return v
}

// CreateVersion returns a map representing a WriteVersion request populated with name and sandbox fields.
func CreateVersion(name string, channel string) map[string]interface{} {
	return map[string]interface{}{
		"parent":          fmt.Sprintf("projects/%v", name),
		"release_channel": channel,
	}
}

// ReadVersion returns a map representing a ReadVersion request populated with name and versionId fields.
func ReadVersion(name string, versionID string) map[string]interface{} {
	return map[string]interface{}{
		"name": fmt.Sprintf("projects/%v/versions/%v", name, versionID),
	}
}

// addConfigFiles adds configFiles w/o a resource bundle to a request.
func addConfigFiles(req map[string]interface{}, configFiles map[string][]byte, root string) error {
	cfgs := make(map[string][]interface{})
	var keys []string
	for k := range configFiles {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, filename := range keys {
		content := configFiles[filename]
		log.Infof("Adding %v to configFiles request\n", filepath.Join(root, filename))
		mp, err := yamlutils.UnmarshalYAMLToMap(content)
		if err != nil {
			return fmt.Errorf("%v has incorrect syntax: %v", filepath.Join(root, filename), err)
		}
		m := make(map[string]interface{})
		m["filePath"] = filename
		switch {
		case studio.IsAccountLinkingSecret(filename):
			m["accountLinkingSecret"] = mp
		case studio.IsManifest(filename):
			m["manifest"] = mp
		case studio.IsSettings(filename):
			m["settings"] = mp
		case studio.IsActions(filename):
			m["actions"] = mp
		case studio.IsWebhookDefinition(filename):
			m["webhook"] = mp
		case studio.IsIntent(filename):
			m["intent"] = mp
		case studio.IsGlobal(filename):
			m["globalIntentEvent"] = mp
		case studio.IsType(filename):
			m["type"] = mp
		case studio.IsPrompt(filename):
			m["staticPrompt"] = mp
		case studio.IsScene(filename):
			m["scene"] = mp
		case studio.IsVertical(filename):
			m["verticalSettings"] = mp
		case studio.IsResourceBundle(filename):
			m["resourceBundle"] = mp
		default:
			return fmt.Errorf("failed to add %v to a request", filepath.Join(root, filename))
		}
		cfgs["configFiles"] = append(cfgs["configFiles"], m)
	}
	req["files"] = map[string]interface{}{
		"configFiles": cfgs,
	}
	return nil
}

// addDataFiles adds a data files from the chunk to a request.
func addDataFiles(req map[string]interface{}, chunk map[string][]byte, root string) error {
	dfs := map[string][]interface{}{}
	for filename, content := range chunk {
		log.Infof("Adding %v to dataFiles request\n", filepath.Join(root, filename))
		if path.Ext(filename) == ".zip" {
			m := map[string]interface{}{
				"filePath":    filename,
				"contentType": "application/zip;zip_type=cloud_function",
				"payload":     content,
			}
			dfs["dataFiles"] = append(dfs["dataFiles"], m)
			continue
		}
		if path.Ext(filename) == ".flr" {
			m := map[string]interface{}{
				"filePath":    filename,
				"contentType": "x-world/x-vrml",
				"payload":     content,
			}
			dfs["dataFiles"] = append(dfs["dataFiles"], m)
			continue
		}
		mime := mime.TypeByExtension(path.Ext(filename))
		switch mime {
		case "audio/mpeg", "image/jpeg", "image/png", "audio/wav", "audio/x-wav":
			{
				m := map[string]interface{}{
					"filePath":    filename,
					"contentType": mime,
					"payload":     content,
				}
				dfs["dataFiles"] = append(dfs["dataFiles"], m)
			}
		default:
			log.Warnf("Can't recognize an extension for %v. The supported extensions are audio/mpeg, image/jpeg, " +
				"image/png, audio/wav, audio/x-wav found %v", filepath.Join(root, filename), mime)
		}
	}
	if len(dfs) > 0 {
		req["files"] = map[string]interface{}{
			"dataFiles": dfs,
		}
	}
	return nil
}

// SDKStreamer provides an interface to obtain the next JSON request that needs to be sent to
// SDK server during HTTP stream. SDK, ESF and GFE each have their own requirements on the
// payload and this type implements them.
type SDKStreamer struct {
	files           map[string][]byte
	sizes           map[string]int // sizes contains a size that a file occupies in a JSON request
	dataFilenames   []string
	configFilenames []string
	makeRequest     func() map[string]interface{}
	root            string
	i               int // index of current item in configFilesnames
	j               int // index of current item in dataFilenames
	chunkSize       int
}

// NewStreamer returns an instance of SDKStreamer, initialized with all of the variables
// from its arguments. Function expects configFiles to have at least base settings and manifest files.
func NewStreamer(configFiles map[string][]byte, dataFiles map[string][]byte, makeRequest func() map[string]interface{}, root string, chunkSize int) SDKStreamer {
	files := map[string][]byte{}
	sizes := map[string]int{}
	var cfgnames, dfnames []string

	for k, v := range configFiles {
		files[k] = v
		cfgnames = append(cfgnames, k)
		sizes[k] = len(v)
	}
	for k, v := range dataFiles {
		files[k] = v
		dfnames = append(dfnames, k)
		// Marshal function of JSON library (https://golang.org/pkg/encoding/json/#Marshal) encodes
		// []byte as a base-64 encoded string. This adds an extra memory overhead when the map is
		// converted to JSON. Each DataFile is []byte, so this is a good approximation.
		sizes[k] = len(base64.StdEncoding.EncodeToString(v))
	}
	// We need to sort config files and datafiles based on their size in bytes.
	// However, settings and manifest files must be inside of the first request,
	// so these two files take precedence.
	sortConfigFiles(cfgnames, files, sizes)
	sort.Slice(dfnames, func(i int, j int) bool {
		return sizes[dfnames[i]] < sizes[dfnames[j]]
	})

	return SDKStreamer{
		files:           files,
		dataFilenames:   dfnames,
		configFilenames: cfgnames,
		makeRequest:     makeRequest,
		root:            root,
		chunkSize:       chunkSize,
		sizes:           sizes,
	}
}

func sortConfigFiles(cfgnames []string, files map[string][]byte, sizes map[string]int) {
	var pos []int
	for i, v := range cfgnames {
		if studio.IsSettings(v) || studio.IsManifest(v) {
			pos = append(pos, i)
		}
	}
	moveToFront(cfgnames, pos)
	needSort := cfgnames[len(pos):]
	sort.Slice(needSort, func(i int, j int) bool {
		return sizes[needSort[i]] < sizes[needSort[j]]
	})
	for i, v := range needSort {
		cfgnames[i+len(pos)] = v
	}
}

func moveToFront(a []string, ps []int) {
	for i := 0; i < len(ps); i++ {
		a[i], a[ps[i]] = a[ps[i]], a[i]
	}
}

// HasNext returns true if there is still another request in the stream.
func (s SDKStreamer) HasNext() bool {
	return (s.i + s.j) < len(s.files)
}

// nextChunk returns the next "chunk" of config files such that
// the sum of the size of each individual config file in the chunk
// is less than s.chunkSize.
func (s *SDKStreamer) nextChunk(a []string, next int) map[string][]byte {
	chunk := map[string][]byte{}
	curSize := 0
	i := 0
	for curSize < s.chunkSize && i+next < len(a) {
		name := a[next+i]
		content := s.files[name]
		curSize += s.sizes[name]
		if curSize > s.chunkSize {
			break
		}
		chunk[name] = content
		i++
	}
	return chunk
}

func (s *SDKStreamer) nextConfigFiles(req map[string]interface{}) error {
	if s.i == 0 {
		log.Outln("Sending configuration files...")
	}
	chunk := s.nextChunk(s.configFilenames, s.i)
	if len(chunk) == 0 {
		return fmt.Errorf("%v exceeds the limit of %v bytes", s.configFilenames[s.i], s.chunkSize)
	}
	if err := addConfigFiles(req, chunk, s.root); err != nil {
		return err
	}
	s.i += len(chunk)
	return nil
}

func (s *SDKStreamer) nextDataFiles(req map[string]interface{}) error {
	if s.j == 0 {
		log.Outln("Sending resources...")
	}
	chunk := s.nextChunk(s.dataFilenames, s.j)
	if len(chunk) == 0 {
		return fmt.Errorf("%v exceeds the limit of %v bytes", s.dataFilenames[s.j], s.chunkSize)
	}
	if err := addDataFiles(req, chunk, s.root); err != nil {
		return err
	}
	s.j += len(chunk)
	return nil
}

// Next returns the next request to be sent to SDK server. It implements following requirements:
// 1. Send all config files
//   1a. First request will have manifest and all of the settings files (i.e. localized and base)
//   1b. Each config file request is less than s.chunkSize.
// 2. Send all of data files in one or several requests. Each request will be less than s.chunkSize.
// It will return an error if the size of the payload is larger than s.chunkSize.
func (s *SDKStreamer) Next() (map[string]interface{}, error) {
	req := s.makeRequest()
	if s.i < len(s.configFilenames) {
		if err := s.nextConfigFiles(req); err != nil {
			return nil, err
		}
	} else if s.j < len(s.dataFilenames) {
		if err := s.nextDataFiles(req); err != nil {
			return nil, err
		}
	}
	return req, nil
}
