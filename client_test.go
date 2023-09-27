package main

import (
	"net/url"
	"testing"

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
	client := newClient("github", nil)
	client = client.(*github.Client)

	// Client for Enterprise Github
	client = newClient("github", customGitHost)
	gotBaseURL := client.(*github.Client).BaseURL
	if gotBaseURL.String() != expectedGitHostBaseURL.String() {
		t.Errorf("Expected BaseURL to be: %v, Got: %v\n", expectedGitHostBaseURL, gotBaseURL)
	}

	// Client for gitlab.com
	client = newClient("gitlab", nil)
	client = client.(*gitlab.Client)

	// Client for custom gitlab installation
	client = newClient("gitlab", customGitHost)
	gotBaseURL = client.(*gitlab.Client).BaseURL()
	if gotBaseURL.String() != expectedGitHostBaseURL.String() {
		t.Errorf("Expected BaseURL to be: %v, Got: %v\n", expectedGitHostBaseURL, gotBaseURL)
	}

	// Client for bitbucket.com
	client = newClient("bitbucket", nil)
	client = client.(*bitbucket.Client)

	// Not yet supported
	client = newClient("notyetsupported", nil)
	if client != nil {
		t.Errorf("Expected nil")
	}

}
