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

// Package push provides an implementation of "gactions push" command.
package push

import (
	"context"
	"errors"
	"fmt"

	"github.com/actions-on-google/gactions/api/sdk"
	"github.com/actions-on-google/gactions/log"
	"github.com/actions-on-google/gactions/project"
	"github.com/actions-on-google/gactions/project/studio"
	"github.com/spf13/cobra"
)

// AddCommand adds the push sub-command to the passed in root command.
func AddCommand(ctx context.Context, root *cobra.Command, proj project.Project) {
	push := &cobra.Command{
		Use:   "push",
		Short: "This command pushes changes in the local files to Actions Console.",
		Long:  "This command pushes changes in the local files to Actions Console.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if proj.ProjectRoot() == "" {
				log.Errorf(`Can't find a project root. This may be because (1) %q was not found in this or any of the parent folders, or (2) if %q was found, but the key "sdkPath" was missing, or (3) if %q and manifest.yaml were both not found.`, project.ConfigName, project.ConfigName, project.ConfigName)
				return errors.New("can not determine project root")
			}
			studioProj, ok := proj.(studio.Studio)
			if !ok {
				return fmt.Errorf("can not convert %T to %T", proj, studio.Studio{})
			}
			if err := (&studioProj).SetProjectID(""); err != nil {
				return err
			}
			return doPush(ctx, cmd, args, studioProj)
		},
		Args: cobra.NoArgs,
	}
	root.AddCommand(push)
}

var doPush = func(ctx context.Context, cmd *cobra.Command, args []string, proj project.Project) error {
	return sdk.WriteDraftJSON(ctx, proj)
}
