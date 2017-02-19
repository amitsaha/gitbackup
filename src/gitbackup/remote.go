package main

import (
	"github.com/google/go-github/github"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
)

// https://github.com/google/go-github/blob/27c7c32b6d369610435bd2ad7b4d8554f235eb01/github/github.go#L301
// https://github.com/xanzy/go-gitlab/blob/3acf8d75e9de17ad4b41839a7cabbf2537760ab4/gitlab.go#L286
type Response struct {
	*http.Response

	// These fields provide the page values for paginating through a set of
	// results.  Any or all of these may be set to the zero value for
	// responses that are not part of a paginated set, or for which there
	// are no additional pages.

	NextPage  int
	PrevPage  int
	FirstPage int
	LastPage  int
}

type Repository struct {
	GitURL string
	Name   string
}

func NewClient(httpClient *http.Client, service string) interface{} {
	if service == "github" {
		githubToken := os.Getenv("GITHUB_TOKEN")
		if githubToken == "" {
			log.Fatal("GITHUB_TOKEN environment variable not set")
		}
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: githubToken},
		)
		tc := oauth2.NewClient(oauth2.NoContext, ts)
		return github.NewClient(tc)
	}

	if service == "gitlab" {
		gitlabToken := os.Getenv("GITLAB_TOKEN")
		if gitlabToken == "" {
			log.Fatal("GITLAB_TOKEN environment variable not set")
		}
		return gitlab.NewClient(httpClient, gitlabToken)
	}
	return nil
}

func getRepositories(service string, gitlabUrl string, githubRepoType string, gitlabRepoVisibility string) ([]*Repository, error) {

	client := NewClient(nil, service)
	if client == nil {
		log.Fatal("Couldn't acquire a client to talk to %s", service)
	}

	var repositories []*Repository

	if service == "github" {
		options := github.RepositoryListOptions{Type: githubRepoType}
		for {
			repos, resp, err := client.(*github.Client).Repositories.List("", &options)
			if err == nil {
				for _, repo := range repos {
					repositories = append(repositories, &Repository{GitURL: *repo.GitURL, Name: *repo.Name})
				}
			} else {
				return nil, err
			}
			if resp.NextPage == 0 {
				break
			}
			options.ListOptions.Page = resp.NextPage
		}
	}

	if service == "gitlab" {
		options := gitlab.ListProjectsOptions{Visibility: &gitlabRepoVisibility}
		if len(gitlabUrl) != 0 {
			gitlabUrlPath, err := url.Parse(gitlabUrl)
			if err != nil {
				log.Fatal("Invalid gitlab URL: %s", gitlabUrl)
			}
			gitlabUrlPath.Path = path.Join(gitlabUrlPath.Path, "api/v3")
			client.(*gitlab.Client).SetBaseURL(gitlabUrlPath.String())
		}
		for {
			repos, resp, err := client.(*gitlab.Client).Projects.ListProjects(&options)
			if err == nil {
				for _, repo := range repos {
					repositories = append(repositories, &Repository{GitURL: repo.SSHURLToRepo, Name: repo.Name})
				}
			} else {
				return nil, err
			}
			if resp.NextPage == 0 {
				break
			}
			options.ListOptions.Page = resp.NextPage
		}
	}
	return repositories, nil
}
