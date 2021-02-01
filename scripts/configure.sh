#  Copyright 2021 Google LLC
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#
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
