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
		Short: "This command performs release channels specific actions.",
		Long:  "This command performs release channels specific actions.",
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
