package main

import (
	"net/url"
	"os"
	"testing"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
	"github.com/google/go-github/v34/github"
	"github.com/ktrysmt/go-bitbucket"
	gitlab "github.com/xanzy/go-gitlab"
)

func TestNewClient(t *testing.T) {
	setupRepositoryTests()
	defer teardownRepositoryTests()

	customGitHost, _ := url.Parse("https://git.mycompany.com")
	// http://stackoverflow.com/questions/23051339/how-to-avoid-end-of-url-slash-being-removed-when-resolvereference-in-go
	api, _ := url.Parse("api/v4/")
	expectedGitHostBaseURL := customGitHost.ResolveReference(api)

	// Client for github.com
	client := newClient("github", "")
	client = client.(*github.Client)

	// Client for Enterprise Github
	client = newClient("github", customGitHost.String())
	gotBaseURL := client.(*github.Client).BaseURL
	if gotBaseURL.String() != expectedGitHostBaseURL.String() {
		t.Errorf("Expected BaseURL to be: %v, Got: %v\n", expectedGitHostBaseURL, gotBaseURL)
	}

	// Client for gitlab.com
	client = newClient("gitlab", "")
	client = client.(*gitlab.Client)

	// Client for custom gitlab installation
	client = newClient("gitlab", customGitHost.String())
	gotBaseURL = client.(*gitlab.Client).BaseURL()
	if gotBaseURL.String() != expectedGitHostBaseURL.String() {
		t.Errorf("Expected BaseURL to be: %v, Got: %v\n", expectedGitHostBaseURL, gotBaseURL)
	}

	// Client for bitbucket.com
	client = newClient("bitbucket", "")
	client = client.(*bitbucket.Client)

	// Client for codeberg
	client = newClient("forgejo", "")
	client = client.(*forgejo.Client)

	// Client for forgejo
	client = newClient("forgejo", customGitHost.String())
	client = client.(*forgejo.Client)

	// Not yet supported
	client = newClient("notyetsupported", "")
	if client != nil {
		t.Errorf("Expected nil")
	}

}

func TestNewBitbucketClientWithToken(t *testing.T) {
	setupRepositoryTests()
	defer teardownRepositoryTests()

	// Set BITBUCKET_TOKEN and unset BITBUCKET_PASSWORD to test token auth path
	os.Setenv("BITBUCKET_TOKEN", "$$$randomtoken")
	os.Unsetenv("BITBUCKET_PASSWORD")
	defer os.Unsetenv("BITBUCKET_TOKEN")

	client := newClient("bitbucket", "")
	if client == nil {
		t.Fatal("Expected non-nil bitbucket client")
	}
	_ = client.(*bitbucket.Client)

	if gitHostToken != "$$$randomtoken" {
		t.Errorf("Expected gitHostToken to be BITBUCKET_TOKEN value, got: %v", gitHostToken)
	}
}
