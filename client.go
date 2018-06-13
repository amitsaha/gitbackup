package main

import (
	"log"
	"net/url"
	"os"

	"github.com/google/go-github/github"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

func NewClient(service string, gitHostUrl string) interface{} {
	var gitHostUrlParsed *url.URL
	var err error

	// If a git host URL has been passed in, we assume it's
	// a gitlab installation
	if len(gitHostUrl) != 0 {
		gitHostUrlParsed, err = url.Parse(gitHostUrl)
		if err != nil {
			log.Fatalf("Invalid gitlab URL: %s", gitHostUrl)
		}
		api, _ := url.Parse("api/v4/")
		gitHostUrlParsed = gitHostUrlParsed.ResolveReference(api)
	}

	if service == "github" {
		githubToken := os.Getenv("GITHUB_TOKEN")
		if githubToken == "" {
			log.Fatal("GITHUB_TOKEN environment variable not set")
		}
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: githubToken},
		)
		tc := oauth2.NewClient(oauth2.NoContext, ts)
		client := github.NewClient(tc)
		if gitHostUrlParsed != nil {
			client.BaseURL = gitHostUrlParsed
		}
		return client
	}

	if service == "gitlab" {
		gitlabToken := os.Getenv("GITLAB_TOKEN")
		if gitlabToken == "" {
			log.Fatal("GITLAB_TOKEN environment variable not set")
		}
		client := gitlab.NewClient(nil, gitlabToken)
		if gitHostUrlParsed != nil {
			client.SetBaseURL(gitHostUrlParsed.String())
		}
		return client
	}
	return nil
}
