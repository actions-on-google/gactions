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

// Package project contains an interface for an AoG project.
package project

const (
	// ConfigName is filename of the file containing CLIConfig.
	ConfigName = ".gactionsrc.yaml"
)

// CLIConfig represents a config file for CLI to read parameters from.
type CLIConfig struct {
	SdkPath string `yaml:"sdkPath"`
}

// SampleProject has information about sample projects that CLI supports.
type SampleProject struct {
	Name      string `json:"name"`
	HostedURL string `json:"hostedUrl"`
}

// ReleaseChannel has information about release channels for the project
// and their current and pending versions.
type ReleaseChannel struct {
	Name           string `json:"name"`
	CurrentVersion string `json:"currentVersion"`
	PendingVersion string `json:"pendingVersion"`
}

// VersionState has information about state of the version.
type VersionState struct {
	Message string `json:"message"`
}

// Version has information about versions and their metadata for a project.
type Version struct {
	ID             string       `json:"name"`
	State          VersionState `json:"versionState"`
	LastModifiedBy string       `json:"creator"`
	ModifiedOn     string       `json:"updateTime"`
}

// Project represents the concept of an AoG project.
// The concrete implementations will include existing types of projects (i.e. Studio)
// This is used by the CLI for various commands.
type Project interface {
	// Download places the files from sample project into dest. Returns an error if any.
	// URL must be a URL pointing to a Git repository.
	Download(sample SampleProject, dest string) error
	// AlreadySetup returns true if pathToWorkDir already contains a complete
	// studio project.
	AlreadySetup(pathToWorkDir string) bool
	// Files returns project files as a (filename string, content []byte) pair, where
	// filename is a relative path starting from the root of the project.
	Files() (map[string][]byte, error)
	// ClientSecretJSON returns a client secret used to communicate with an external API.
	ClientSecretJSON() ([]byte, error)
	// ProjectRoot returns a root directory of a project. If root directory is not found,
	// the returned string will be empty (i.e. ""). Otherwise, the path will be an
	// OS-native path (i.e. separated with "\\" for Windows or "/" for Mac and Linux)
	ProjectRoot() string
	// ProjectID returns a Google Project ID associated with developer's Action, which should be safe to insert into the URL.
	ProjectID() string
}
