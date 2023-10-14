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

	defaultGithubConfig := appConfig{
		service:    "github",
		gitHostURL: "https://github.com",
	}
	customGithubConfig := appConfig{
		service:    "github",
		gitHostURL: "https://git.mycompany.com",
	}

	defaultGitlabConfig := appConfig{
		service:    "gitlab",
		gitHostURL: "https://gitlab.com",
	}
	customGitlabConfig := appConfig{
		service:    "gitlab",
		gitHostURL: "https://git.mycompany.com",
	}

	defaultBitbucketConfig := appConfig{
		service:    "bitbucket",
		gitHostURL: "https://bitbucket.org",
	}

	// Client for github.com
	client := newClient(&defaultGithubConfig)
	client = client.(*github.Client)

	// Client for Enterprise Github
	client = newClient(&customGithubConfig)
	gotBaseURL := client.(*github.Client).BaseURL
	if gotBaseURL.String() != customGithubConfig.gitHostURL {
		t.Errorf("Expected BaseURL to be: %v, Got: %v\n", customGithubConfig.gitHostURL, gotBaseURL)
	}

	// Client for gitlab.com
	client = newClient(&defaultGitlabConfig)
	client = client.(*gitlab.Client)

	// http://stackoverflow.com/questions/23051339/how-to-avoid-end-of-url-slash-being-removed-when-resolvereference-in-go
	api, _ := url.Parse("api/v4/")
	customGitHostParsed, _ := url.Parse(customGitlabConfig.gitHostURL)
	expectedGitHostBaseURL := customGitHostParsed.ResolveReference(api)

	// Client for custom gitlab installation
	client = newClient(&customGitlabConfig)
	gotBaseURL = client.(*gitlab.Client).BaseURL()
	if gotBaseURL.String() != expectedGitHostBaseURL.String() {
		t.Errorf("Expected BaseURL to be: %v, Got: %v\n", expectedGitHostBaseURL, gotBaseURL)
	}

	// Client for bitbucket.com
	client = newClient(&defaultBitbucketConfig)
	client = client.(*bitbucket.Client)

	// Not yet supported
	unsupportedServiceConfig := appConfig{
		service: "notyetsupported",
	}
	client = newClient(&unsupportedServiceConfig)
	if client != nil {
		t.Errorf("Expected nil")
	}

}
