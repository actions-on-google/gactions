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
		fmt.Fprintf(w, "%v\t%v\t%v\t\n", getReleaseChannelName(releaseChannel.Name), getVersionID(releaseChannel.CurrentVersion), getVersionID(releaseChannel.PendingVersion))
	}
	fmt.Fprintln(w)
	w.Flush()
}

func getReleaseChannelName(releaseChannel string) string {
	releaseChannelNameRegExp := regexp.MustCompile("^projects/[^//]+/releaseChannels/(?P<releaseChannelName>[^//]+)$")
	if releaseChannelMatch := releaseChannelNameRegExp.FindStringSubmatch(releaseChannel); releaseChannelMatch == nil {
		return "N/A"
	}
	return releaseChannelNameRegExp.FindStringSubmatch(releaseChannel)[releaseChannelNameRegExp.SubexpIndex("releaseChannelName")]
}

func getVersionID(version string) string {
	versionIDRegExp := regexp.MustCompile("^projects/[^//]+/versions/(?P<versionID>[^//]+)$")
	if versionIDMatch := versionIDRegExp.FindStringSubmatch(version); versionIDMatch == nil {
		return "N/A"
	}
	return versionIDRegExp.FindStringSubmatch(version)[versionIDRegExp.SubexpIndex("versionID")]
}
