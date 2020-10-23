#!/bin/bash

# Check that Bazel exists
bazel version > /dev/null || {
  echo "You need to install Bazel. Please visit https://bazel.build/"
  exit 1
}

# Append our PATH to point to compiled gactions
PATH_TO_GACTIONS="$(pwd)/bazel-bin/gactions_/"
PATH_TO_GACTIONS_DEBUG="$(pwd)/bazel-bin/gactions_debug_/"
export PATH="${PATH_TO_GACTIONS}:${PATH}"
export PATH="${PATH_TO_GACTIONS_DEBUG}:${PATH}"
