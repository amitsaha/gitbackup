package main

import (
	"github.com/google/go-github/github"
	"github.com/xanzy/go-gitlab"
	"net/url"
	"testing"
)

func TestNewClient(t *testing.T) {
	setup()
	defer teardown()

	customGitHost, _ := url.Parse("https://git.mycompany.com")
	// http://stackoverflow.com/questions/23051339/how-to-avoid-end-of-url-slash-being-removed-when-resolvereference-in-go
	api, _ := url.Parse("api/v3/")
	expectedGitHostBaseURL := customGitHost.ResolveReference(api)

	// Client for github.com
	client := NewClient("github", "")
	client = client.(*github.Client)

	// Client for Enterprise Github
	client = NewClient("github", customGitHost.String())
	gotBaseURL := client.(*github.Client).BaseURL
	if gotBaseURL.String() != expectedGitHostBaseURL.String() {
		t.Errorf("Expected BaseURL to be: %v, Got: %v\n", expectedGitHostBaseURL, gotBaseURL)
	}

	// Client for gitlab.com
	client = NewClient("gitlab", "")
	client = client.(*gitlab.Client)

	// Client for custom gitlab installation
	client = NewClient("gitlab", customGitHost.String())
	gotBaseURL = client.(*gitlab.Client).BaseURL()
	if gotBaseURL.String() != expectedGitHostBaseURL.String() {
		t.Errorf("Expected BaseURL to be: %v, Got: %v\n", expectedGitHostBaseURL, gotBaseURL)
	}

	// Not yet supported
	client = NewClient("notyetsupported", "")
	if client != nil {
		t.Errorf("Expected nil")
	}

}
