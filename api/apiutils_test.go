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

package apiutils

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/actions-on-google/gactions/api/testutils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/oauth2"
)

func TestRemoveTokenExists(t *testing.T) {
	ogTCF := tokenCacheFile
	ogRT := revokeToken
	t.Cleanup(func() {
		tokenCacheFile = ogTCF
		revokeToken = ogRT
	})
	d, err := ioutil.TempDir(testutils.TestTmpDir, ".credentials")
	if err != nil {
		t.Errorf("Failed to create a temp dir: got %v", err)
	}
	f, err := ioutil.TempFile(d, "test-creds")
	if err != nil {
		t.Errorf("Failed to create a temp file: got %v", err)
	}
	defer func() {
		os.RemoveAll(d)
	}()
	tokenCacheFile = func() (string, error) {
		return f.Name(), nil
	}
	revokeToken = func(tokenFile []byte) error {
		return nil
	}
	if err := RemoveToken(); err != nil {
		t.Errorf("RemoveToken returned %v, want %v", err, nil)
	}
}

func TestRemoveTokenDoesNotExist(t *testing.T) {
	if err := RemoveToken(); err == nil {
		t.Error("RemoveToken returned %v, want error", err)
	}
}

func createCachedTokenFile(cachedToken *oauth2.Token) (string, string, error) {
	dirName, err := ioutil.TempDir(testutils.TestTmpDir, ".credentials")
	if err != nil {
		return "", "", err
	}
	cachedTokenBytes, _ := json.Marshal(cachedToken)
	err = ioutil.WriteFile(path.Join(dirName, "gactions-test-secret.json"), cachedTokenBytes, 0644)
	return dirName, path.Join(dirName, "gactions-test-secret.json"), err
}

func TestTokenWhenCachedTokenExists(t *testing.T) {
	cachedToken := oauth2.Token{
		AccessToken:  "123",
		RefreshToken: "456",
	}
	conf := oauth2.Config{}
	cachedFileDir, cachedFilename, err := createCachedTokenFile(&cachedToken)
	defer os.RemoveAll(cachedFileDir)
	if err != nil {
		t.Fatalf("Can't create temporary files under %q: %v", cachedFilename, err)
	}
	tok, err := token(context.Background(), &conf, cachedFilename, false)
	if err != nil {
		t.Errorf("GetToken returned %v, but want %v", err, nil)
	}
	if *tok != cachedToken {
		t.Errorf("GetToken returns %v when the cached token is %v", *tok, cachedToken)
	}
}

func TestTokenWhenCachedTokenDoesNotExist(t *testing.T) {
	cachedToken := oauth2.Token{
		AccessToken:  "123",
		RefreshToken: "456",
	}
	conf := oauth2.Config{}
	originalFn := interactiveTokenCopyPaste
	interactiveTokenCopyPaste = func(ctx context.Context, conf *oauth2.Config) (*oauth2.Token, error) {
		return &cachedToken, nil
	}
	defer func() {
		interactiveTokenCopyPaste = originalFn
	}()
	tok, err := token(context.Background(), &conf, "", false)
	if err != nil {
		t.Errorf("GetToken returned %v, but want %v", err, nil)
	}
	if *tok != cachedToken {
		t.Errorf("GetToken returns %v when the cached token is %v", *tok, cachedToken)
	}
}

func TestNewHTTPClientWhenCachedTokenDoesNotExist(t *testing.T) {
	// Pass in a null token file
	_, err := NewHTTPClient(context.Background(), []byte(`{"installed":{"redirect_uris":["urn:ietf:wg:oauth:2.0:oob","http://localhost"]}}`), "/tmp/token")
	if err == nil {
		t.Errorf("NewHTTPClient should throw an error when the cached token does not exist")
	}
}

func TestAuthSavesToken(t *testing.T) {
	originalToken := token
	originalCacheFile := tokenCacheFile
	defer func() {
		token = originalToken
		tokenCacheFile = originalCacheFile
	}()
	want := oauth2.Token{
		AccessToken:  "123",
		RefreshToken: "456",
	}
	token = func(ctx context.Context, config *oauth2.Config, tokenCacheFilename string, launch bool) (*oauth2.Token, error) {
		return &want, nil
	}
	d, err := ioutil.TempDir(testutils.TestTmpDir, ".credentials")
	if err != nil {
		t.Fatalf("Failed to create a temporary directory: got %v", err)
	}
	defer os.RemoveAll(d)
	tokenCacheFile = func() (string, error) {
		return filepath.Join(d, "file.json"), nil
	}
	err = Auth(context.Background(), []byte(`{"installed":{"redirect_uris":["urn:ietf:wg:oauth:2.0:oob","http://localhost"]}}`))
	if err != nil {
		t.Errorf("Auth returned %v, but want %v", err, nil)
	}
	b, err := ioutil.ReadFile(filepath.Join(d, "file.json"))
	if err != nil {
		t.Errorf("Failed to read a file containing the token created by Auth: got %v", err)
	}
	var got oauth2.Token
	if err := json.Unmarshal(b, &got); err != nil {
		t.Errorf("Auth should have written a syntactically correct JSON, but got %v", err)
	}
	if !cmp.Equal(got, want, cmpopts.IgnoreUnexported(oauth2.Token{})) {
		t.Errorf("Auth should have saved %v to disc, but wrote %v instead", want, got)
	}
}
