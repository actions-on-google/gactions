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

package testutils

import (
	"io/ioutil"
	"log"
	"os"
	"runtime"
)

const (
	TestTmpDir = "/tmp"
)

func TestTmpRoot() string {
	if runtime.GOOS == "windows" {
		return "C:\\"
	}
	return ""
}

// ReadFileOrDie is a version of ReadFile that is fatal if not successful.
func ReadFileOrDie(path string) []byte {
	cwd, err := os.Getwd()
	r, err := os.Open(path)
	if err != nil {
		log.Fatalf("Cannot open file %v%v%v", cwd, os.PathSeparator, path)
	}
	defer r.Close()
	b, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatalf("Cannot read file %v%v%v", cwd, os.PathSeparator, path)
	}
	return b
}
