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

package ginit

import (
	"bytes"
	"context"
	"testing"

	"github.com/actions-on-google/gactions/project"
	"github.com/spf13/cobra"
)

type MockStudio struct {
}

func (MockStudio) ProjectID() string {
	return ""
}

func (p MockStudio) Download(sample project.SampleProject, dest string) error {
	return nil
}

// studio project.
func (p MockStudio) AlreadySetup(pathToWorkDir string) bool {
	return false
}

func (p MockStudio) Files() (map[string][]byte, error) {
	return map[string][]byte{}, nil
}

func (p MockStudio) ClientSecretJSON() ([]byte, error) {
	return []byte{}, nil
}

func (p MockStudio) ProjectRoot() string {
	return ""
}

func TestCmdExecute(cmd *cobra.Command, args []string) (string, error) {
	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return output.String(), err
}

func execute(args ...string) (string, error) {
	cmd := &cobra.Command{}
	project := MockStudio{}
	AddCommand(context.Background(), cmd, project)
	return TestCmdExecute(cmd, args)
}

func TestInitWithInvalidArgs(t *testing.T) {
	og := availableProjects
	availableProjects = func(ctx context.Context, p project.Project) ([]project.SampleProject, error) {
		return []project.SampleProject{
			project.SampleProject{"question", "https://google.com"},
		}, nil
	}
	defer func() {
		availableProjects = og
	}()
	tests := []struct {
		invalidArgs []string
	}{
		{
			invalidArgs: []string{"init"},
		},
		{
			invalidArgs: []string{"init", "foo"},
		},
	}
	for _, tc := range tests {
		if _, err := execute(tc.invalidArgs...); err == nil {
			t.Errorf("init didn't fail when %v were passed", tc.invalidArgs)
		}
	}
}

func TestInitWithValidArgs(t *testing.T) {
	og := availableProjects
	availableProjects = func(ctx context.Context, p project.Project) ([]project.SampleProject, error) {
		return []project.SampleProject{
			project.SampleProject{"question", "https://google.com"},
		}, nil
	}
	defer func() {
		availableProjects = og
	}()
	tests := []struct {
		validArgs []string
	}{
		{
			validArgs: []string{"init", "question"},
		},
		{
			validArgs: []string{"init", "question", "--dest", "~/Code/"},
		},
	}

	for _, tc := range tests {
		if _, err := execute(tc.validArgs...); err != nil {
			t.Errorf("init failed when %v were passed", tc.validArgs)
		}
	}
}
