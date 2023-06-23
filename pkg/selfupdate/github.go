// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package selfupdate

// derived from http://github.com/restic/restic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"
)

var (
	errUnexpectedResponse = errors.New("unexpected response")
	errEmptyTag           = errors.New("tag name for latest release is empty")
	errInvalidTag         = errors.New("invalid tag name")
)

// Release collects data about a single release on GitHub.
type Release struct {
	Name        string    `json:"name"`
	TagName     string    `json:"tag_name"`
	Draft       bool      `json:"draft"`
	PreRelease  bool      `json:"prerelease"` //nolint:tagliatelle
	PublishedAt time.Time `json:"published_at"`
	Assets      []Asset   `json:"assets"`

	Version string `json:"-"` // set manually in the code
}

// Asset is a file uploaded and attached to a release.
type Asset struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

func (r Release) String() string {
	return fmt.Sprintf("%v %v, %d assets",
		r.TagName,
		r.PublishedAt.Local().Format("2006-01-02 15:04:05"),
		len(r.Assets))
}

const githubAPITimeout = 30 * time.Second

// githubError is returned by the GitHub API, e.g. for rate-limiting.
type githubError struct {
	Message string
}

// GitHubLatestRelease uses the GitHub API to get information about the latest
// release of a repository.
func GitHubLatestRelease(ctx context.Context) (*Release, error) {
	ctx, cancel := context.WithTimeout(ctx, githubAPITimeout)
	defer cancel()

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", githubUser, githubRepo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}

	// pin API version 3
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		content := res.Header.Get("Content-Type")
		if strings.Contains(content, "application/json") {
			// try to decode error message
			var msg githubError
			if err := json.NewDecoder(res.Body).Decode(&msg); err == nil {
				return nil, fmt.Errorf("%w %v (%v) returned, message: %v", errUnexpectedResponse, res.StatusCode, res.Status, msg.Message)
			}
		}

		_ = res.Body.Close()
		return nil, fmt.Errorf("%w %v (%v) returned", errUnexpectedResponse, res.StatusCode, res.Status)
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		_ = res.Body.Close()
		return nil, err
	}

	if err = res.Body.Close(); err != nil {
		return nil, err
	}

	release := &Release{}
	if err = json.Unmarshal(buf, release); err != nil {
		return nil, err
	}

	if release.TagName == "" {
		return nil, errEmptyTag
	}

	if !strings.HasPrefix(release.TagName, "v") {
		return nil, fmt.Errorf("%w: %s, does not start with 'v'", errInvalidTag, release.TagName)
	}

	release.Version = release.TagName[1:]

	return release, nil
}

func getGithubData(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Request binary data
	req.Header.Set("Accept", "application/octet-stream")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w %v (%v) returned", errUnexpectedResponse, res.StatusCode, res.Status)
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		_ = res.Body.Close()
		return nil, err
	}

	if err = res.Body.Close(); err != nil {
		return nil, err
	}

	return buf, nil
}

func getGithubDataFile(ctx context.Context, assets []Asset, suffix string) (filename string, data []byte, err error) {
	var url string
	for _, a := range assets {
		if strings.HasSuffix(a.Name, suffix) {
			url = a.URL
			filename = a.Name
			break
		}
	}

	if url == "" {
		return "", nil, fmt.Errorf("%w: unable to find file with suffix %v", syscall.ENOENT, suffix)
	}

	data, err = getGithubData(ctx, url)
	if err != nil {
		return "", nil, err
	}

	return filename, data, nil
}
