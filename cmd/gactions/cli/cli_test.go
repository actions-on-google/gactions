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

package cli

import (
	"context"
	"testing"

	"github.com/actions-on-google/gactions/api/sdk"
	"github.com/actions-on-google/gactions/log/log"
	"github.com/spf13/cobra"
)

func TestCommandDebugSet(t *testing.T) {
	old := log.Severity
	defer func() {
		log.Severity = old
	}()
	cmd := Command(context.Background(), "gactions", true, "")
	// CLI sets logging at runtime, so need to simulate execution
	cmd.RunE = func(*cobra.Command, []string) error {
		return nil
	}
	_ = Execute(cmd)
	if log.Severity != log.DebugLevel {
		t.Errorf("Command set severity to %v, but want %v", log.Severity, log.DebugLevel)
	}
}

func TestCommandDebugNotSet(t *testing.T) {
	old := log.Severity
	defer func() {
		log.Severity = old
	}()
	cmd := Command(context.Background(), "gactions", false, "")
	// CLI sets logging at runtime, so need to simulate execution
	cmd.RunE = func(*cobra.Command, []string) error {
		return nil
	}
	_ = Execute(cmd)
	if log.Severity != log.WarnLevel {
		t.Errorf("Command set severity to %v, but want %v", log.Severity, log.WarnLevel)
	}
	// check debug flags
	debugFlags := []string{"--env=foo", "--cookie=abc"}
	for _, v := range debugFlags {
		cmd.SetArgs([]string{v})
		code := Execute(cmd)
		if code != 1 {
			t.Errorf("Executed returned %v, but want %v when %v flag is set.", code, 1, v)
		}
	}
}

func TestCommandEnvFlagDebugSet(t *testing.T) {
	old := sdk.CurEnv
	defer func() {
		sdk.CurEnv = old
	}()
	cmd := Command(context.Background(), "gactions", true, "")
	// CLI sets logging at runtime, so need to simulate execution
	cmd.RunE = func(*cobra.Command, []string) error {
		return nil
	}
	// case 1
	cmd.SetArgs([]string{"--env=prod"})
	code := Execute(cmd)
	if code != 0 {
		t.Errorf("Execute returned %v, but want %v", code, 0)
	}
	if sdk.CurEnv != "prod" {
		t.Errorf("Expected to set CurEnv to %v, but got %v", "prod", sdk.CurEnv)
	}
	// case 2
	cmd.SetArgs([]string{"--env=foo"})
	code = Execute(cmd)
	if code != 1 {
		t.Errorf("Executed returned %v, but want %v", code, 1)
	}
	if sdk.CurEnv != "prod" {
		t.Errorf("Expected CurEnv to remain %v, but got %v", "prod", sdk.CurEnv)
	}
}
