// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Binary gactions is a command-line interface to the Actions SDK.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/actions-on-google/gactions/cli"
)

// Those variables are passed in via -X flags (see BUILD)
var (
	version    string
)

func cliVersion(semver string) string {
	return fmt.Sprintf("%s", semver)
}

func main() {
	ctx := context.Background()
	cmd := cli.Command(ctx, "gactions", false, cliVersion(version))
	os.Exit(cli.Execute(cmd))
}
