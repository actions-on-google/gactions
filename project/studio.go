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

// Package studio contains a Studio implementation of a project.Project interface.
package studio

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/actions-on-google/gactions/api/yamlutils"
	"github.com/actions-on-google/gactions/log"
	"github.com/actions-on-google/gactions/project"
)

// Studio is an implementation of the AoG Studio project.
type Studio struct {
	files            map[string][]byte
	clientSecretJSON []byte
	root             string
	projectID        string
}

// New returns a new instance of Studio.
// Note(atulep): Defined this here to allow testing (otherwise was getting build errors)
func New(secret []byte, projectRoot string) Studio {
	return Studio{clientSecretJSON: secret, root: projectRoot}
}

// Download places the files from sample project into dest. Returns an error if any.
func (p Studio) Download(sample project.SampleProject, dest string) error {
	return downloadFromGit(sample.Name, sample.HostedURL, dest)
}

func downloadFromGit(projectTitle, url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("can not download from %v", url)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return unzipZippedDir(dest, b)
}

func unzipZippedDir(dest string, content []byte) error {
	// Open a zip archive for reading.
	r, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dest, 0750); err != nil {
		return err
	}
	// The shortest name will be directory name that was unzipped.
	sort.Slice(r.File, func(i, j int) bool {
		return r.File[i].Name < r.File[j].Name
	})
	dir := filepath.Join(filepath.FromSlash(dest), r.File[0].Name)
	log.Infof("Unzipping %v", dir)
	for _, f := range r.File[1:] {
		fp, err := filepath.Rel(r.File[0].Name, f.Name)
		if err != nil {
			return err
		}
		fp = filepath.Join(dest, fp)
		fp = filepath.FromSlash(fp)

		if f.Mode().IsDir() {
			if err := os.MkdirAll(fp, 0750); err != nil {
				return err
			}
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		b, err := ioutil.ReadAll(rc)
		if err != nil {
			return err
		}
		log.Infof("Writing %v\n", fp)
		if err := ioutil.WriteFile(fp, b, 0640); err != nil {
			return err
		}
		if err := rc.Close(); err != nil {
			return err
		}
	}
	return nil
}

// isLocalizedSettings returns whether a file named filename is a
// localized settings file. An example of localized settings is
// "settings/zh-TW/settings.yaml", and example of non-localized settings is
// "settings/settings.yaml", where "zh-TW" represents a locale.
func isLocalizedSettings(filename string) bool {
	// This is a heuristic that checks if the parent directory of
	// the filename is not "settings", which means it's probably a locale.
	subpaths := strings.Split(filename, string(os.PathSeparator))
	if len(subpaths) < 2 {
		return false
	}
	secondToLast := subpaths[len(subpaths)-2]
	return secondToLast != "settings"
}

func isConfigFile(filename string) bool {
	return IsVertical(filename) ||
		IsManifest(filename) ||
		IsSettings(filename) ||
		IsActions(filename) ||
		IsIntent(filename) ||
		IsGlobal(filename) ||
		IsScene(filename) ||
		IsType(filename) ||
		IsWebhookDefinition(filename) ||
		IsResourceBundle(filename) ||
		IsPrompt(filename) ||
		IsAccountLinkingSecret(filename)
}

// IsWebhookDefinition reteurns true if the file contains a  yaml definition of the webhook.
func IsWebhookDefinition(filename string) bool {
	return IsWebhook(filename) && path.Ext(filename) == ".yaml"
}

// IsVertical returns true if the file contains vertical config files.
func IsVertical(filename string) bool {
	return strings.HasPrefix(filename, "verticals") && path.Ext(filename) == ".yaml"
}

// IsManifest returns true if the file contains a manifest of an Actions project.
func IsManifest(filename string) bool {
	return path.Base(filename) == "manifest.yaml"
}

// IsSettings returns true if the file contains settings of an Actions project.
func IsSettings(filename string) bool {
	return path.Base(filename) == "settings.yaml"
}

// IsActions returns true if the file contains an Action declaration of an Actions project.
func IsActions(filename string) bool {
	return path.Base(filename) == "actions.yaml"
}

// IsIntent returns true if the file contains an intent definition of an Actions project.
func IsIntent(filename string) bool {
	return strings.HasPrefix(filename, path.Join("custom", "intents")) && path.Ext(filename) == ".yaml"
}

// IsGlobal returns true if the file contains a global scene interaction declaration
// of an Actions project.
func IsGlobal(filename string) bool {
	return strings.HasPrefix(filename, path.Join("custom", "global")) && path.Ext(filename) == ".yaml"
}

// IsScene returns true if the file contains a scene declaration of an Actions project.
func IsScene(filename string) bool {
	return strings.HasPrefix(filename, path.Join("custom", "scenes")) && path.Ext(filename) == ".yaml"
}

// IsType returns true if the file contains a type declaration of an Actions project.
func IsType(filename string) bool {
	return strings.HasPrefix(filename, path.Join("custom", "types")) && path.Ext(filename) == ".yaml"
}

// IsWebhook returns true if the file contains a webhook files of an Actions project.
// This includes yaml and code files.
func IsWebhook(filename string) bool {
	return strings.HasPrefix(filename, path.Join("webhooks"))
}

// IsPrompt returns true if the file contains a prompt of an Actions project.
func IsPrompt(filename string) bool {
	return strings.HasPrefix(filename, path.Join("custom", "prompts")) && path.Ext(filename) == ".yaml"
}

// IsResourceBundle returns true if the file contains a resource bundle. This will return true if
// filename for either localized or base resource bundle.
func IsResourceBundle(filename string) bool {
	return strings.HasPrefix(filename, path.Join("resources", "strings")) && path.Ext(filename) == ".yaml"
}

// IsAccountLinkingSecret returns true if the file contains an account linking secret. The file
// must have the name settings/accountLinkingSecret.yaml.
func IsAccountLinkingSecret(filename string) bool {
	return strings.HasPrefix(filename, path.Join("settings", "accountLinkingSecret.yaml"))
}

// ConfigFiles finds configuration files from the files of a project.
func ConfigFiles(files map[string][]byte) map[string][]byte {
	configFiles := map[string][]byte{}
	for k, v := range files {
		if isConfigFile(k) {
			configFiles[k] = v
		}
	}
	return configFiles
}

var askYesNo = func(msg string) (string, error) {
	log.Outf("%v. [y/n]", msg)
	var ans string
	_, err := fmt.Scan(&ans)
	if err != nil {
		return "", err
	}
	norm := strings.ToLower(ans)
	if norm == "y" || norm == "yes" {
		return "yes", nil
	}
	if norm == "n" || norm == "no" {
		return "no", nil
	}
	return "", fmt.Errorf("invalid option specified: %v", ans)
}

// WriteToDisk writes content into path located in local file system. Path is relative
// to project root (i.e. same level as manifest.yaml). This function will appropriately
// combine value of path with project root to write the file in an appropriate location.
// ContentType needs to be non-empty for data files; config files can have an empty string.
func WriteToDisk(proj project.Project, path string, contentType string, payload []byte, force bool) error {
	path = filepath.FromSlash(path)
	if proj.ProjectRoot() != "" {
		path = filepath.Join(proj.ProjectRoot(), path)
	}
	if contentType == "application/zip;zip_type=cloud_function" {
		path = path[:len(path)-len(".zip")]
	}
	if exists(path) {
		var ans string
		if !force {
			r, err := askYesNo(fmt.Sprintf("%v already exists. Would you like to overwrite it?", path))
			if err != nil {
				return err
			}
			ans = r
		}
		if ans == "yes" || force {
			log.Infof("Removing %v\n", path)
			if err := os.RemoveAll(path); err != nil {
				return err
			}
		} else {
			log.Infof("Skipping %v\n", path)
			return nil
		}
	}
	// proj.ProjectRoot() already exists, but old value of path may have project-specific subdirs that need to be created.
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return err
	}
	if contentType == "application/zip;zip_type=cloud_function" {
		return unzipFiles(path, payload)
	}
	log.Infof("Writing %v\n", path)
	return ioutil.WriteFile(path, payload, 0640)
}

func unzipFiles(dir string, content []byte) error {
	// Open a zip archive for reading.
	r, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return err
	}
	for _, f := range r.File {
		fp := filepath.Join(dir, f.Name)
		fp = filepath.FromSlash(fp)
		rc, err := f.Open()
		if err != nil {
			return err
		}
		b, err := ioutil.ReadAll(rc)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(fp), 0750); err != nil {
			return err
		}
		log.Infof("Writing %v\n", fp)
		if err := ioutil.WriteFile(fp, b, 0640); err != nil {
			return err
		}
		rc.Close()
	}
	return nil
}

func zipFiles(files map[string][]byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	for name, content := range files {
		// Server expects Cloud Functions to have the filePath stripped
		// (i.e. webhooks/myfunction/index.js -> ./index.js)
		f, err := w.Create(path.Base(name))
		if err != nil {
			return nil, err
		}
		_, err = f.Write(content)
		if err != nil {
			return nil, err
		}
	}
	err := w.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// addInlineWebhooks adds a zipped inline webhook code, if any, to dataFiles.
func addInlineWebhooks(dataFiles map[string][]byte, files map[string][]byte, root string) error {
	yamls := map[string][]byte{}
	// "code" includes all of the code files under the webhooks directory.
	// This includes both external and inline cloud functions. It will be
	// be used to include inline cloud functions later in the function.
	code := map[string][]byte{}
	for k, v := range files {
		if IsWebhook(k) {
			if IsWebhookDefinition(k) {
				yamls[k] = v
			} else {
				code[k] = v
			}
		}
	}
	for k, v := range yamls {
		mp, err := yamlutils.UnmarshalYAMLToMap(v)
		if err != nil {
			return fmt.Errorf("%v has incorrect syntax: %v", filepath.Join(root, k), err)
		}
		if _, ok := mp["inlineCloudFunction"]; ok {
			filesToZip := map[string][]byte{}
			// Name of the file must match the name of the folder hosting the code for the inline function
			// For example, "webhooks/a.yaml" means "webhooks/a/*" must exist.
			basename := path.Base(k)
			name := basename[:len(basename)-len(path.Ext(basename))]
			funcFolder := path.Join("webhooks", name)
			for k2, v2 := range code {
				// Inline cloud function should just have index.js and package.json
				if strings.HasPrefix(k2, funcFolder) && !strings.Contains(k2, "node_modules") && (path.Ext(k2) == ".js" || path.Ext(k2) == ".json") {
					filesToZip[k2] = v2
				}
			}
			if len(filesToZip) == 0 {
				return fmt.Errorf("folder for inline cloud function is not found for %v", k)
			}
			content, err := zipFiles(filesToZip)
			if err != nil {
				return err
			}
			dataFiles[funcFolder+".zip"] = content
		} else {
			log.Debugf("Found external cloud function: %v\n", filepath.Join(root, k))
		}
	}
	return nil
}

// DataFiles finds data files from the files of a project.
func DataFiles(p project.Project) (map[string][]byte, error) {
	dataFiles := map[string][]byte{}
	files, err := p.Files()
	if err != nil {
		return nil, err
	}
	for k, v := range files {
		if strings.HasPrefix(k, "resources/") && !IsResourceBundle(k) {
			dataFiles[k] = v
		}
	}
	if err := addInlineWebhooks(dataFiles, files, p.ProjectRoot()); err != nil {
		return nil, err
	}
	return dataFiles, nil
}

// ProjectID finds a project id of a project.
func ProjectID(proj project.Project) (string, error) {
	// Note: `k` may have some parent subpath that is hard to predict, so
	// forced to iterate through keys instead of indexing directly.
	files, err := proj.Files()
	if err != nil {
		return "", err
	}
	for k, v := range files {
		if path.Base(k) == "settings.yaml" && !isLocalizedSettings(k) {
			mp, err := yamlutils.UnmarshalYAMLToMap(v)
			if err != nil {
				return "", fmt.Errorf("%v has incorrect syntax: %v", k, err)
			}
			if pid, present := mp["projectId"]; present {
				if pid == "placeholder_project" {
					log.Warnf("%v is not a valid project id. Update settings/settings.yaml file with your Google project id found in your GCP console. E.g. \"123456789\"", pid)
				}
				spid, ok := pid.(string)
				if !ok {
					return "", fmt.Errorf("invalid project ID: %v", pid)
				}
				return spid, nil
			}
			return "", errors.New("projectId is not present in the settings file")
		}
	}
	return "", errors.New("can't find a project id: settings.yaml not found")
}

// AlreadySetup returns true if pathToWorkDir already contains a complete
// studio project.
func (p Studio) AlreadySetup(pathToWorkDir string) bool {
	// Note: This will return true when pathToWorkDir contains
	// hidden directories, such .git
	return exists(pathToWorkDir) && !isDirEmpty(pathToWorkDir)
}

// exists returns whether the given file or directory exists or not
func exists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return os.IsExist(err)
	}
	return true
}

// isDirEmpty returns true if the given directory is empty, otherwise false.
func isDirEmpty(dir string) bool {
	l, err := ioutil.ReadDir(dir)
	if err != nil {
		return false
	}
	var norm []os.FileInfo
	// Skip hidden files and directories, such as .git.
	for _, v := range l {
		if !strings.HasPrefix(v.Name(), ".") {
			norm = append(norm, v)
		}
	}
	return len(norm) <= 0
}

// winToUnix converts path from win to unix
func winToUnix(path string) string {
	return strings.Replace(path, "\\", "/", -1)
}

// ProjectRoot returns a root directory of a project. If root directory is not found, the
// returned string will be empty (i.e. "")
func (p Studio) ProjectRoot() string {
	return p.root
}

func isHidden(path string) bool {
	slashed := filepath.ToSlash(path)
	parts := strings.Split(slashed, "/")
	for _, v := range parts {
		if strings.HasPrefix(v, ".") {
			return true
		}
	}
	return false
}

// Files returns project files as a (filename string, content []byte) pair.
func (p Studio) Files() (map[string][]byte, error) {
	if p.files != nil {
		return p.files, nil
	}
	var m = make(map[string][]byte)
	err := filepath.Walk(p.ProjectRoot(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := relativePath(p.ProjectRoot(), path)
		if err != nil {
			return err
		}
		if !info.IsDir() && !isHidden(relPath) {
			// SDK server expects filepath to be separated using a '/'.
			if runtime.GOOS == "windows" {
				m[winToUnix(relPath)], err = ioutil.ReadFile(path)
			} else {
				// Do not convert a Unix path because it may have a mix of \\ and / in the path
				// as Linux allows it (i.e. mkdir hello\\world is valid on Linux)
				m[relPath], err = ioutil.ReadFile(path)
			}
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	p.files = m
	return m, nil
}

// ClientSecretJSON returns a client secret used to communicate with an external API.
func (p Studio) ClientSecretJSON() ([]byte, error) {
	return p.clientSecretJSON, nil
}

// ProjectID returns a Google Project ID associated with developer's Action, which should be safe to insert into the URL.
func (p Studio) ProjectID() string {
	return url.PathEscape(p.projectID)
}

// SetProjectID sets projectID for studio. It can come from two possible places:
// settings.yaml or command line flag. In case both are specified, the CLI
// will give preference to command line flag, but will provide a warning to a developer.
func (p *Studio) SetProjectID(flag string) error {
	if p.ProjectID() != "" {
		return errors.New("can not reset the project ID")
	}
	pid, err := pidFromSettings(p.ProjectRoot())
	if err != nil && flag == "" {
		log.Errorf(`Project ID is missing. Specify the project ID in settings/settings.yaml, or via flag, if applicable.`)
		return errors.New("no project ID is specified")
	}
	if err == nil && flag != "" && flag != pid {
		log.Warnf("Two Google Project IDs are specified: %q via the flag, %q via the settings file. %q takes a priority and will be used in the remainder of the command.", flag, pid, flag)
		p.projectID = flag
	} else if pid != "" {
		p.projectID = pid
	} else {
		p.projectID = flag
	}
	log.Infof("Using %q.\n", p.ProjectID())
	return nil
}

// SetProjectRoot sets project a root for studio project. It should only be called
// if project root doesn't yet exist, but will be created as a result of a subroutine
// that called SetProjectRoot. In this case, project root will become current working directory.
func (p *Studio) SetProjectRoot() error {
	if p.root != "" {
		return errors.New("can not reset project root")
	}
	r, err := FindProjectRoot()
	if err != nil {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		p.root = wd
		return nil
	}
	p.root = r
	return nil
}

// FindProjectRoot locates the root of the SDK project.
func FindProjectRoot() (string, error) {
	cur, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for !exists(filepath.Join(cur, "manifest.yaml")) {
		parent := filepath.Dir(cur)
		if parent == cur {
			return cur, errors.New("manifest.yaml was not found")
		}
		cur = parent
	}
	return cur, nil
}

func pidFromSettings(root string) (string, error) {
	fp := filepath.Join(root, "settings", "settings.yaml")
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		return "", err
	}
	mp, err := yamlutils.UnmarshalYAMLToMap(b)
	if err != nil {
		return "", fmt.Errorf("%v has incorrect syntax: %v", fp, err)
	}
	type settings struct {
		ProjectID string `json:"projectId"`
	}
	b, err = json.Marshal(mp)
	if err != nil {
		return "", err
	}
	set := settings{}
	if err := json.Unmarshal(b, &set); err != nil {
		return "", err
	}
	if set.ProjectID == "" {
		return "", errors.New("projectId is not present in the settings file")
	}
	if set.ProjectID == "placeholder_project" {
		log.Warnf("%v is not a valid project id. Update %v file with your Google project id found in your GCP console. E.g. \"123456789\"", set.ProjectID, fp)
	}
	return set.ProjectID, nil
}

func relativePath(root, path string) (string, error) {
	// root has OS specific separators, but path does not.
	platSpecific := filepath.FromSlash(path)
	return filepath.Rel(root, platSpecific)
}
