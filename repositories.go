package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/google/go-github/github"
	"github.com/xanzy/go-gitlab"
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
	GitURL    string
	Name      string
	Namespace string
}

func getRepositories(client interface{}, service string, githubRepoType string, gitlabRepoVisibility string) ([]*Repository, error) {

	if client == nil {
		log.Fatalf("Couldn't acquire a client to talk to %s", service)
	}

	var repositories []*Repository

	if service == "github" {
		ctx := context.Background()
		options := github.RepositoryListOptions{Type: githubRepoType}
		for {
			repos, resp, err := client.(*github.Client).Repositories.List(ctx, "", &options)
			if err == nil {
				for _, repo := range repos {
					namespace := strings.Split(*repo.FullName, "/")[0]
					repositories = append(repositories, &Repository{GitURL: *repo.GitURL, Name: *repo.Name, Namespace: namespace})
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
		var visibility gitlab.VisibilityValue
		switch gitlabRepoVisibility {
		case "public":
			visibility = gitlab.PublicVisibility
		case "private":
			visibility = gitlab.PrivateVisibility
		case "internal":
			fallthrough
		case "default":
			visibility = gitlab.InternalVisibility
		}
		options := gitlab.ListProjectsOptions{Visibility: &visibility}
		for {
			repos, resp, err := client.(*gitlab.Client).Projects.ListProjects(&options)
			if err == nil {
				for _, repo := range repos {
					namespace := strings.Split(repo.PathWithNamespace, "/")[0]
					repositories = append(repositories, &Repository{GitURL: repo.SSHURLToRepo, Name: repo.Name, Namespace: namespace})
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
