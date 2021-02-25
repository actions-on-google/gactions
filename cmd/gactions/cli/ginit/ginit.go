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

// Package ginit provides an implementation of "gactions init" command.
// Note(atulep): Switching package name to "init" led to compilation errors.
package ginit

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/actions-on-google/gactions/api/sdk"
	"github.com/actions-on-google/gactions/log"
	"github.com/actions-on-google/gactions/project"

	"github.com/spf13/cobra"
)

var (
	samples []project.SampleProject
)

func isValidProject(projectTitle string) bool {
	for _, v := range samples {
		if v.Name == projectTitle {
			return true
		}
	}
	return false
}

func printSamples(samples []project.SampleProject) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 4, 0, '\t', 0)
	for i, v := range samples {
		fmt.Fprintf(w, "%v) %v\t\n", i+1, v.Name)
	}
	fmt.Fprintln(w)
	w.Flush()
}

var availableProjects = func(ctx context.Context, project project.Project) ([]project.SampleProject, error) {
	return sdk.ListSampleProjectsJSON(ctx, project)
}

// AddCommand adds the init sub-command to the passed in root command.
func AddCommand(ctx context.Context, root *cobra.Command, project project.Project) {
	init := &cobra.Command{
		Use:   "init",
		Short: "Initialize a directory for a new project.",
		Long:  "This command places sample Actions SDK project files into the current directory. You can choose from a list of sample projects. Current directory must be empty.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return doInit(cmd, args, project)
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("unexpected arguments: %v", args)
			}
			l, err := availableProjects(ctx, project)
			if err != nil {
				return err
			}
			samples = l
			if len(args) < 1 || !isValidProject(args[0]) {
				log.Outf("Invalid sample specified: %v. Please select one of the following:\n\n", args)
				printSamples(samples)
				return fmt.Errorf("invalid sample specified: %v", args)
			}
			return nil
		},
	}
	init.Flags().String("dest", ".", `Specify a directory for placing the project files (the default directory is ".")`)
	root.AddCommand(init)
}

func doInit(cmd *cobra.Command, args []string, proj project.Project) error {
	destination, _ := cmd.Flags().GetString("dest")
	if alreadySetup := proj.AlreadySetup(destination); alreadySetup {
		log.Outf("%s is not empty. Make sure to create an empty directory and run \"gactions init\" from there.", destination)
		return fmt.Errorf("%s is not empty", destination)
	}
	log.Outf("Writing sample files for %v to %s\n", args[0], destination)
	var s project.SampleProject
	for _, v := range samples {
		if v.Name == args[0] {
			s = v
		}
	}
	if err := proj.Download(s, destination); err != nil {
		return err
	}
	log.DoneMsgln("Please checkout the following documentation - https://developers.google.com/assistant/conversational/build on the next steps on how to get started.")
	return nil
}
