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

// Package version implements "gactions version" command.
package version

import (
	"github.com/actions-on-google/gactions/log"
	"github.com/actions-on-google/gactions/versions"
	"github.com/spf13/cobra"
)

// AddCommand adds the push sub-command to the passed in root command.
func AddCommand(root *cobra.Command) {
	version := &cobra.Command{
		Use:   "version",
		Short: "Prints current version of the CLI.",
		Long:  "Prints current version of the CLI.",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Outf("%s\n", versions.CliVersion)
			return nil
		},
		Args: cobra.NoArgs,
	}
	root.AddCommand(version)
}
