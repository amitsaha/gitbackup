package main

import (
	"net/url"
	"testing"

	"strings"

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
	client, err := newClient(&defaultGithubConfig)
	if err != nil {
		t.Fatal(err)
	}
	client = client.(*github.Client)

	// Client for Enterprise Github
	client, err = newClient(&customGithubConfig)
	if err != nil {
		t.Fatal(err)
	}

	gotBaseURL := client.(*github.Client).BaseURL
	if gotBaseURL.String() != customGithubConfig.gitHostURL {
		t.Errorf("Expected BaseURL to be: %v, Got: %v\n", customGithubConfig.gitHostURL, gotBaseURL)
	}

	// Client for gitlab.com
	client, err = newClient(&defaultGitlabConfig)
	if err != nil {
		t.Fatal(err)
	}

	client = client.(*gitlab.Client)

	// http://stackoverflow.com/questions/23051339/how-to-avoid-end-of-url-slash-being-removed-when-resolvereference-in-go
	api, _ := url.Parse("api/v4/")
	customGitHostParsed, _ := url.Parse(customGitlabConfig.gitHostURL)
	expectedGitHostBaseURL := customGitHostParsed.ResolveReference(api)

	// Client for custom gitlab installation
	client, err = newClient(&customGitlabConfig)
	if err != nil {
		t.Fatal(err)
	}

	gotBaseURL = client.(*gitlab.Client).BaseURL()
	if gotBaseURL.String() != expectedGitHostBaseURL.String() {
		t.Errorf("Expected BaseURL to be: %v, Got: %v\n", expectedGitHostBaseURL, gotBaseURL)
	}

	// Client for bitbucket.org
	client, err = newClient(&defaultBitbucketConfig)
	if err != nil {
		t.Fatal(err)
	}

	client = client.(*bitbucket.Client)

	// Not yet supported
	unsupportedServiceConfig := appConfig{
		service: "notyetsupported",
	}
	client, err = newClient(&unsupportedServiceConfig)
	if !strings.Contains(err.Error(), "invalid service") {
		t.Fatal(err)
	}
}
