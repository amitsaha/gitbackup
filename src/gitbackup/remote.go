package main

import (
	"errors"
	"github.com/google/go-github/github"
	"github.com/xanzy/go-gitlab"
	"log"
	"net/http"
	"os"
)

type ListRepositoriesOptions struct {
	repoType    string
	Sort        string
	Direction   string
	ListOptions github.ListOptions
}

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
		return github.NewClient(httpClient)
	}

	if service == "gitlab" {
		gitLabToken := os.Getenv("GITLAB_TOKEN")
		if gitLabToken == "" {
			log.Fatal("GITLAB_TOKEN environment variable not set")
		}
		return gitlab.NewClient(httpClient, gitLabToken)
	}
	return nil
}

func getRepositories(service string, client interface{}, username string, opt *ListRepositoriesOptions) ([]*Repository, *Response, error) {

	if service == "github" {
		options := github.RepositoryListOptions{Type: opt.repoType, Sort: opt.Sort, Direction: opt.Direction, ListOptions: opt.ListOptions}
		repos, resp, err := client.(*github.Client).Repositories.List(username, &options)
		var repositories []*Repository
		if err == nil {
			for _, repo := range repos {
				repositories = append(repositories, &Repository{GitURL: *repo.GitURL, Name: *repo.Name})
			}
			return repositories, &Response{NextPage: resp.NextPage}, nil
		} else {
			return nil, &Response{Response: resp.Response}, err
		}
	}

	if service == "gitlab" {
		// TODO: other configuration options
		// https://docs.gitlab.com/ce/api/projects.html#list-projects
		//v := "internal"
		options := gitlab.ListProjectsOptions{} //{Visibility: *v}
		repos, resp, err := client.(*gitlab.Client).Projects.ListProjects(&options)
		var repositories []*Repository
		if err == nil {
			for _, repo := range repos {
				repositories = append(repositories, &Repository{GitURL: repo.SSHURLToRepo, Name: repo.Name})
			}
			return repositories, &Response{NextPage: resp.NextPage}, nil
		} else {
			return nil, &Response{Response: resp.Response}, err
		}
	}

	// TODO: fix
	return nil, nil, errors.New("Something unexpected happened")
}
