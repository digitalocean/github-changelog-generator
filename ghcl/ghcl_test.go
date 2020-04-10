/*
Copyright 2019 - 2020 DigitalOcean
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ghcl_test

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/digitalocean/github-changelog-generator/ghcl"
)

// TestBuild tests against actual GitHub repos using your network connection.
// Use the `-short` flag when invoking the test suite to skip these.
func TestBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests")
	}

	t.Run("doctl entries should not error", func(t *testing.T) {
		doChangelogService := ghcl.NewGitHubChangelogService("digitalocean",
			"doctl", os.Getenv("GITHUB_TOKEN"), "")
		if err := ghcl.Build(doChangelogService); err != nil {
			t.Fatal(err)
		}
	})
}

var _ ghcl.ChangelogService = &mockChangelogService{}

type mockChangelogService struct {
	t       time.Time
	entries []*ghcl.ChangelogEntry
}

func (mcs *mockChangelogService) FetchReleaseTime() (time.Time, error) {
	return mcs.t, nil
}

func (mcs *mockChangelogService) FetchChangelogEntriesUntil(t time.Time) ([]*ghcl.ChangelogEntry, error) {
	entries := make([]*ghcl.ChangelogEntry, 0)
	for _, entry := range mcs.entries {
		if &entry.MergedAt != nil {
			if entry.MergedAt.After(t) {
				entries = append(entries, entry)
			}
		}
	}
	return entries, nil
}

func TestGHCL(t *testing.T) {
	mcs := &mockChangelogService{
		t: time.Now().Add(-60 * time.Minute),
		entries: []*ghcl.ChangelogEntry{
			&ghcl.ChangelogEntry{
				Number:   2,
				Body:     "this should be in the changelog",
				Username: "digitalocean",
				MergedAt: time.Now().Add(-30 * time.Minute),
			},
			&ghcl.ChangelogEntry{
				Number:   1,
				Body:     "this should NOT be in the changelog",
				Username: "digitalocean",
				MergedAt: time.Now().Add(-90 * time.Minute),
			},
		},
	}

	entries, err := ghcl.FetchChangelogEntries(mcs)
	if err != nil {
		t.Fatalf("unexpected error getting changelog entries: %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("expeted at least one changelog entry, but got %d", len(entries))
	}
	if !reflect.DeepEqual(mcs.entries[0], entries[0]) {
		t.Fatalf("expected changelog to contain %+v but got %+v", mcs.entries[0], entries[0])
	}
}

func TestNewChangelogService(t *testing.T) {
	t.Run("passing an empty URL string should not panic", func(t *testing.T) {
		ghcl.NewGitHubChangelogService("digitalocean", "github-changelog-generator", "secret", "")
	})

	t.Run("passing the public GitHub API string should not panic", func(t *testing.T) {
		ghcl.NewGitHubChangelogService("digitalocean", "github-changelog-generator", "secret", "https://api.github.com/")
	})

	t.Run("passing a valid enterprise URL should not panic", func(t *testing.T) {
		ghcl.NewGitHubChangelogService("digitalocean", "github-changelog-generator", "secret", "https://github.enterprise.example.com/api/v3/")
	})
}
