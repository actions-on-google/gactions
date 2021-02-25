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

// Package login provides an implementation of "gactions login" command.
package login

import (
	"context"

	"github.com/actions-on-google/gactions/api/apiutils"
	"github.com/actions-on-google/gactions/log"
	"github.com/actions-on-google/gactions/project"
	"github.com/spf13/cobra"
)

// AddCommand adds the push sub-command to the passed in root command.
func AddCommand(ctx context.Context, root *cobra.Command, proj project.Project) {
	login := &cobra.Command{
		Use:   "login",
		Short: "Authenticate gactions CLI to your Google account via web browser.",
		Long:  "Authenticate gactions CLI to your Google account via web browser.",
		RunE: func(cmd *cobra.Command, args []string) error {
			secret, err := proj.ClientSecretJSON()
			if err != nil {
				return err
			}
			if err := apiutils.Auth(ctx, secret); err != nil {
				return err
			}
			log.DoneMsgln("Successfully logged in.")
			return nil
		},
		Args: cobra.NoArgs,
	}
	root.AddCommand(login)
}
