package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/oauth2"

	"github.com/google/go-github/v34/github"
	bitbucket "github.com/ktrysmt/go-bitbucket"
	gitlab "github.com/xanzy/go-gitlab"

	"github.com/99designs/keyring"
	"github.com/cli/oauth/device"
)

var keyringServiceName = "gitbackup-cli"
var gitbackupClientID = "7b56a77c7dfba0800524"

func startOAuthFlow() (*string, error) {
	clientID := gitbackupClientID
	scopes := []string{"repo", "user", "admin:org"}
	httpClient := http.DefaultClient

	code, err := device.RequestCode(httpClient, "https://github.com/login/device/code", clientID, scopes)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Copy code: %s\n", code.UserCode)
	fmt.Printf("then open: %s\n", code.VerificationURI)

	accessToken, err := device.PollToken(httpClient, "https://github.com/login/oauth/access_token", clientID, code)
	if err != nil {
		return nil, err
	}

	return &accessToken.Token, nil
}

func saveToken(service string, token *string) error {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: keyringServiceName,
	})
	if err != nil {
		return err
	}

	err = ring.Set(keyring.Item{
		Key:  service + "_TOKEN",
		Data: []byte(*token),
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

func newClient(c *appConfig) (interface{}, error) {
	var err error
	var gitHostURLParsed *url.URL

	if c != nil && len(c.gitHostURL) != 0 {
		gitHostURLParsed, err = url.Parse(c.gitHostURL)
		if err != nil {
			return nil, err
		}
	}

	switch c.service {
	case "github":
		githubToken := os.Getenv("GITHUB_TOKEN")
		if githubToken == "" {
			githubToken, err = getToken("GITHUB")
			if err != nil {
				githubToken, err := startOAuthFlow()
				if err != nil {
					return nil, err
				}
				if githubToken == nil {
					return nil, fmt.Errorf("GitHub token not available")
				}
				err = saveToken("GITHUB", githubToken)
				if err != nil {
					return nil, fmt.Errorf("Error saving token")
				}
			}
		}
		gitHostToken = githubToken
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: githubToken},
		)
		tc := oauth2.NewClient(context.Background(), ts)
		client := github.NewClient(tc)
		if gitHostURLParsed != nil {
			client.BaseURL = gitHostURLParsed
		}
		return client, nil

	case "gitlab":

		if gitHostURLParsed != nil {
			api, _ := url.Parse("api/v4/")
			gitHostURLParsed = gitHostURLParsed.ResolveReference(api)
		}
		gitlabToken := os.Getenv("GITLAB_TOKEN")
		if gitlabToken == "" {
			return nil, fmt.Errorf("GITLAB_TOKEN environment variable not set")
		}
		gitHostToken = gitlabToken

		var baseUrlOption gitlab.ClientOptionFunc
		if gitHostURLParsed != nil {
			baseUrlOption = gitlab.WithBaseURL(gitHostURLParsed.String())
		}
		client, err := gitlab.NewClient(gitlabToken, baseUrlOption)
		if err != nil {
			return nil, fmt.Errorf("Error creating gitlab client: %v", err)
		}
		return client, nil

	case "bitbucket":
		bitbucketUsername := os.Getenv("BITBUCKET_USERNAME")
		bitbucketPassword := os.Getenv("BITBUCKET_PASSWORD")
		if bitbucketUsername == "" || bitbucketPassword == "" {
			return nil, fmt.Errorf("BITBUCKET_USERNAME and BITBUCKET_PASSWORD environment variables must be set")
		}
		gitHostToken = bitbucketPassword
		client := bitbucket.NewBasicAuth(bitbucketUsername, bitbucketPassword)
		if gitHostURLParsed != nil {
			client.SetApiBaseURL(gitHostURLParsed.String())
		}
		return client, nil
	default:
		return nil, errors.New("invalid service")
	}
}
