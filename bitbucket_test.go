package main

import (
	"reflect"
	"testing"
)

func TestExtractBitbucketCloneURLs(t *testing.T) {
	tests := []struct {
		name         string
		links        map[string]interface{}
		wantHTTPSURL string
		wantSSHURL   string
	}{
		{
			name: "valid links with both HTTPS and SSH",
			links: map[string]interface{}{
				"clone": []interface{}{
					map[string]interface{}{
						"name": "https",
						"href": "https://bitbucket.org/user/repo.git",
					},
					map[string]interface{}{
						"name": "ssh",
						"href": "git@bitbucket.org:user/repo.git",
					},
				},
			},
			wantHTTPSURL: "https://bitbucket.org/user/repo.git",
			wantSSHURL:   "git@bitbucket.org:user/repo.git",
		},
		{
			name: "empty links",
			links: map[string]interface{}{
				"clone": []interface{}{},
			},
			wantHTTPSURL: "",
			wantSSHURL:   "",
		},
		{
			name:         "missing clone key",
			links:        map[string]interface{}{},
			wantHTTPSURL: "",
			wantSSHURL:   "",
		},
		{
			name: "only HTTPS URL",
			links: map[string]interface{}{
				"clone": []interface{}{
					map[string]interface{}{
						"name": "https",
						"href": "https://bitbucket.org/user/repo.git",
					},
				},
			},
			wantHTTPSURL: "https://bitbucket.org/user/repo.git",
			wantSSHURL:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHTTPSURL, gotSSHURL := extractBitbucketCloneURLs(tt.links)
			if gotHTTPSURL != tt.wantHTTPSURL {
				t.Errorf("extractBitbucketCloneURLs() httpsURL = %v, want %v", gotHTTPSURL, tt.wantHTTPSURL)
			}
			if gotSSHURL != tt.wantSSHURL {
				t.Errorf("extractBitbucketCloneURLs() sshURL = %v, want %v", gotSSHURL, tt.wantSSHURL)
			}
		})
	}
}

func TestBuildRepoPaths(t *testing.T) {
	repos := []*Repository{
		{Name: "repo1", Namespace: "user1"},
		{Name: "repo2", Namespace: "org1"},
		{Name: "repo3", Namespace: "user1"},
	}

	expected := []string{"user1/repo1", "org1/repo2", "user1/repo3"}
	result := buildRepoPaths(repos)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("buildRepoPaths() = %v, want %v", result, expected)
	}
}

func TestBuildRepoPathsEmpty(t *testing.T) {
	repos := []*Repository{}
	result := buildRepoPaths(repos)

	if len(result) != 0 {
		t.Errorf("buildRepoPaths() for empty input = %v, want empty slice", result)
	}
}
