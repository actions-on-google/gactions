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

// Package apiutils contains utility functions to simplify working with gRPC libraries.
package apiutils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"
	"text/template"
	"time"

	"github.com/actions-on-google/gactions/log"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2"
)

const (
	builderAPIScope = "https://www.googleapis.com/auth/actions.builder"
	loginPrompt     = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>gactions CLI</title>

    <style media="screen">
      body { background: #ECEFF1; color: rgba(0,0,0,0.87); font-family: Roboto, Helvetica, Arial, sans-serif; margin: 0; padding: 0; }
      #message { background: white; max-width: 360px; margin: 100px auto 16px; padding: 32px 24px 8px; border-radius: 3px; }
      #message h2 { color: #4caf50; font-weight: bold; font-size: 16px; margin: 0 0 8px; }
      #message h1 { font-size: 22px; font-weight: 300; color: rgba(0,0,0,0.6); margin: 0 0 16px;}
      #message p { line-height: 140%; margin: 16px 0 24px; font-size: 14px; }
      #message a { display: block; text-align: center; background: #039be5; text-transform: uppercase; text-decoration: none; color: white; padding: 16px; border-radius: 4px; }
      #message, #message a { box-shadow: 0 1px 3px rgba(0,0,0,0.12), 0 1px 2px rgba(0,0,0,0.24); }
      #load { color: rgba(0,0,0,0.4); text-align: center; font-size: 13px; }
      @media (max-width: 600px) {
        body, #message { margin-top: 0; background: white; box-shadow: none; }
        body { border-top: 16px solid #4caf50; }
      }
      code { font-size: 18px; color: #999; }
    </style>
  </head>
  <body>
    <div id="message">
      <h2>{{.H2}}</h2>
      <h1>{{.H1}}</h1>
      <p>{{.P}}</p>
    </div>
  </body>
</html>
`
)

// NewHTTPClient returns a *http.Client created with all required scopes and permissions.
func NewHTTPClient(ctx context.Context, clientSecretKeyFile []byte, tokenFilepath string) (*http.Client, error) {
	config, err := google.ConfigFromJSON(clientSecretKeyFile, builderAPIScope)
	if err != nil {
		return nil, err
	}
	tokenCacheFilename := ""
	if tokenFilepath == "" {
		tokenCacheFilename, err = tokenCacheFile()
		if err != nil {
			return nil, err
		}
	} else {
		tokenCacheFilename = tokenFilepath
	}
	if !exists(tokenCacheFilename) {
		log.Infoln("Could not locate OAuth2 token")
		return nil, errors.New(`command requires authentication. try to run "gactions login" first`)
	}
	tok, err := tokenFromFile(tokenCacheFilename)
	if err != nil {
		return nil, err
	}
	return config.Client(ctx, tok), nil
}

// Auth prompts user for authentication token and writes it to disc.
func Auth(ctx context.Context, clientSecretKeyFile []byte) error {
	config, err := google.ConfigFromJSON(clientSecretKeyFile, []string{builderAPIScope}...)
	if err != nil {
		return err
	}
	// Get OAuth2 token from the user. It will be written into cacheFilename.
	tokenCacheFilename, err := tokenCacheFile()
	if err != nil {
		return err
	}
	// Check the shell is appropriate for use of launched browsers, otherwise present the copy/paste
	// flow.
	nonSSH := checkShell()
	notWindows := runtime.GOOS != "windows"
	tok, err := token(ctx, config, tokenCacheFilename, nonSSH && notWindows)
	if err != nil {
		return err
	}
	if err := saveToken(tokenCacheFilename, tok); err != nil {
		return err
	}
	return nil
}

// RemoveToken deletes the stored token
func RemoveToken() error {
	s, err := tokenCacheFile()
	if err != nil {
		return err
	}
	if !exists(s) {
		log.Outf("Already logged out.")
		return errors.New("already logged out")
	}
	b, err := ioutil.ReadFile(s)
	if err != nil {
		return err
	}
	log.Infof("Removing %s\n", s)
	if err := os.Remove(s); err != nil {
		return err
	}
	log.Infof("Successfully removed %s\n", s)
	return revokeToken(b)
}

var revokeToken = func(file []byte) error {
	type tokenFile struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	var out tokenFile
	if err := json.Unmarshal(file, &out); err != nil {
		return err
	}
	// Revokes an access token or if it's expired, revokes the refresh token
	// If the token has expired, been tampered with, or had its permissions revoked,
	// Google's authorization server returns an error message in the JSON object.
	// The error surfaces as a 400 error. Revoking an access token also revokes
	// a refresh token associated with it.
	// Reference: https://developers.google.com/youtube/v3/live/guides/auth/client-side-web-apps
	for i := 0; i < 2; i++ {
		var token string
		if i == 0 {
			token = out.AccessToken
		} else {
			token = out.RefreshToken
		}
		log.Infof("Attempt %v: revoking a token.\n", i)
		url := fmt.Sprintf("https://accounts.google.com/o/oauth2/revoke?token=%s", token)
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		if resp.StatusCode == 200 {
			log.Infof("Attempt %v: successfully revoked a token.\n", i)
			break
		}
	}
	return nil
}

// token retrieves OAuth2 token with the given OAuth2 config. It tries looking up in tokenCacheFilename, and
// if token is not found, will prompt the user to get an interactive code to exchange for OAuth2 token.
var token = func(ctx context.Context, config *oauth2.Config, tokenCacheFilename string, launchBrowser bool) (*oauth2.Token, error) {
	var tok *oauth2.Token
	var err error
	tok, err = tokenFromFile(tokenCacheFilename)
	if err == nil {
		return tok, nil
	}
	if launchBrowser {
		tok, err = interactiveTokenWeb(ctx, config)
	} else {
		tok, err = interactiveTokenCopyPaste(ctx, config)
	}
	return tok, err
}

// Checks if the shell is not SSH.
func checkShell() bool {
	// https://en.wikibooks.org/wiki/OpenSSH/Client_Applications
	return len(os.Getenv("SSH_CLIENT")) == 0
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.Unmarshal(b, t)
	if err != nil {
		return nil, err
	}
	return t, err
}

// interactiveToken gets OAuth2 token from an authorization code received from the user.
var interactiveTokenCopyPaste = func(ctx context.Context, conf *oauth2.Config) (*oauth2.Token, error) {
	requestURL := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	log.Outln("Gactions needs access to your Google account. Please copy & paste the URL below into a web browser and follow the instructions there. Then copy and paste the authorization code from the browser back here.")
	log.Outf("Visit this URL: \n%s\n", requestURL)
	log.Out("Enter authorization code: ")
	var code string
	_, err := fmt.Scan(&code)
	if err != nil {
		return nil, err
	}
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return tok, nil
}

// interactiveToken gets OAuth2 token from an authorization code received from the user.
var interactiveTokenWeb = func(ctx context.Context, configIn *oauth2.Config) (*oauth2.Token, error) {
	// Start server on localhost and let net pick the open port.
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, err
	}
	defer listener.Close()
	tcpAddr, err := net.ResolveTCPAddr("tcp", listener.Addr().String())
	if err != nil {
		return nil, err
	}
	redirectPath := "/oauth"
	redirectPort := tcpAddr.Port
	urlPrefix := fmt.Sprintf("http://localhost:%d", redirectPort)
	// Make a copy of the config and patch its RedirectURL member.
	config := *configIn
	config.RedirectURL = urlPrefix + redirectPath

	// Launch browser (note: this would not work in a SSH session).
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	var cmdName string
	switch runtime.GOOS {
	case "linux":
		cmdName = "xdg-open"
	case "darwin":
		cmdName = "open"
	default:
		return nil, fmt.Errorf("can not automatically open a browser on %v", runtime.GOOS)
	}
	cmd := exec.Command(cmdName, authURL)
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// Setup server handle functions.
	errCh := make(chan error)
	codes := make(chan string)
	http.HandleFunc(redirectPath, func(w http.ResponseWriter, request *http.Request) {
		query := request.URL.Query()
		type loginPromptData struct {
			H1 string
			H2 string
			P  string
		}
		var t *template.Template
		var errTemplate error
		t = template.Must(template.New("login").Parse(loginPrompt))
		s := ""
		buf := bytes.NewBufferString(s)
		if err := query.Get("error"); err != "" {
			errCh <- fmt.Errorf("OAuth error response: %v", err)
			errTemplate = t.Execute(buf, loginPromptData{
				H2: "Oops!",
				H1: "gactions CLI Login Failed",
				P:  "The gactions CLI login request was rejected or an error occurred. Please run gactions login again.",
			})
		} else if code := query.Get("code"); code == "" {
			errCh <- fmt.Errorf("OAuth error empty")
			errTemplate = t.Execute(buf, loginPromptData{
				H2: "Oops!",
				H1: "gactions CLI Login Failed",
				P:  "The gactions CLI login request was rejected or an error occurred. Please run gactions login again.",
			})
		} else {
			codes <- code
			errTemplate = t.Execute(buf, loginPromptData{
				H2: "Great!",
				H1: "gactions CLI Login Successful",
				P:  "You are logged in to the gactions Command-Line interface. You can immediately close this window and continue using the CLI.",
			})
		}
		if errTemplate != nil {
			fmt.Fprint(w, "<html><body><h1>gactions login failed. Please try again.</h1></body>")
		} else {
			fmt.Fprint(w, buf.String())
		}
	})

	// Start server, defer shutdown to end of function.
	server := http.Server{}
	go server.Serve(listener)

	// Have server running for only 1 minute and then stop.
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	defer server.Shutdown(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Obtain either code or error.
	select {
	case err = <-errCh:
		return nil, err
	case code := <-codes:
		log.Infoln("OAuth key code obtained.")
		return config.Exchange(ctx, code)
	case <-stop:
		return nil, errors.New("caught interrupt signal")
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			log.Infof("Deadline exceeded: %s", ctx.Err().Error())
			return nil, errors.New("waited for user input for too long")
		}
		return nil, errors.New("unable to retrieve OAuth key code")
	}
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) error {
	if exists(file) {
		return nil
	}
	log.Infof("Saving credential file to: %s\n", file)
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("unable to marshal token into json: %v", err)
	}
	return ioutil.WriteFile(file, tokenJSON, 0644)
}

// exists returns whether the given file or directory exists or not
func exists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return os.IsExist(err)
	}
	return true
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
var tokenCacheFile = func() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("gactions-actions.googleapis.com-go.json")), err
}
