rem CI script

rem Install Bazel using Choco
echo "Installing/updating bazel"
choco install bazel -y

rem Use placeholder secret file
mkdir data\
echo '
    {"installed":{"client_id":"PLACEHOLDER.apps.googleusercontent.com","project_id":"PLACEHOLDER","auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://oauth2.googleapis.com/token","auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs","client_secret":"PLACEHOLDER","redirect_uris":["urn:ietf:wg:oauth:2.0:oob","http://localhost"]}}
' > data\client_not_so_secret.json

rem Ensure that our code builds
bazel build gactions
rem Print version
bazel run gactions -- version

rem Run all tests
bazel test --test_output=errors ...
