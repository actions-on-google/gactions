//  Copyright 2021 Google LLC
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.
//
// Package versions provides an implementation of an action on "versions".
package versions

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"text/tabwriter"

	"github.com/actions-on-google/gactions/api/sdk"
	"github.com/actions-on-google/gactions/project"
	"github.com/actions-on-google/gactions/project/studio"
	"github.com/spf13/cobra"
)

var versionIDRegExp = regexp.MustCompile(`^projects/[^/]+/versions/(?P<versionID>[^/]+)$`)
var modifiedOnRegExp = regexp.MustCompile(`(?P<date>\d{4}-\d{2}-\d{2})+T+(?P<time>\d{2}:\d{2}:\d{2})(\.\d{6})+Z`)

// AddCommand adds the release-channels list sub-command to the passed in root command.
func AddCommand(ctx context.Context, root *cobra.Command, project project.Project) {
	versions := &cobra.Command{
		Use:   "versions",
		Short: "This is the main command for viewing and managing versions. See below for a complete list of sub-commands.",
		Long:  "This is the main command for viewing and managing versions. See below for a complete list of sub-commands.",
		Args:  cobra.MinimumNArgs(1),
	}
	list := &cobra.Command{
		Use:   "list",
		Short: "This command lists all versions and their metadata.",
		Long:  "This command lists all versions and their metadata.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			studioProj, ok := project.(studio.Studio)
			if !ok {
				return fmt.Errorf("can not convert %T to %T", project, studio.Studio{})
			}
			pid, err := cmd.Flags().GetString("project-id")
			if err != nil {
				return err
			}
			if err := (&studioProj).SetProjectID(pid); err != nil {
				return err
			}
			res, err := sdk.ListVersionsJSON(ctx, studioProj)
			if err != nil {
				return err
			}
			return printVersions(res)
		},
	}
	list.Flags().String("project-id", "", "List versions of the project specified by the ID. The value provided in this flag will overwrite the value from settings file, if present.")
	versions.AddCommand(list)
	root.AddCommand(versions)
}

func printVersions(versions []project.Version) error {
	w := new(tabwriter.Writer)
	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 20, 8, 1, '\t', 0)
	fmt.Fprintln(w, "Version\tStatus\tLast Modified By\tModified On\t")
	for _, version := range versions {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t\n", versionID(version.ID), version.State.Message, version.LastModifiedBy, formatModifiedOn(version.ModifiedOn))
	}
	fmt.Fprintf(w, "To learn more about release channels, visit https://developers.google.com/assistant/actionssdk/reference/rest/Shared.Types/ReleaseChannel.")
	fmt.Fprintln(w)
	return w.Flush()
}

func versionID(version string) string {
	versionIDMatch := versionIDRegExp.FindStringSubmatch(version)
	if versionIDMatch == nil {
		return "N/A"
	}
	return versionIDMatch[versionIDRegExp.SubexpIndex("versionID")]
}

func formatModifiedOn(modifiedOn string) string {
	modifiedOnMatch := modifiedOnRegExp.FindStringSubmatch(modifiedOn)
	if modifiedOnMatch == nil {
		return "N/A"
	}

	return modifiedOnMatch[modifiedOnRegExp.SubexpIndex("date")] + " " + modifiedOnMatch[modifiedOnRegExp.SubexpIndex("time")]
}
