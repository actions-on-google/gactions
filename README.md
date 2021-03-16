# gactions CLI

This repository contains source code for the CLI used to communicate with
[Actions SDK](https://developers.google.com/assistant/actionssdk/gactions).

## Install

The latest version of this tool can be installed [directly from developers.google.com]((https://developers.google.com/assistant/actionssdk/gactions))
or via NPM:

```
npm install -g @assistant/gactions
```

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

## Google Cloud Project Setup

1.  Create a [Google Cloud project](https://console.developers.google.com).
1.  Enable the Actions API for your project:
    1. Visit the [Google API console](https://console.developers.google.com/apis/library) and select your project from the **Select a project** dropdown.
    1. If the Actions API is not enabled, search for *"Actions API"* and click **Enable**.
1.  Create a OAuth Client Id:
    1. Visit the [Google Cloud console credentials page](https://console.developers.google.com/apis/credentials) and select your project from the **Select a project** dropdown.
    1. Click the "Create credentials" button and then select "OAuth Client Id".
    1. Select **Application Type** "Desktop app"
    1. Enter an OAuth client name and click **Create**.
    1. Configure the [OAuth consent screen](https://console.cloud.google.com/apis/credentials/consent) and add your email as a test account.
    1. Return to the [Google Cloud console credentials page](https://console.developers.google.com/apis/credentials)
    1. Click the download icon next to the newly-created OAuth client name to download the JSON file.
    1. Move the file to a file named `data/client_not_so_secret.json` in this directory.

**Note**: If you are running this alongside the [official gactions binary](https://developers.google.com/assistant/actionssdk/gactions)
, you will need to regenerate authentication credentials. To reset these
credentials:

1. Run `bazel run gactions logout`.
1. Run `bazel run gactions login`.

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
