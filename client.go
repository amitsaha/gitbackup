package main

import (
	"context"
	"fmt"
	"log"
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

func newClient(service string, gitHostURLParsed *url.URL) interface{} {
	var err error

	switch service {
	case "github":
		githubToken := os.Getenv("GITHUB_TOKEN")
		if githubToken == "" {
			githubToken, err = getToken("GITHUB")
			if err != nil {
				githubToken = startOAuthFlow()
			}
			if githubToken == "" {
				log.Fatal("GitHub token not available")
			}
			err := saveToken("GITHUB", githubToken)
			if err != nil {
				log.Fatal("Error saving token")
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
		return client

	case "gitlab":
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

	case "bitbucket":
		bitbucketUsername := os.Getenv("BITBUCKET_USERNAME")
		bitbucketPassword := os.Getenv("BITBUCKET_PASSWORD")
		if bitbucketUsername == "" || bitbucketPassword == "" {
			log.Fatal("BITBUCKET_USERNAME and BITBUCKET_PASSWORD environment variables must be set")
		}
		gitHostToken = bitbucketPassword
		client := bitbucket.NewBasicAuth(bitbucketUsername, bitbucketPassword)
		if gitHostURLParsed != nil {
			client.SetApiBaseURL(gitHostURLParsed.String())
		}
		return client
	default:
		return nil
	}
}
