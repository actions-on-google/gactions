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

// Package decrypt provides an implementation of "gactions decrypt" command.
package decrypt

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/actions-on-google/gactions/api/sdk"
	"github.com/actions-on-google/gactions/log"
	"github.com/actions-on-google/gactions/project"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func parseClientSecret(files map[string][]byte) (string, error) {
	type secretFile struct {
		EncryptedClientSecret string `yaml:"encryptedClientSecret"`
	}
	in, ok := files["settings/accountLinkingSecret.yaml"]
	if !ok {
		log.Infoln("accountLinkingSecret.yaml not found in project files")
		return "", errors.New("accountLinkingSecret.yaml not found in project files. " +
			"Try encrypting your client secret first, or pulling an existing project with a client secret")
	}
	f := secretFile{}
	if err := yaml.Unmarshal(in, &f); err != nil {
		return "", err
	}
	return f.EncryptedClientSecret, nil
}

func expandTilde(p string) (string, error) {
	if !strings.HasPrefix(p, "~/") || runtime.GOOS == "windows" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return p, err
	}
	return filepath.Join(home, p[2:]), nil
}

func normPath(p, root string) string {
	norm := filepath.FromSlash(p)
	if strings.HasPrefix(p, "~/") {
		t, err := expandTilde(p)
		if err == nil {
			norm = t
		}
	}
	if strings.HasPrefix(p, "./") || strings.HasPrefix(p, "../") || strings.HasPrefix(p, ".\\") || strings.HasPrefix(p, "..\\") {
		norm = filepath.Clean(filepath.Join(root, p))
	}
	if !filepath.IsAbs(norm) {
		norm = filepath.Join(root, p)
	}
	return norm
}

// AddCommand adds decrypt sub-command to the passed in root command.
func AddCommand(ctx context.Context, root *cobra.Command, proj project.Project) {
	decrypt := &cobra.Command{
		Use:   "decrypt <plaint-text-file>",
		Short: "Decrypt client secret.",
		Long:  "This command decrypts the client secret key used in Account Linking. Specify a file path for the decrypt output. This can be a relative or absolute path.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if proj.ProjectRoot() == "" {
				log.Errorf("Can't find a project root: manifest.yaml was not found in this or any of the parent folders.")
				return errors.New("can not find manifest.yaml")
			}
			files, err := proj.Files()
			if err != nil {
				return err
			}
			s, err := parseClientSecret(files)
			if err != nil {
				return err
			}
			out := normPath(args[0], proj.ProjectRoot())
			return sdk.DecryptSecretJSON(ctx, proj, s, out)
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("unexpected arguments: %v", args)
			}
			if len(args) < 1 {
				return fmt.Errorf(`<plain-text-file> argument is missing. Try "gactions decrypt <pathToPlainTextFile>"`)
			}
			return nil
		},
	}
	root.AddCommand(decrypt)
}
