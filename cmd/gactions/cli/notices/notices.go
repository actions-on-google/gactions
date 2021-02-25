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

// Package notices provides an implementation of "gactions third-party-notices" command.
package notices

import (
	"fmt"

	"github.com/actions-on-google/gactions/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type licenseObj struct {
	Title   string `yaml:"title"`
	Content string `yaml:"content"`
}

func parse(v []byte) ([]licenseObj, error) {
	obj := []licenseObj{}
	if err := yaml.Unmarshal(v, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

// AddCommand adds the push sub-command to the passed in root command.
func AddCommand(root *cobra.Command) {
	notices := &cobra.Command{
		Use:   "third-party-notices",
		Short: "Prints license files of third-party software used.",
		Long:  "Prints license files of third-party software used in CLI source code.",
		Run: func(cmd *cobra.Command, args []string) {
			// licenseFiles is a map where a title is the name of the library and content is its license.
			for _, v := range licenseFiles {
				licenses, err := parse(v)
				if err != nil {
					fmt.Println(err)
					return
				}
				for _, v := range licenses {
					log.Outf("Software: %s\n", string(v.Title))
					log.Outf("License:\n%s\n", string(v.Content))
				}
			}
		},
		Args: cobra.NoArgs,
	}
	root.AddCommand(notices)
}
