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

// Package encrypt provides an implementation of "gactions encrypt" command.
package encrypt

import (
	"context"
	"errors"
	"syscall"

	"github.com/actions-on-google/gactions/api/sdk"
	"github.com/actions-on-google/gactions/log"
	"github.com/actions-on-google/gactions/project"
	"github.com/golang/crypto/ssh/terminal"
	"github.com/spf13/cobra"
)

func askForSecret() (string, error) {
	log.Outf("Write your secret: ")
	secret, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	return string(secret), nil
}

// AddCommand adds encrypt sub-command to the passed in root command.
func AddCommand(ctx context.Context, root *cobra.Command, proj project.Project) {
	encrypt := &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt client secret.",
		Long:  "This commands encrypts the client secret key used in Account linking.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if proj.ProjectRoot() == "" {
				log.Errorf(`Can't find a project root. This may be because (1) %q was not found in this or any of the parent folders, or (2) if %q was found, but the key "sdkPath" was missing, or (3) if %q and manifest.yaml were both not found.`, project.ConfigName, project.ConfigName, project.ConfigName)
				return errors.New("can not determine project root")
			}
			s, err := askForSecret()
			if err != nil {
				return err
			}
			return sdk.EncryptSecretJSON(ctx, proj, s)
		},
		Args: cobra.NoArgs,
	}
	root.AddCommand(encrypt)
}
