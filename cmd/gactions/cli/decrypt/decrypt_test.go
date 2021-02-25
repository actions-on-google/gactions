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

package decrypt

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func osAbs(p string) string {
	if runtime.GOOS == "windows" && strings.HasPrefix(p, "/") {
		return filepath.Join("c:", filepath.FromSlash(p))
	}
	return filepath.FromSlash(p)
}

func TestNormPath(t *testing.T) {
	tests := []struct {
		p    string
		root string
		want string
	}{
		{
			p:    "foo.txt",
			root: osAbs("/home/user/myproject/"),
			want: filepath.Join(osAbs("/home/user/myproject/"), "foo.txt"),
		},
		{
			p:    osAbs("/home/foo.txt"),
			root: osAbs("/home/user/myproject/"),
			want: osAbs("/home/foo.txt"),
		},
		{
			p:    osAbs("./bar/foo.txt"),
			root: osAbs("/home/user/myproject/"),
			want: osAbs("/home/user/myproject/bar/foo.txt"),
		},
		{
			p:    osAbs("../bar/foo.txt"),
			root: osAbs("/home/user/myproject/"),
			want: osAbs("/home/user/bar/foo.txt"),
		},
	}
	for _, tc := range tests {
		if got := normPath(tc.p, tc.root); got != tc.want {
			t.Errorf("normPath(%v, %v) returned %v, but want %v", tc.p, tc.root, got, tc.want)
		}
	}
}
