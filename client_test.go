package main

import (
	"net/url"
	"testing"

	"github.com/google/go-github/v34/github"
	gitlab "github.com/xanzy/go-gitlab"
)

func TestNewClient(t *testing.T) {
	setup()
	defer teardown()

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

	// Not yet supported
	client = newClient("notyetsupported", "")
	if client != nil {
		t.Errorf("Expected nil")
	}

}
