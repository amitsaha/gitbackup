package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/oauth2"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
	"github.com/google/go-github/v34/github"
	bitbucket "github.com/ktrysmt/go-bitbucket"
	gitlab "github.com/xanzy/go-gitlab"

	"github.com/99designs/keyring"
	"github.com/cli/oauth/device"
)

var keyringServiceName = "gitbackup-cli"
var gitbackupClientID = "7b56a77c7dfba0800524"

func startOAuthFlow() string {
	clientID := gitbackupClientID
	scopes := []string{"repo", "user", "admin:org"}
	httpClient := http.DefaultClient

	code, err := device.RequestCode(httpClient, "https://github.com/login/device/code", clientID, scopes)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Copy code: %s\n", code.UserCode)
	fmt.Printf("then open: %s\n", code.VerificationURI)

	accessToken, err := device.PollToken(httpClient, "https://github.com/login/oauth/access_token", clientID, code)
	if err != nil {
		panic(err)
	}

	return accessToken.Token
}

func saveToken(service string, token string) error {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: keyringServiceName,
	})
	if err != nil {
		return err
	}

	err = ring.Set(keyring.Item{
		Key:  service + "_TOKEN",
		Data: []byte(token),
	})
	return err
}

func getToken(service string) (string, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: keyringServiceName,
	})
	if err != nil {
		return "", err
	}
	i, err := ring.Get(service + "_TOKEN")
	if err != nil {
		return "", err
	}
	return string(i.Data), nil
}

func newClient(service string, gitHostURL string) interface{} {
	gitHostURLParsed := parseGitHostURL(gitHostURL, service)

	switch service {
	case "github":
		return newGitHubClient(gitHostURLParsed)
	case "gitlab":
		return newGitLabClient(gitHostURLParsed)
	case "bitbucket":
		return newBitbucketClient(gitHostURLParsed)
	case "forgejo":
		return newForgejoClient(gitHostURLParsed)
	default:
		return nil
	}
}

// parseGitHostURL parses the git host URL if provided
// TODO: there is a chance, this parsing breaks more than
// one git service
// https://github.com/amitsaha/gitbackup/issues/195
func parseGitHostURL(gitHostURL string, service string) *url.URL {
	if len(gitHostURL) == 0 {
		return nil
	}

	gitHostURLParsed, err := url.Parse(gitHostURL)
	if err != nil {
		log.Fatalf("Invalid git host URL: %s", gitHostURL)
	}

	// temp fix for https://github.com/amitsaha/gitbackup/issues/193
	if service == "forgejo" {
		return gitHostURLParsed
	}
	api, _ := url.Parse("api/v4/")
	return gitHostURLParsed.ResolveReference(api)
}

// newGitHubClient creates a new GitHub client
func newGitHubClient(gitHostURLParsed *url.URL) *github.Client {
	githubToken := getOrCreateGitHubToken()
	gitHostToken = githubToken

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	if gitHostURLParsed != nil {
		client.BaseURL = gitHostURLParsed
	}
	return client
}

// getOrCreateGitHubToken retrieves or creates a GitHub token
func getOrCreateGitHubToken() string {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken != "" {
		return githubToken
	}

	githubToken, err := getToken("GITHUB")
	if err != nil {
		githubToken = startOAuthFlow()
	}

	if githubToken == "" {
		log.Fatal("GitHub token not available")
	}

	err = saveToken("GITHUB", githubToken)
	if err != nil {
		log.Fatal("Error saving token")
	}

	return githubToken
}

// newGitLabClient creates a new GitLab client
func newGitLabClient(gitHostURLParsed *url.URL) *gitlab.Client {
	gitlabToken := os.Getenv("GITLAB_TOKEN")
	if gitlabToken == "" {
		log.Fatal("GITLAB_TOKEN environment variable not set")
	}
	gitHostToken = gitlabToken

	var baseUrlOption gitlab.ClientOptionFunc
	if gitHostURLParsed != nil {
		baseUrlOption = gitlab.WithBaseURL(gitHostURLParsed.String())
	}

	client, err := gitlab.NewClient(gitlabToken, baseUrlOption)
	if err != nil {
		log.Fatalf("Error creating gitlab client: %v", err)
	}
	return client
}

// newBitbucketClient creates a new Bitbucket client
func newBitbucketClient(gitHostURLParsed *url.URL) *bitbucket.Client {
	bitbucketUsername := os.Getenv("BITBUCKET_USERNAME")
	if bitbucketUsername == "" {
		log.Fatal("BITBUCKET_USERNAME environment variable not set")
	}

	bitbucketPassword := os.Getenv("BITBUCKET_TOKEN")
	if bitbucketPassword == "" {
		bitbucketPassword = os.Getenv("BITBUCKET_PASSWORD")
	}
	if bitbucketPassword == "" {
		log.Fatal("BITBUCKET_TOKEN or BITBUCKET_PASSWORD environment variable must be set")
	}

	gitHostToken = bitbucketPassword
	client := bitbucket.NewBasicAuth(bitbucketUsername, bitbucketPassword)

	if gitHostURLParsed != nil {
		client.SetApiBaseURL(*gitHostURLParsed)
	}
	return client
}

// newForgejoClient creates a new Forgejo client.
func newForgejoClient(gitHostURLParsed *url.URL) *forgejo.Client {
	forgejoToken := os.Getenv("FORGEJO_TOKEN")
	if forgejoToken == "" {
		log.Fatal("FORGEJO_TOKEN environment variable not set")
	}

	url := "https://" + knownServices["forgejo"]
	if gitHostURLParsed != nil {
		url = gitHostURLParsed.String()
	}

	log.Println("Creating forgejo client", url)
	client, err := forgejo.NewClient(url, forgejo.SetToken(forgejoToken), forgejo.SetForgejoVersion(""))
	if err != nil {
		log.Fatalf("Error creating forgejo client: %v", err)
	}

	return client
}
