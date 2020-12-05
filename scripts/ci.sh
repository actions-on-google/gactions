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
