package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/google/go-github/v34/github"
	bitbucket "github.com/ktrysmt/go-bitbucket"
	gitlab "github.com/xanzy/go-gitlab"
)

// Response is derived from the following sources:
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

// Repository is a container for the details for a repository
// we will backup
type Repository struct {
	CloneURL  string
	Name      string
	Namespace string
	Private   bool
}

func getRepositories(client interface{}, service string, githubRepoType string, gitlabRepoVisibility string, gitlabProjectType string, ignoreFork bool) ([]*Repository, error) {

	if client == nil {
		log.Fatalf("Couldn't acquire a client to talk to %s", service)
	}

	var repositories []*Repository
	var cloneURL string

	if service == "github" {
		ctx := context.Background()
		options := github.RepositoryListOptions{Type: githubRepoType}
		for {
			repos, resp, err := client.(*github.Client).Repositories.List(ctx, "", &options)
			if err == nil {
				for _, repo := range repos {
					if *repo.Fork && ignoreFork {
						continue
					}
					namespace := strings.Split(*repo.FullName, "/")[0]
					if useHTTPSClone != nil && *useHTTPSClone {
						cloneURL = *repo.CloneURL
					} else {
						cloneURL = *repo.SSHURL
					}
					repositories = append(repositories, &Repository{CloneURL: cloneURL, Name: *repo.Name, Namespace: namespace, Private: *repo.Private})
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
		var options gitlab.ListProjectsOptions

		var owned bool
		var memberOf bool

		if gitlabProjectType == "owner" {
			owned = true
		}
		if gitlabProjectType == "member" {
			memberOf = true
		}

		if gitlabProjectType == "all" {
			owned = true
			memberOf = true
		}

		if gitlabRepoVisibility != "all" {
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
			options = gitlab.ListProjectsOptions{Visibility: &visibility, Membership: &memberOf, Owned: &owned}
		} else {
			options = gitlab.ListProjectsOptions{Membership: &memberOf, Owned: &owned}
		}

		for {
			repos, resp, err := client.(*gitlab.Client).Projects.ListProjects(&options)
			if err == nil {
				for _, repo := range repos {
					namespace := strings.Split(repo.PathWithNamespace, "/")[0]
					if useHTTPSClone != nil && *useHTTPSClone {
						cloneURL = repo.WebURL
					} else {
						cloneURL = repo.SSHURLToRepo
					}
					repositories = append(repositories, &Repository{CloneURL: cloneURL, Name: repo.Name, Namespace: namespace, Private: repo.Public})
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

	if service == "bitbucket" {
		resp, err := client.(*bitbucket.Client).Workspaces.List()
		if err != nil {
			return nil, err
		}

		for _, workspace := range resp.Workspaces {
			options := &bitbucket.RepositoriesOptions{Owner: workspace.Slug}

			resp, err := client.(*bitbucket.Client).Repositories.ListForAccount(options)
			if err != nil {
				return nil, err
			}

			for _, repo := range resp.Items {
				namespace := strings.Split(repo.Full_name, "/")[0]

				linkmaps, ok := repo.Links["clone"].([]interface{})

				var httpsURL string
				var sshURL string

				if ok {
					for _, linkmaps := range linkmaps {
						linkmap, ok := linkmaps.(map[string]interface{})

						if ok {
							if linkmap["name"] == "https" {
								httpsURL = linkmap["href"].(string)
							}

							if linkmap["name"] == "ssh" {
								sshURL = linkmap["href"].(string)
							}
						}
					}
				}

				if useHTTPSClone != nil && *useHTTPSClone {
					cloneURL = httpsURL
				} else {
					cloneURL = sshURL
				}

				repositories = append(repositories, &Repository{CloneURL: cloneURL, Name: repo.Slug, Namespace: namespace, Private: repo.Is_private})
			}
		}
	}

	return repositories, nil
}
