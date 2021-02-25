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

package push

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/actions-on-google/gactions/project"
	"github.com/actions-on-google/gactions/project/studio"
	"github.com/spf13/cobra"
)

func TestCmdExecute(cmd *cobra.Command, args []string) (string, error) {
	output := new(bytes.Buffer)
	cmd.SetOutput(output)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return output.String(), err
}

func execute(args ...string) (string, error) {
	cmd := &cobra.Command{}
	project := studio.New([]byte{}, ".")
	AddCommand(context.Background(), cmd, project)
	return TestCmdExecute(cmd, args)
}

func TestPush(t *testing.T) {
	// TODO: Need to setup a settings.yaml file so command can read it to get projectID.
	t.Skip()
	originalDoPush := doPush
	defer func() {
		doPush = originalDoPush
	}()
	doPush = func(ctx context.Context, cmd *cobra.Command, args []string, proj project.Project) error {
		if proj == nil {
			return fmt.Errorf("proj is %v, want not nil", proj)
		}
		return nil
	}
	if _, err := execute("push"); err != nil {
		t.Errorf("push failed and returned %v, want %v", err.Error(), nil)
	}
}
