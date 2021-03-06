// Copyright 2021 Google LLC
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
//
// Package releasechannels provides an implementation of an action on "release-channels".
package releasechannels

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

var releaseChannelNameRegExp = regexp.MustCompile(`^projects/[^/]+/releaseChannels/(?P<releaseChannelName>[^/]+)$`)
var releaseChannelPrefixRegExp = regexp.MustCompile(`^actions[\.]channels[\.](?P<unknownBuiltInReleaseChannelName>[^/]+)$`)
var versionIDRegExp = regexp.MustCompile(`^projects/[^/]+/versions/(?P<versionID>[^/]+)$`)

// AddCommand adds the release-channels list sub-command to the passed in root command.
func AddCommand(ctx context.Context, root *cobra.Command, project project.Project) {
	releaseChannels := &cobra.Command{
		Use:   "release-channels",
		Short: "This is the main command for viewing and managing release channels. See below for a complete list of sub-commands.",
		Long:  "This is the main command for viewing and managing release channels. See below for a complete list of sub-commands.",
		Args:  cobra.MinimumNArgs(1),
	}
	list := &cobra.Command{
		Use:   "list",
		Short: "This command lists information about release channels for the project and their current and pending versions.",
		Long:  "This command lists information about release channels for the project and their current and pending versions.",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			res, err := sdk.ListReleaseChannelsJSON(ctx, studioProj)
			if err != nil {
				return err
			}
			printReleaseChannels(res)
			return nil
		},
	}
	list.Flags().String("project-id", "", "List release channels of the project specified by the ID. The value provided in this flag will overwrite the value from settings file, if present.")
	releaseChannels.AddCommand(list)
	root.AddCommand(releaseChannels)
}

func printReleaseChannels(releaseChannels []project.ReleaseChannel) {
	w := new(tabwriter.Writer)
	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 40, 8, 1, '\t', 0)
	fmt.Fprintln(w, "Release Channel\tCurrent Version\tPending Version\t")
	for _, releaseChannel := range releaseChannels {
		fmt.Fprintf(w, "%v\t%v\t%v\t\n", releaseChannelName(releaseChannel.Name), versionID(releaseChannel.CurrentVersion), versionID(releaseChannel.PendingVersion))
	}
	fmt.Fprintf(w, "To learn more about release channels, visit https://developers.google.com/assistant/actionssdk/reference/rest/Shared.Types/ReleaseChannel.")
	fmt.Fprintln(w)
	w.Flush()
}

func releaseChannelName(releaseChannel string) string {
	releaseChannelMatch := releaseChannelNameRegExp.FindStringSubmatch(releaseChannel)
	if releaseChannelMatch == nil {
		return "N/A"
	}
	releaseChannelName := releaseChannelMatch[releaseChannelNameRegExp.SubexpIndex("releaseChannelName")]

	// If release channel is a known built-in release channel with a short name, fetch the short name and display it.
	displayReleaseChannelName, found := sdk.BuiltInReleaseChannels[releaseChannelName]
	if found {
		return displayReleaseChannelName
	}

	// Else, check for prefix "actions.channels", and if present, remove it. This is used as a catch for built in channels without short names in the map.
	releaseChannelPrefixMatch := releaseChannelPrefixRegExp.FindStringSubmatch(releaseChannelName)
	if releaseChannelPrefixMatch == nil {
		return releaseChannelName
	}
	return releaseChannelPrefixMatch[releaseChannelPrefixRegExp.SubexpIndex("unknownBuiltInReleaseChannelName")]
}

func versionID(version string) string {
	if versionIDMatch := versionIDRegExp.FindStringSubmatch(version); versionIDMatch == nil {
		return "N/A"
	}
	return versionIDRegExp.FindStringSubmatch(version)[versionIDRegExp.SubexpIndex("versionID")]
}
