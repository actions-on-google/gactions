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
# CI script

# Install Bazel
# CI environment script
bash use_bazel.sh 3.5.0

# Should install the latest stable version of Bazel
echo $(bazel version)

# Use placeholder secret file
mkdir -p ./data/
cat <<< '
    {"installed":{"client_id":"PLACEHOLDER.apps.googleusercontent.com","project_id":"PLACEHOLDER","auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://oauth2.googleapis.com/token","auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs","client_secret":"PLACEHOLDER","redirect_uris":["urn:ietf:wg:oauth:2.0:oob","http://localhost"]}}
' > ./data/client_not_so_secret.json

# Ensure that our code builds
bazel build gactions
# Print version
echo $(bazel run gactions -- version)

# Run all tests
bazel test --test_output=errors ...
