// Copyright 2016 Google Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to writing, software distributed
// under the License is distributed on a "AS IS" BASIS, WITHOUT WARRANTIES OR
// CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.

package embedmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Fetcher provides an abstraction on a file system.
// The Fetch function is called anytime some content needs to be fetched.
// For now this includes files and URLs.
// The first parameter is the base directory that could be used to resolve
// relative paths. This base directory will be ignored for absolute paths,
// such as URLs.
type Fetcher interface {
	Fetch(dir, path string) ([]byte, error)
}

// fetcher implements the Fetcher interface with an injectable HTTP client.
type fetcher struct {
	client *http.Client
}

// NewFetcher creates a new fetcher with the provided HTTP client.
// If no client is provided, it defaults to http.DefaultClient.
func NewFetcher(client *http.Client) Fetcher {
	if client == nil {
		client = http.DefaultClient
	}
	return &fetcher{client: client}
}

// Fetch fetches the content of a file or URL.
func (f *fetcher) Fetch(dir, path string) ([]byte, error) {
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		// Check that path is not absolute
		if !filepath.IsAbs(path) {
			path = filepath.Join(dir, filepath.FromSlash(path))
		}
		return os.ReadFile(path)
	}

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	if val, ok := os.LookupEnv("GITHUB_TOKEN"); ok {
		req.Header.Add("Authorization", "Bearer "+val)
	}

	res, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %s", res.Status)
	}
	return io.ReadAll(res.Body)
}
