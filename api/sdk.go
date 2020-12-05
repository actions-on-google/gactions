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

// Package sdk implements the adapter to talk to Actions SDK server.
package sdk

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/actions-on-google/gactions/api/apiutils"
	"github.com/actions-on-google/gactions/api/request"
	"github.com/actions-on-google/gactions/cmd/version"
	"github.com/actions-on-google/gactions/log"
	"github.com/actions-on-google/gactions/project"
	"github.com/actions-on-google/gactions/project/studio"
	"gopkg.in/yaml.v2"
)

const (
	actionsProdURL             = "actions.googleapis.com"
	actionsConsoleProdURL      = "console.actions.google.com"
	encryptEndpoint            = "v2:encryptSecret"
	decryptEndpoint            = "v2:decryptSecret"
	listSampleProjectsEndpoint = "v2/sampleProjects"
	// Prod version of CurEnv
	Prod = "prod"
	// ProdChannel of AoG release
	ProdChannel = "actions.channels.Production"
	// AlphaChannel of AoG release
	AlphaChannel = "actions.channels.Alpha"
	// BetaChannel of AoG release
	BetaChannel = "actions.channels.ClosedBeta"
)

var (
	// CurEnv determines which version of the Actions API to call.
	CurEnv      = Prod
	consoleAddr = "https://" + urlMap[CurEnv]["consoleURL"]
	// Consumer holds the string identifying the caller to Google. This is based on a command line flag.
	Consumer = ""
	// responseBodyReadTimeout is a time limit to read body of HTTP response after response object is received.
	responseBodyReadTimeout = 5 * time.Second
	BuiltInReleaseChannels = map[string]string{
		ProdChannel:     "prod",
	}
)

var urlMap = map[string]map[string]string{
	Prod: map[string]string{
		"apiURL":     actionsProdURL,
		"consoleURL": actionsConsoleProdURL,
	},
}

// CreateVersionHTTPResponse represents the expected fields the CLI expects from the CreateVersion API.
// CLI will use those fields to print an output message to a user. All other fields from an API
// response will be ignored.
type CreateVersionHTTPResponse struct {
	Name string `json:"name"`
}

// WriteDraftHTTPResponse represents the expected fields the CLI expects from the WriteDraft API.
// CLI will use those fields to print an output message to a user. All other fields from an API
// response will be ignored.
type WriteDraftHTTPResponse struct {
	Name              string `json:"name"`
	ValidationResults struct {
		Results []validationResult `json:"results"`
	} `json:"validationResults"`
}

type validationResult struct {
	ValidationMessage string `json:"validationMessage"`
	ValidationContext struct {
		LanguageCode string `json:"languageCode"`
	} `json:"validationContext"`
}

// WritePreviewHTTPResponse represents the expected fields the CLI expects from the WritePreview
// API. CLI will use those fields to print an output message to a user. All other fields from
// an API response will be ignored.
type WritePreviewHTTPResponse struct {
	Name              string `json:"name"`
	ValidationResults struct {
		Results []validationResult `json:"results"`
	} `json:"validationResults"`
	SimulatorURL string `json:"simulatorUrl"`
}

// EncryptSecretHTTPResponse represents the expected fields the CLI expects from the EncryptSecret endpoint.
type EncryptSecretHTTPResponse struct {
	AccountLinkingSecret map[string]interface{} `json:"accountLinkingSecret"`
}

// PublicError represents a public error structure inside of an API response. All other fields
// (for example, containing the internal details) will be omitted.
type PublicError struct {
	Error struct {
		Code    int                      `json:"code,omitempty"`
		Message string                   `json:"message,omitempty"`
		Details []map[string]interface{} `json:"details,omitempty"`
	} `json:"error,omitempty"`
}

type configFiles struct {
	ConfigFiles []map[string]interface{} `json:"configFiles"`
}

type dataFiles struct {
	DataFiles []struct {
		Filepath    string `json:"filePath"`
		Payload     []byte `json:"payload"`
		ContentType string `json:"contentType"`
	} `json:"dataFiles"`
}

type streamRecord struct {
	Files struct {
		ConfigFiles *configFiles `json:"configFiles"`
		DataFiles   *dataFiles   `json:"dataFiles"`
	} `json:"files"`
}

func httpAddr(endpoint string) string {
	return "https://" + urlMap[CurEnv]["apiURL"] + "/" + endpoint
}

func writeDraftHTTPEndpoint(projectID string) string {
	return fmt.Sprintf("v2/projects/%s/draft:write", projectID)
}

func previewHTTPEndpoint(projectID string) string {
	return fmt.Sprintf("v2/projects/%s/preview:write", projectID)
}

func versionHTTPEndpoint(projectID string) string {
	return fmt.Sprintf("v2/projects/%s/versions:create", projectID)
}

func readDraftHTTPEndpoint(projectID string) string {
	return fmt.Sprintf("v2/projects/%s/draft:read", projectID)
}

func readVersionHTTPEndpoint(projectID, versionID string) string {
	return fmt.Sprintf("v2/projects/%s/versions/%s:read", projectID, versionID)
}

func listReleaseChannelsHTTPEndpoint(projectID string) string {
	return fmt.Sprintf("v2/projects/%s/releaseChannels", projectID)
}

func listVersionsHTTPEndpoint(projectID string) string {
	return fmt.Sprintf("v2/projects/%s/versions", projectID)
}

func check(cfgs map[string][]byte) error {
	if len(cfgs) == 0 {
		return errors.New("configuration files for your Action were not found")
	}
	// base settings file
	if _, ok := cfgs["settings/settings.yaml"]; !ok {
		return errors.New("settings/settings.yaml for your Action was not found")
	}
	if _, ok := cfgs["manifest.yaml"]; !ok {
		return errors.New("manifest.yaml for your Action was not found")
	}
	return nil
}

func printSize(req map[string]interface{}) {
	b, err := json.Marshal(req)
	if err != nil {
		log.Infof("Tried marshalling request into JSON: %v\n", err)
		return
	}
	log.Infof("Total request size is %v bytes.", len(b))
}

// sendFilesToServerJSON will stream series of requests based on proj to w.
// The function performs client-side streaming via HTTP/JSON. This is done by
// sending an array of JSON requests.
func sendFilesToServerJSON(p project.Project, w *io.PipeWriter, makeRequest func() map[string]interface{}) (err error) {
	// Important - must close w to avoid deadlock for the reader end of the pipe.
	defer func() {
		// Don't want to overwrite other errors raised in the func.
		// If any other error happened, then the PipeWriter error is not significant.
		err2 := w.Close()
		if err == nil {
			err = err2
		}
	}()
	files, err := p.Files()
	if err != nil {
		return err
	}
	configFiles := studio.ConfigFiles(files)
	dataFiles, err := studio.DataFiles(p)
	if err != nil {
		return err
	}
	if err := check(configFiles); err != nil {
		return err
	}
	encoder := json.NewEncoder(w)
	_, err = w.Write([]byte("["))
	if err != nil {
		return err
	}
	streamer := request.NewStreamer(configFiles, dataFiles, makeRequest, p.ProjectRoot(), request.MaxChunkSizeBytes-request.Padding)
	for streamer.HasNext() {
		req, err := streamer.Next()
		if err != nil {
			return err
		}
		printSize(req)
		if err = encoder.Encode(req); err != nil {
			// Ignore this error because it's possible for this error
			// to happen when server closed the connection (i.e. the read end of the pipe gets closed)
			// due to a failing internal server logic after processing of configuration files.
			log.Infof("Failed to send previous request: %v\n", err)
			return nil
		}
		if streamer.HasNext() {
			if _, err = w.Write([]byte(",")); err != nil {
				// Ignore this error because it's possible for this error
				// to happen when server closed the connection (i.e. the read end of the pipe gets closed)
				// due to a failing internal server logic after processing of configuration files.
				log.Infof("Failed to send previous request: %v\n", err)
				return nil
			}
		}
	}
	if _, err = w.Write([]byte("]")); err != nil {
		// Ignore this error because it's possible for this error
		// to happen when server closed the connection (i.e. the read end of the pipe gets closed)
		// due to a failing internal server logic after processing of the last data file.
		log.Infof("Failed to send previous request: %v\n", err)
		return nil
	}
	return err
}

// readBodyWithTimeout reads content from body until EOF is encountered, or timer expired.
// Timer starts when this function starts execution.
func readBodyWithTimeout(body io.Reader, timeout time.Duration) ([]byte, error) {
	// buf is initialized with 1 character to ensure a caller (Read) doesn't wait
	// for EOF to be sent from server.
	buf := make([]byte, 1)
	jsonString := ""
	// Buffered channels should protect against leaked go-routines.
	errCh := make(chan error, 1)
	go func() {
		for {
			n, err := body.Read(buf)
			if n > 0 {
				jsonString += string(buf)
			}
			if err != nil {
				errCh <- err
				break
			}
		}
	}()
	select {
	case <-time.After(timeout):
		return []byte(jsonString), nil
	case err := <-errCh:
		if err == io.EOF {
			return []byte(jsonString), nil
		}
		return nil, err
	}
}

// postprocessJSONResponse performs error handling of the JSON response, and also processes
// specific fields from the response body based on a callback function.
func postprocessJSONResponse(resp *http.Response, errCh chan error, proc func(body []byte) error) {
	body, err := readBodyWithTimeout(resp.Body, responseBodyReadTimeout)
	if err != nil {
		errCh <- err
		return
	}
	if resp.StatusCode != 200 {
		errCh <- parseError(body)
		return
	}
	// proc should perform a response specific processing; e.g. extracting specific fields. Only relevant if
	// if response code is 200.
	if err := proc(body); err != nil {
		errCh <- err
	}
	errCh <- nil
}

func parseError(body []byte) error {
	log.Debugln(string(body))
	publicError := &PublicError{}
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(publicError); err != nil {
		// This means the error is not a JSON. This happens when the API URL is malformed, and
		// one platform returns an HTML response. In this case, we print the HTML and disregard the json decoding error.
		return fmt.Errorf(string(body))
	}
	return fmt.Errorf("Server did not return HTTP 200.\n%v", errorMessage(publicError))
}

func errorMessage(in *PublicError) string {
	out := PublicError{}
	// Only allow details to be surfaced if the error code is 400.
	// 400 corresponds to gRPC FAILED_PRECONDITION and INVALID_ARGUMENT
	switch in.Error.Code {
	case 400:
		out.Error = in.Error
	// 403 is returned when user denied the permission to use API, which
	// is the case when they need to enable the API. The error message
	// contains a helpful info, including the link to the API manager.
	case 403, 404:
		out.Error.Message = in.Error.Message
		out.Error.Code = in.Error.Code
	default:
		out.Error.Message = "Internal error occurred"
		out.Error.Code = in.Error.Code
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		log.Warnf("%v\n", err)
		return ""
	}
	return string(b)
}

func printValidationResults(results []validationResult) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 2, 4, 2, ' ', 0)
	fmt.Fprintln(w, "  Locale\tValidation Result\t")
	for _, v := range results {
		fmt.Fprintf(w, "  %v\t%v\t\n", v.ValidationContext.LanguageCode, v.ValidationMessage)
	}
	fmt.Fprint(w)
	w.Flush()
}

func procWriteDraftResponse(body []byte) error {
	resp := &WriteDraftHTTPResponse{}
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(resp); err != nil {
		return errors.New(string(body))
	}
	if len(resp.ValidationResults.Results) > 0 {
		log.Warnln("Server found validation issues (however, your files were still pushed):")
		printValidationResults(resp.ValidationResults.Results)
	}
	return nil
}

// WriteDraftJSON implements WriteDraft functionality of the SDK server via HTTP/JSON streaming.
func WriteDraftJSON(ctx context.Context, proj project.Project) error {
	clientSecret, err := proj.ClientSecretJSON()
	if err != nil {
		return err
	}
	client, err := apiutils.NewHTTPClient(
		ctx,
		clientSecret,
		"",
	)
	if err != nil {
		return err
	}
	projectID := proj.ProjectID()
	log.Outf("Pushing files in the project %q to Actions Console. This may take a few minutes.\n", projectID)
	requestURL := httpAddr(writeDraftHTTPEndpoint(projectID))
	r, w := io.Pipe()
	errCh := make(chan error, 1)
	// This goroutine will exit after HTTP call is finished.
	// The sendFilesToServerJSON below and client.Post communicate via the pipe
	// and former will keep writing stream of bytes, which client post will
	// keep reading in a blocking fashion. sendFilesToServerJSON is guaranteed
	// to close the writer end of the pipe, thus unblocking the reader and allowing
	// the goroutine to exit.
	go func() {
		req, err := http.NewRequest("POST", requestURL, r)
		if err != nil {
			errCh <- err
			return
		}
		req.Header.Add("Content-Type", "application/json")
		// This is done to help server to select the quota attributed to a
		// projectID (i.e. developer's project), instead of the CLI project.
		req.Header.Add("X-Goog-User-Project", projectID)
		addClientHeaders(req)

		resp, err := client.Do(req)

		if err != nil {
			errCh <- err
			return
		}
		defer resp.Body.Close()
		postprocessJSONResponse(resp, errCh, func(body []byte) error {
			return procWriteDraftResponse(body)
		})
	}()
	if err := sendFilesToServerJSON(proj, w, func() map[string]interface{} {
		return request.WriteDraft(projectID)
	}); err != nil {
		return err
	}
	log.Outf("Waiting for server to respond...")
	err = <-errCh
	if err != nil {
		return err
	}
	log.DoneMsgln(fmt.Sprintf(`Files were pushed to Actions Console, and you can now view your project with this URL: %v/project/%v/overview. If you want to test your changes, run "gactions deploy preview", or navigate to the Test section in the Console.`, consoleAddr, projectID))
	return nil
}

func procWritePreviewResponse(body []byte) (string, error) {
	resp := &WritePreviewHTTPResponse{}
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(resp); err != nil {
		return "", errors.New(string(body))
	}
	if len(resp.ValidationResults.Results) > 0 {
		log.Warnln("Server found validation issues (however, your files were still pushed):")
		printValidationResults(resp.ValidationResults.Results)
	}
	simulatorURL := resp.SimulatorURL
	if simulatorURL == "" {
		log.Warnf("The API response body doesn't contain the simulator link.")
	}
	return simulatorURL, nil
}

// WritePreviewJSON implements WritePreview functionality of the SDK server via HTTP/JSON streaming.
func WritePreviewJSON(ctx context.Context, proj project.Project, sandbox bool) error {
	clientSecret, err := proj.ClientSecretJSON()
	if err != nil {
		return err
	}
	client, err := apiutils.NewHTTPClient(ctx, clientSecret, "")
	if err != nil {
		return err
	}
	projectID := proj.ProjectID()
	log.Outf("Deploying files in the project %q to Actions Console for preview. This may take a few minutes.\n", projectID)
	requestURL := httpAddr(previewHTTPEndpoint(projectID))
	r, w := io.Pipe()
	errCh := make(chan error, 1)
	var simulatorURL string
	// This goroutine will exit after HTTP call is finished.
	// The sendFilesToServerJSON below and client.Post communicate via the pipe
	// and former will keep writing stream of bytes, which client post will
	// keep reading in a blocking fashion. sendFilesToServerJSON is guaranteed
	// to close the writer end of the pipe, thus unblocking the reader and allowing
	// the goroutine to exit.
	go func() {
		req, err := http.NewRequest("POST", requestURL, r)
		if err != nil {
			errCh <- err
			return
		}
		req.Header.Add("Content-Type", "application/json")
		// This is done to help server select the quota attributed to a
		// projectID (i.e. developer's project), instead of the CLI project.
		// https://cloud.google.com/storage/docs/xml-api/reference-headers#xgooguserproject
		req.Header.Add("X-Goog-User-Project", projectID)
		// Sets timeout because Cloud Function deployment can take 1-2 minutes.
		const timeoutSec = "180"
		req.Header.Add("X-Server-Timeout", fmt.Sprintf("%v", timeoutSec))
		addClientHeaders(req)

		resp, err := client.Do(req)
		if err != nil {
			errCh <- err
			return
		}
		defer resp.Body.Close()
		postprocessJSONResponse(resp, errCh, func(body []byte) error {
			v, err := procWritePreviewResponse(body)
			simulatorURL = v
			return err
		})
	}()
	if err := sendFilesToServerJSON(proj, w, func() map[string]interface{} {
		return request.WritePreview(projectID, sandbox)
	}); err != nil {
		return err
	}
	log.Outf("Waiting for server to respond. It could take up to 1 minute if your cloud function needs to be redeployed.")
	err = <-errCh
	if err != nil {
		return err
	}
	log.DoneMsgln(fmt.Sprintf("You can now test your changes in Simulator with this URL: %s", simulatorURL))
	return nil
}

func procCreateVersionResponse(channel string, body []byte) (string, error) {
	resp := &CreateVersionHTTPResponse{}
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(resp); err != nil {
		return "", errors.New(string(body))
	}
	versionIDRegExp := regexp.MustCompile("^projects/[^//]+/versions/(?P<versionID>[^//]+)$")
	if versionIDMatch := versionIDRegExp.FindStringSubmatch(resp.Name); versionIDMatch == nil {
		log.Debugln(fmt.Sprintf("version id absent in the response %s returned from the server ", resp.Name))
		return "", nil
	}
	return versionIDRegExp.FindStringSubmatch(resp.Name)[versionIDRegExp.SubexpIndex("versionID")], nil
}

// CreateVersionJSON implements CreateVersion functionality of the SDK server via HTTP/JSON streaming.
func CreateVersionJSON(ctx context.Context, proj project.Project, channel string) error {
	clientSecret, err := proj.ClientSecretJSON()
	if err != nil {
		return err
	}
	client, err := apiutils.NewHTTPClient(ctx, clientSecret, "")
	if err != nil {
		return err
	}
	projectID := proj.ProjectID()
	log.Outf("Deploying files in the project %q to the %q release channel...", projectID, channel)
	requestURL := httpAddr(versionHTTPEndpoint(projectID))
	r, w := io.Pipe()
	errCh := make(chan error, 1)
	var versionID string
	// This goroutine will exit after HTTP call is finished.
	// The sendFilesToServerJSON below and client.Post communicate via the pipe
	// and former will keep writing stream of bytes, which client post will
	// keep reading in a blocking fashion. sendFilesToServerJSON is guaranteed
	// to close the writer end of the pipe, thus unblocking the reader and allowing
	// the goroutine to exit.
	go func() {
		req, err := http.NewRequest("POST", requestURL, r)
		if err != nil {
			errCh <- err
			return
		}
		req.Header.Add("Content-Type", "application/json")
		// This is done to help server select the quota attributed to a
		// projectID (i.e. developer's project), instead of the CLI project.
		// https://cloud.google.com/storage/docs/xml-api/reference-headers#xgooguserproject
		req.Header.Add("X-Goog-User-Project", projectID)
		addClientHeaders(req)

		resp, err := client.Do(req)
		if err != nil {
			errCh <- err
			return
		}
		defer resp.Body.Close()
		// TODO: Change signature of postProcessJSONResponse to return an error, and pipe that error to channel here.
		postprocessJSONResponse(resp, errCh, func(body []byte) error {
			v, err := procCreateVersionResponse(channel, body)
			versionID = v
			return err
		})
	}()
	if err := sendFilesToServerJSON(proj, w, func() map[string]interface{} {
		return request.CreateVersion(projectID, channel)
	}); err != nil {
		return err
	}
	log.Outf("Waiting for server to respond...")
	if err := <-errCh; err != nil {
		return err
	}
	if _, ok := BuiltInReleaseChannels[channel]; ok {
		channel = BuiltInReleaseChannels[channel]
	}

	log.DoneMsgln(fmt.Sprintf("Version %s has been successfully created and submitted for deployment to %s channel. ", versionID, channel))
	return nil
}

func keyInConfigResp(path string) (string, error) {
	var k string
	switch {
	case studio.IsWebhookDefinition(path):
		k = "webhook"
	case studio.IsVertical(path):
		k = "verticalSettings"
	case studio.IsManifest(path):
		k = "manifest"
	case studio.IsActions(path):
		k = "actions"
	case studio.IsIntent(path):
		k = "intent"
	case studio.IsGlobal(path):
		k = "globalIntentEvent"
	case studio.IsScene(path):
		k = "scene"
	case studio.IsType(path):
		k = "type"
	case studio.IsEntitySet(path):
		k = "entitySet"
	case studio.IsPrompt(path):
		k = "staticPrompt"
	case studio.IsDeviceFulfillment(path):
		k = "deviceFulfillment"
	case studio.IsResourceBundle(path):
		k = "resourceBundle"
	case studio.IsSettings(path):
		k = "settings"
	case studio.IsAccountLinkingSecret(path):
		k = "accountLinkingSecret"
	default:
		return k, fmt.Errorf("%v is unknown config file type to CLI", path)
	}
	return k, nil
}

func receiveConfigFiles(proj project.Project, cfgs *configFiles, force bool, seen map[string]bool) error {
	for _, cfg := range cfgs.ConfigFiles {
		p, ok := cfg["filePath"]
		if !ok {
			return fmt.Errorf("%v doesn't have required filePath field", cfg)
		}
		path, ok := p.(string)
		if !ok {
			return fmt.Errorf("%v has a key of %v of incorrect type %T, want string", cfg, p, p)
		}
		k, err := keyInConfigResp(path)
		if err != nil {
			return err
		}
		v := cfg[k]
		// Transform v into YAML.
		mp, ok := v.(map[string]interface{})
		if !ok {
			return fmt.Errorf("%v has a key %v of incorrect type %T", cfg, v, v)
		}
		b, err := yaml.Marshal(mp)
		if err != nil {
			return err
		}
		// TODO: Can be spun as go-routine.
		if err := studio.WriteToDisk(proj, path, "", b, force); err != nil {
			return err
		}
		seen[path] = true
	}
	return nil
}

func receiveDataFiles(proj project.Project, dfs *dataFiles, force bool, seen map[string]bool) error {
	for _, df := range dfs.DataFiles {
		if err := studio.WriteToDisk(proj, df.Filepath, df.ContentType, df.Payload, force); err != nil {
			return err
		}
		if df.ContentType != "application/zip;zip_type=cloud_function" {
			seen[df.Filepath] = true
			continue
		}
		// WriteToDisk will unzip cloud function folder. Need to record the names of extracted files.
		names, err := namesFromZip(df.Payload)
		if err != nil {
			return err
		}
		for _, v := range names {
			fp := path.Join(df.Filepath[:len(df.Filepath)-len(".zip")], v)
			seen[fp] = true
		}
	}
	return nil
}

func receiveStream(proj project.Project, body io.Reader, force bool, seen map[string]bool) error {
	dec := json.NewDecoder(body)
	log.Debugln("Starts processing the stream")
	// Reads "[".
	t, err := dec.Token()
	if err != nil {
		return err
	}
	if t != json.Delim('[') {
		return fmt.Errorf("expected [ got %v", t)
	}
	for dec.More() {
		var rec streamRecord
		if err := dec.Decode(&rec); err != nil {
			return err
		}
		if rec.Files.ConfigFiles != nil {
			if err := receiveConfigFiles(proj, rec.Files.ConfigFiles, force, seen); err != nil {
				return err
			}
		}
		if rec.Files.DataFiles != nil {
			if err := receiveDataFiles(proj, rec.Files.DataFiles, force, seen); err != nil {
				return err
			}
		}
	}
	// Reads "]".
	t, err = dec.Token()
	if err != nil {
		return err
	}
	if t != json.Delim(']') {
		return fmt.Errorf("expected ] got %v", t)
	}
	log.Debugln("Finished processing the stream")
	return nil
}

func namesFromZip(content []byte) ([]string, error) {
	r, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return nil, err
	}
	var names []string
	for _, f := range r.File {
		names = append(names, f.Name)
	}
	return names, nil
}

func findExtra(a map[string][]byte, b map[string]bool) []string {
	u := map[string]bool{}
	for k := range a {
		u[k] = true
	}
	for k := range b {
		u[k] = true
	}
	// flag f in a, that are not in b
	var extra []string
	for k := range u {
		if _, ok := b[k]; !ok {
			extra = append(extra, k)
		}
	}
	return extra
}

func addClientHeaders(req *http.Request) {
	if Consumer != "" {
		req.Header.Add("Gactions-Consumer", Consumer)
	}
	ua := fmt.Sprintf("gactions/%s (%s %s)", version.CliVersion, runtime.GOOS, runtime.GOARCH)
	req.Header.Add("User-Agent", ua)
}

func parseEncryptionKeyVersion(files map[string][]byte) string {
	type secretFile struct {
		EncryptionKeyVersion string `yaml:"encryptionKeyVersion"`
	}
	in, ok := files["settings/accountLinkingSecret.yaml"]
	if !ok {
		return ""
	}
	f := secretFile{}
	if err := yaml.Unmarshal(in, &f); err != nil {
		return ""
	}
	return f.EncryptionKeyVersion
}

// ReadDraftJSON implements ReadDraft functionality of SDK server.
func ReadDraftJSON(ctx context.Context, proj project.Project, force bool, clean bool) error {
	client, err := setupClient(ctx, proj)
	if err != nil {
		return err
	}
	projectID := proj.ProjectID()
	log.Outf("Pulling files in the project %q from Actions Console...\n", projectID)
	requestURL := httpAddr(readDraftHTTPEndpoint(projectID))
	warn := "%v is not present in the draft of your Action"
	files, err := proj.Files()
	if err != nil {
		return err
	}
	body, err := json.Marshal(request.ReadDraft(projectID, parseEncryptionKeyVersion(files)))
	if err != nil {
		return err
	}
	return sendRequest(client, requestURL, body, files, proj, warn, force, clean)
}

func procEncryptSecretResponse(proj project.Project, body []byte) error {
	r := EncryptSecretHTTPResponse{}
	if err := json.Unmarshal(body, &r); err != nil {
		return err
	}
	b, err := yaml.Marshal(r.AccountLinkingSecret)
	if err != nil {
		return err
	}
	if err := studio.WriteToDisk(proj, "settings/accountLinkingSecret.yaml", "", b, false); err != nil {
		return err
	}
	log.DoneMsgln(fmt.Sprintf("Encrypted secret is in %s", filepath.Join(proj.ProjectRoot(), "settings", "accountLinkingSecret.yaml")))
	return nil
}

// EncryptSecretJSON implements Encrypt functionality of SDK server.
func EncryptSecretJSON(ctx context.Context, proj project.Project, secret string) error {
	clientSecret, err := proj.ClientSecretJSON()
	if err != nil {
		return err
	}
	client, err := apiutils.NewHTTPClient(ctx, clientSecret, "")
	if err != nil {
		return err
	}
	log.Outf("Encrypting your client secret...")
	// Using a channel and goroutine is not ideal here, but this allows one to
	// reuse postprocessJSONResponse function.
	// Should to refactor postprocessJSONResponse to avoid channels.
	errCh := make(chan error, 1)
	go func() {
		requestURL := httpAddr(encryptEndpoint)
		body, err := json.Marshal(request.EncryptSecret(secret))
		if err != nil {
			errCh <- err
		}
		req, err := http.NewRequest("POST", requestURL, bytes.NewReader(body))
		if err != nil {
			errCh <- err
		}
		req.Header.Add("Content-Type", "application/json")
		addClientHeaders(req)
		resp, err := client.Do(req)
		if err != nil {
			errCh <- err
		}
		defer resp.Body.Close()
		postprocessJSONResponse(resp, errCh, func(body []byte) error {
			return procEncryptSecretResponse(proj, body)
		})
	}()
	if err := <-errCh; err != nil {
		return err
	}
	return nil
}

func procDecryptSecretResponse(proj project.Project, body []byte, out string) error {
	type resp struct {
		ClientSecret string `json:"clientSecret"`
	}
	r := resp{}
	if err := json.Unmarshal(body, &r); err != nil {
		return err
	}
	rel, err := filepath.Rel(proj.ProjectRoot(), out)
	if err != nil {
		return err
	}
	if err := studio.WriteToDisk(proj, rel, "", []byte(r.ClientSecret), false); err != nil {
		return err
	}
	log.Warnf("Decrypted key will be stored at %s. Committing this file to source control is not recommend.\n", out)
	log.DoneMsgln(fmt.Sprintf("Decrypted client secret key is in %s.", out))
	return nil
}

// DecryptSecretJSON implements Decrypt functionality of SDK server.
func DecryptSecretJSON(ctx context.Context, proj project.Project, secret string, out string) error {
	clientSecret, err := proj.ClientSecretJSON()
	if err != nil {
		return err
	}
	client, err := apiutils.NewHTTPClient(ctx, clientSecret, "")
	if err != nil {
		return err
	}
	log.Outf("Decrypting your client secret...")
	requestURL := httpAddr(decryptEndpoint)
	body, err := json.Marshal(request.DecryptSecret(secret))
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", requestURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	addClientHeaders(req)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Using a channel and goroutine is not ideal here, but this allows one to
	// reuse postprocessJSONResponse function.
	// Should to refactor postprocessJSONResponse to avoid channels.
	errCh := make(chan error, 1)
	postprocessJSONResponse(resp, errCh, func(body []byte) error {
		return procDecryptSecretResponse(proj, body, out)
	})
	return <-errCh
}

func sendListRequest(pageToken, requestURL string, client *http.Client) ([]byte, error) {
	// List API must not have a body, so encoding request fields into a URL.
	u, err := url.Parse(requestURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("pageToken", pageToken)
	u.RawQuery = q.Encode()
	requestURL = u.String()
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}
	addClientHeaders(req)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, parseError(body)
	}
	return body, nil
}

// ListSampleProjectsJSON implements ListSampleProjects endpoint of SDK server.
func ListSampleProjectsJSON(ctx context.Context, proj project.Project) ([]project.SampleProject, error) {
	clientSecret, err := proj.ClientSecretJSON()
	if err != nil {
		return nil, err
	}
	client, err := apiutils.NewHTTPClient(ctx, clientSecret, "")
	if err != nil {
		return nil, err
	}
	requestURL := httpAddr(listSampleProjectsEndpoint)
	var res []project.SampleProject
	pageToken := ""

	for {
		body, err := sendListRequest(pageToken, requestURL, client)
		if err != nil {
			return nil, err
		}
		type listSampleProjectsResponse struct {
			SampleProjects []project.SampleProject `json:"sampleProjects"`
			NextPageToken  string                  `json:"nextPageToken"`
		}
		r := listSampleProjectsResponse{}
		if err = json.Unmarshal(body, &r); err != nil {
			return nil, err
		}
		pageToken = r.NextPageToken
		for _, v := range r.SampleProjects {
			// API returns sampleProjects/{sampleName}.
			v.Name = strings.TrimPrefix(v.Name, "sampleProjects/")
			res = append(res, v)
		}
		if pageToken == "" {
			break
		}
	}
	return res, nil
}

// ReadVersionJSON implements ReadVersion functionality of SDK server.
func ReadVersionJSON(ctx context.Context, proj project.Project, force bool, clean bool, versionID string) error {
	client, err := setupClient(ctx, proj)
	if err != nil {
		return err
	}

	projectID := proj.ProjectID()
	log.Outf("Pulling version %q of the project %q from Actions Console...\n", versionID, projectID)
	requestURL := httpAddr(readVersionHTTPEndpoint(projectID, versionID))
	warning := "%v is not present in the version of your Action"

	files, err := proj.Files()
	if err != nil {
		return err
	}
	body, err := json.Marshal(request.ReadVersion(projectID, parseEncryptionKeyVersion(files)))
	if err != nil {
		return err
	}

	return sendRequest(client, requestURL, body, files, proj, warning, force, clean)
}

func setupClient(ctx context.Context, proj project.Project) (*http.Client, error) {
	clientSecret, err := proj.ClientSecretJSON()
	if err != nil {
		return nil, err
	}
	client, err := apiutils.NewHTTPClient(ctx, clientSecret, "")
	if err != nil {
		return nil, err
	}
	return client, nil
}

func sendRequest(client *http.Client, requestURL string, body []byte, files map[string][]byte, proj project.Project, warning string, force, clean bool) error {
	projectID := proj.ProjectID()

	req, err := http.NewRequest("POST", requestURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	// This is done to help server select the quota attributed to a
	// projectID (i.e. developer's project), instead of the CLI project.
	// https://cloud.google.com/storage/docs/xml-api/reference-headers#xgooguserproject
	req.Header.Add("X-Goog-User-Project", projectID)
	addClientHeaders(req)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		// In case of an error, it's okay to read entire response body because
		// it will be small.
		body, err := readBodyWithTimeout(resp.Body, responseBodyReadTimeout)
		if err != nil {
			return err
		}
		log.Debugln(string(body))
		publicErrors := []PublicError{}
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&publicErrors); err != nil {
			// This means the error is not a JSON. This happens when the API URL is malformed, and
			// one platform returns an HTML response. In this case, we print the HTML and disregard the json decoding error.
			return fmt.Errorf(string(body))
		}
		if len(publicErrors) > 0 {
			return fmt.Errorf("server did not return HTTP 200\n%v", errorMessage(&publicErrors[0]))
		}
		return errors.New("server did not return HTTP 200")
	}
	seen := map[string]bool{}
	if err := receiveStream(proj, resp.Body, force, seen); err != nil {
		return err
	}
	extra := findExtra(files, seen)
	for _, v := range extra {
		fp := filepath.Join(proj.ProjectRoot(), filepath.FromSlash(v))
		warn := fmt.Sprintf(warning, fp)
		if clean {
			log.Warnf("%v. Removing %v.\n", warn, fp)
			if err := os.RemoveAll(fp); err != nil {
				return err
			}
		} else {
			log.Warnf("%v. To remove, run pull with --clean flag.\n", warn)
		}
	}
	return nil
}

// ListReleaseChannelsJSON implements ListReleaseChannels endpoint of SDK server.
func ListReleaseChannelsJSON(ctx context.Context, proj project.Project) ([]project.ReleaseChannel, error) {
	clientSecret, err := proj.ClientSecretJSON()
	if err != nil {
		return nil, err
	}
	client, err := apiutils.NewHTTPClient(ctx, clientSecret, "")
	if err != nil {
		return nil, err
	}
	requestURL := httpAddr(listReleaseChannelsHTTPEndpoint(proj.ProjectID()))
	var res []project.ReleaseChannel
	pageToken := ""

	for {
		body, err := sendListRequest(pageToken, requestURL, client)
		if err != nil {
			return nil, err
		}
		type listReleaseChannelsResponse struct {
			ReleaseChannels []project.ReleaseChannel `json:"releaseChannels"`
			NextPageToken   string                   `json:"nextPageToken"`
		}
		r := listReleaseChannelsResponse{}
		if err = json.Unmarshal(body, &r); err != nil {
			return nil, err
		}
		pageToken = r.NextPageToken
		for _, v := range r.ReleaseChannels {
			// API returns releaseChannels/{releaseChannelName}.
			v.Name = strings.TrimPrefix(v.Name, "releaseChannels/")
			res = append(res, v)
		}
		if pageToken == "" {
			break
		}
	}
	return res, nil
}

// ListVersionsJSON implements ListVersions endpoint of SDK server.
func ListVersionsJSON(ctx context.Context, proj project.Project) ([]project.Version, error) {
	clientSecret, err := proj.ClientSecretJSON()
	if err != nil {
		return nil, err
	}
	client, err := apiutils.NewHTTPClient(ctx, clientSecret, "")
	if err != nil {
		return nil, err
	}
	requestURL := httpAddr(listVersionsHTTPEndpoint(proj.ProjectID()))
	var res []project.Version
	pageToken := ""

	for {
		body, err := sendListRequest(pageToken, requestURL, client)
		if err != nil {
			return nil, err
		}
		type listVersionsResponse struct {
			Versions      []project.Version `json:"versions"`
			NextPageToken string            `json:"nextPageToken"`
		}
		r := listVersionsResponse{}
		if err := json.Unmarshal(body, &r); err != nil {
			return nil, err
		}
		pageToken = r.NextPageToken
		for _, v := range r.Versions {
			// API returns versions/{versionName}.
			v.ID = strings.TrimPrefix(v.ID, "versions/")
			res = append(res, v)
		}
		if pageToken == "" {
			break
		}
	}
	return res, nil
}
