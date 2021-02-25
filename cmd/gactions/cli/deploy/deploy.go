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

// Package deploy provides an implementation of "gactions deploy" command.
package deploy

import (
	"context"
	"fmt"

	"github.com/actions-on-google/gactions/api/sdk"
	"github.com/actions-on-google/gactions/project"
	"github.com/actions-on-google/gactions/project/studio"
	"github.com/spf13/cobra"
)

func setProjectID(project *project.Project) error {
	studioProj, ok := (*project).(studio.Studio)
	if !ok {
		return fmt.Errorf("can not convert %T to %T", project, studio.Studio{})
	}
	if err := (&studioProj).SetProjectID(""); err != nil {
		return err
	}
	*project = studioProj
	return nil
}

// AddCommand adds the deploy sub-command to the passed in root command.
func AddCommand(ctx context.Context, root *cobra.Command, project project.Project) {
	deploy := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy an Action to the specified channel.",
		Long:  "This command deploys an Action to the specified channel.",
		Args:  cobra.MinimumNArgs(1),
	}
	preview := &cobra.Command{
		Use:   "preview",
		Short: "Deploy for preview.",
		Long:  "This command deploys an Action to preview, so you can test your Action in the simulator.",
		RunE: func(cmd *cobra.Command, args []string) error {
			sandbox, _ := cmd.Flags().GetBool("sandbox")
			if err := setProjectID(&project); err != nil {
				return err
			}
			return sdk.WritePreviewJSON(ctx, project, sandbox)
		},
	}
	preview.Flags().Bool("sandbox", true,
		"Indicates whether or not to run certain operations, such as transactions, in sandbox mode. The default value is set to true")
	alpha := &cobra.Command{
		Use:   "alpha",
		Short: "Deploy to alpha channel.",
		Long:  "This command deploys to alpha channel.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := setProjectID(&project); err != nil {
				return err
			}
			return sdk.CreateVersionJSON(ctx, project, sdk.AlphaChannel)
		},
	}
	beta := &cobra.Command{
		Use:   "beta",
		Short: "Deploy to beta channel.",
		Long:  "This command deploys to beta channel.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := setProjectID(&project); err != nil {
				return err
			}
			return sdk.CreateVersionJSON(ctx, project, sdk.BetaChannel)
		},
	}
	prod := &cobra.Command{
		Use:   "prod",
		Short: "Deploy to production channel.",
		Long:  "This command deploys to production channel.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := setProjectID(&project); err != nil {
				return err
			}
			return sdk.CreateVersionJSON(ctx, project, sdk.ProdChannel)
		},
	}
	deploy.AddCommand(preview)
	deploy.AddCommand(alpha)
	deploy.AddCommand(beta)
	deploy.AddCommand(prod)
	root.AddCommand(deploy)
}
