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

	var defaultGithubConfig appConfig
	defaultGithubConfig.gitHostURL = "https://github.com"

	var defaultGitlabConfig appConfig
	defaultGitlabConfig.gitHostURL = "https://gitlab.com"

	var defaultBitbucketConfig appConfig
	defaultBitbucketConfig.gitHostURL = "https://bitbucket.org"

	var customGithostConfig appConfig
	customGitHost := "https://git.mycompany.com"
	customGithostConfig.gitHostURL = customGitHost

	// http://stackoverflow.com/questions/23051339/how-to-avoid-end-of-url-slash-being-removed-when-resolvereference-in-go
	api, _ := url.Parse("api/v4/")
	customGitHostParsed, _ := url.Parse(customGitHost)
	expectedGitHostBaseURL := customGitHostParsed.ResolveReference(api)

	// Client for github.com
	client := newClient("github", &defaultGithubConfig)
	client = client.(*github.Client)

	// Client for Enterprise Github
	client = newClient("github", &customGithostConfig)
	gotBaseURL := client.(*github.Client).BaseURL
	if gotBaseURL.String() != customGitHostParsed.String() {
		t.Errorf("Expected BaseURL to be: %v, Got: %v\n", expectedGitHostBaseURL, gotBaseURL)
	}

	// Client for gitlab.com
	client = newClient("gitlab", &defaultGitlabConfig)
	client = client.(*gitlab.Client)

	// Client for custom gitlab installation
	client = newClient("gitlab", &customGithostConfig)
	gotBaseURL = client.(*gitlab.Client).BaseURL()
	if gotBaseURL.String() != expectedGitHostBaseURL.String() {
		t.Errorf("Expected BaseURL to be: %v, Got: %v\n", expectedGitHostBaseURL, gotBaseURL)
	}

	// Client for bitbucket.com
	client = newClient("bitbucket", &defaultBitbucketConfig)
	client = client.(*bitbucket.Client)

	// Not yet supported
	client = newClient("notyetsupported", nil)
	if client != nil {
		t.Errorf("Expected nil")
	}

}
