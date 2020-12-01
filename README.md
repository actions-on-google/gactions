# gactions v2

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
