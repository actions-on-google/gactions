# gactions v2

This repository contains source code for the CLI used to communicate with
[Actions SDK](https://developers.google.com/assistant/actionssdk/gactions).

## Build Instructions

To build this project, you will require [Bazel](https://bazel.build/).

Build the CLI using `:gactions` Bazel rule.:

*   Linux: `bazel build //gactions`.

*   Windows: `bazel build --cpu=x86_64-windows //gactions`

*   Mac `bazel build --config=darwin_x86_64 //gactions`

You will find the binary in `bazel-bin/gactions_/gactions`.

### Run Instructions

To build and run the project:

1. `bazel run gactions`

Pass parameters:

1. `bazel run gactions -- <args>`

Example:

`bazel run gactions -- version`

This should print out the current version from `versions/version_names.bzl`.
