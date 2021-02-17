# gactions CLI

This repository contains source code for the CLI used to communicate with
[Actions SDK](https://developers.google.com/assistant/actionssdk/gactions).

## Usage

```
Command Line Interface for Google Actions SDK

Usage:
  gactions [command]

Available Commands:
  decrypt             Decrypt client secret.
  deploy              Deploy an Action to the specified channel.
  encrypt             Encrypt client secret.
  help                Help about any command
  init                Initialize a directory for a new project.
  login               Authenticate gactions CLI to your Google account via web browser.
  logout              Log gactions CLI out of your Google Account.
  pull                This command pulls files from Actions Console into the local file system.
  push                This command pushes changes in the local files to Actions Console.
  release-channels    This is the main command for viewing and managing release channels. See below for a complete list of sub-commands.
  third-party-notices Prints license files of third-party software used.
  version             Prints current version of the CLI.
  versions            This is the main command for viewing and managing versions. See below for a complete list of sub-commands.

Flags:
  -h, --help      help for gactions
  -v, --verbose   Display additional error information

Use "gactions [command] --help" for more information about a command.
```

### Quick Start

```bash
# Start an authentication flow.
# This requires a web browser.
# When the flow is complete, the CLI automatically authenticates.
gactions login

# Initialize a sample project
gactions init hello-world --dest hello-world-sample
cd hello-world-sample

# Open the `sdk/settings/settings.yaml` file and change the value of
# `projectId` to your project's ID.
$EDITOR sdk/settings/settings.

# From the hello-world-sample/sdk/ directory, run the following
# command to push the local version of your Actions project to the
# console as a draft version.
cd sdk
gactions push

# From the hello-world-sample/sdk/ directory, run the following
# command to test your Actions project in the simulator.
gactions deploy preview
```

Read the [quick start documentation](https://developers.google.com/assistant/conversational/quickstart) to learn more.

## Build Instructions

To build this project, you will require [Bazel](https://bazel.build/).

Build the CLI using `:gactions` Bazel rule.:

*   Linux: `bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 gactions`.

*   Windows: `bazel build --platforms=@io_bazel_rules_go//go/toolchain:windows_amd64 gactions`

*   Mac `bazel build --platforms=@io_bazel_rules_go//go/toolchain:darwin_amd64 gactions`

You will find the binary in `bazel-bin/gactions_/gactions`.

**Note**: `bazel build gactions` will build for your local platform.

### Run Instructions

To simultaneously build and run the project:

1. `bazel run gactions`

Pass parameters:

1. `bazel run gactions -- <args>`

Example:

`bazel run gactions -- version`

This should print out the current version from `versions/version_names.bzl`.

## References & Issues
+ Questions? Go to [StackOverflow](https://stackoverflow.com/questions/tagged/actions-on-google) or [Assistant Developer Community on Reddit](https://www.reddit.com/r/GoogleAssistantDev/).
+ For bugs, please report [an issue](https://github.com/actions-on-google/gactions/issues/new) on Github.
+ Actions on Google [Documentation](https://developers.google.com/assistant)
+ Actions on Google [Codelabs](https://codelabs.developers.google.com/?cat=Assistant).

## Make Contributions
Please read and follow the steps in the [CONTRIBUTING.md](CONTRIBUTING.md).

## License
See [LICENSE](LICENSE).
