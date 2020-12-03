# gactions CLI

This repository contains source code for the CLI used to communicate with
[Actions SDK](https://developers.google.com/assistant/actionssdk/gactions).

## Build Instructions

To build this project, you will require [Bazel](https://bazel.build/).

Build the CLI using `:gactions` Bazel rule.:

*   Linux: `bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 gactions`.

*   Windows: `bazel build --platforms=@io_bazel_rules_go//go/toolchain:windows_amd64 gactions`

*   Mac `bazel build --platforms=@io_bazel_rules_go//go/toolchain:darwin_amd64 gactions`

You will find the binary in `bazel-bin/gactions_/gactions`.

**Note**: `bazel build gactions` will build for your local platform.

### Run Instructions

To setup your environment:

1. `source ./scripts/configure.sh`

To simultaneously build and run the project:

1. `bazel run gactions`

Pass parameters:

1. `bazel run gactions -- <args>`

Example:

`bazel run gactions -- version`

This should print out the current version from `versions/version_names.bzl`.

## References & Issues
+ Questions? Go to [StackOverflow](https://stackoverflow.com/questions/tagged/actions-on-google) or [Assistant Developer Community on Reddit](https://www.reddit.com/r/GoogleAssistantDev/).
+ For bugs, please report an issue on Github.
+ Actions on Google [Documentation](https://developers.google.com/assistant)
+ Actions on Google [Codelabs](https://codelabs.developers.google.com/?cat=Assistant).

## Make Contributions
Please read and follow the steps in the [CONTRIBUTING.md](CONTRIBUTING.md).

## License
See [LICENSE](LICENSE).
