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

// Package pull provides an implementation of "gactions pull" command.
package pull

import (
	"context"
	"fmt"
	"net/url"

	"github.com/actions-on-google/gactions/api/sdk"
	"github.com/actions-on-google/gactions/log"
	"github.com/actions-on-google/gactions/project"
	"github.com/actions-on-google/gactions/project/studio"
	"github.com/spf13/cobra"
)

// AddCommand adds the push sub-command to the passed in root command.
func AddCommand(ctx context.Context, root *cobra.Command, project project.Project) {
	pull := &cobra.Command{
		Use:   "pull",
		Short: "This command pulls files from Actions Console into the local file system.",
		Long:  "This command pulls files from Actions Console into the local file system.",
		RunE: func(cmd *cobra.Command, args []string) error {
			studioProj, ok := project.(studio.Studio)
			if !ok {
				return fmt.Errorf("can not convert %T to %T", project, studio.Studio{})
			}
			// Developer may run pull from an empty directory, in which case projectRoot doesn't yet
			// exist. In that case, os.Getwd() would be used.
			if studioProj.ProjectRoot() == "" {
				if err := (&studioProj).SetProjectRoot(); err != nil {
					return err
				}
			}
			pid, err := cmd.Flags().GetString("project-id")
			if err != nil {
				return err
			}
			if err := (&studioProj).SetProjectID(pid); err != nil {
				return err
			}
			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return err
			}
			clean, err := cmd.Flags().GetBool("clean")
			if err != nil {
				return err
			}
			versionID, err := cmd.Flags().GetString("version-id")
			if err != nil {
				return err
			}
			if versionID == "" {
				if err := sdk.ReadDraftJSON(ctx, studioProj, force, clean); err != nil {
					return err
				}
			} else {
				versionID = url.PathEscape(versionID)
				if err := sdk.ReadVersionJSON(ctx, studioProj, force, clean, versionID); err != nil {
					return err
				}
			}
			log.DoneMsgln(fmt.Sprintf("You should see the files written in %s", studioProj.ProjectRoot()))
			return nil
		},
		Args: cobra.NoArgs,
	}
	pull.Flags().String("project-id", "", "Pull from the project specified by the ID. The value provided in this flag will overwrite the value from settings file, if present.")
	pull.Flags().BoolP("force", "f", false, "Overwrite existing local files without asking.")
	pull.Flags().Bool("clean", false, "Remove any local files that are not in the files pulled from Actions Builder.")
	pull.Flags().String("version-id", "", "Pull the version specified by the ID.")
	root.AddCommand(pull)
}
