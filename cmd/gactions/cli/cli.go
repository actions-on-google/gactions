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

// Package cli contains shared CLI initialization steps.
package cli

import (
	"context"

	"github.com/actions-on-google/gactions/api/sdk"
	"github.com/actions-on-google/gactions/cmd/gactions/cli/decrypt"
	"github.com/actions-on-google/gactions/cmd/gactions/cli/deploy"
	"github.com/actions-on-google/gactions/cmd/gactions/cli/encrypt"
	"github.com/actions-on-google/gactions/cmd/gactions/cli/ginit"
	"github.com/actions-on-google/gactions/cmd/gactions/cli/login"
	"github.com/actions-on-google/gactions/cmd/gactions/cli/logout"
	"github.com/actions-on-google/gactions/cmd/gactions/cli/notices"
	"github.com/actions-on-google/gactions/cmd/gactions/cli/pull"
	"github.com/actions-on-google/gactions/cmd/gactions/cli/push"
	"github.com/actions-on-google/gactions/cmd/gactions/cli/releasechannels"
	"github.com/actions-on-google/gactions/cmd/gactions/cli/version"
	"github.com/actions-on-google/gactions/cmd/gactions/cli/versions"
	"github.com/actions-on-google/gactions/log"
	"github.com/actions-on-google/gactions/project/studio"
	"github.com/spf13/cobra"
)

const (
	verboseFlagName  = "verbose"
	consumerFlagName = "consumer"
)

// Command returns a *cobra.Command setup with the common set of commands
// and configuration already done.
func Command(ctx context.Context, name string, debug bool, ver string) *cobra.Command {
	root := &cobra.Command{
		Use:           name,
		Short:         "Command Line Interface for Google Actions SDK",
		SilenceUsage:  true,
		SilenceErrors: true, // Would like to print errors ourselves.
	}
	root.PersistentFlags().BoolP(verboseFlagName, "v", false, "Display additional error information")

	root.PersistentFlags().String(consumerFlagName, "", "String identifying the caller to Google")
	// This field is hidden as it's not documented and only used by tooling partners using the CLI.
	root.PersistentFlags().MarkHidden(consumerFlagName)

	projectRoot, err := studio.FindProjectRoot()
	if err != nil {
		projectRoot = "" // not found
	}
	// clientNotSoSecretJSON comes from go_embed_data rule in the BUILD file.
	// The client secret is encoded directly into the source code. It's okay
	// to do this based on the Google OAuth2 docs (see reference below).
	// Reference:
	//   https://developers.google.com/identity/protocols/OAuth2#installed
	project := studio.New(clientNotSoSecretJSON, projectRoot)
	ginit.AddCommand(ctx, root, project)
	push.AddCommand(ctx, root, project)
	deploy.AddCommand(ctx, root, project)
	login.AddCommand(ctx, root, project)
	logout.AddCommand(root, project)
	pull.AddCommand(ctx, root, project)
	encrypt.AddCommand(ctx, root, project)
	decrypt.AddCommand(ctx, root, project)
	version.AddCommand(root)
	notices.AddCommand(root)
	releasechannels.AddCommand(ctx, root, project)
	versions.AddCommand(ctx, root, project)

	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Init logging first since functions below may call log.
		if err := initLogging(cmd, debug); err != nil {
			return err
		}
		if err := setConsumer(cmd); err != nil {
			return err
		}
		return nil
	}
	return root
}

func setConsumer(cmd *cobra.Command) error {
	consumer, err := cmd.Flags().GetString(consumerFlagName)
	if err != nil {
		return err
	}
	sdk.Consumer = consumer
	log.Debugf("Set consumer to %s\n", consumer)
	return nil
}

func initLogging(cmd *cobra.Command, debug bool) error {
	isVerbose, err := cmd.Flags().GetBool(verboseFlagName)
	if err != nil {
		return err
	}
	if isVerbose {
		log.Severity = log.InfoLevel
	}
	// debug is the most permissive level
	if debug {
		log.Severity = log.DebugLevel
	}
	return nil
}

// Execute runs the command and displays errors. Returns the exit code for the CLI.
func Execute(cmd *cobra.Command) int {
	if err := cmd.Execute(); err != nil {
		log.Error(err)
		return 1
	}
	return 0
}
