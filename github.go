package main

import (
	"context"
	"log"
	"strings"

	"github.com/google/go-github/v34/github"
)

func getGithubRepositories(
	client interface{},
	service string, githubRepoType string, githubNamespaceWhitelist []string,
	gitlabProjectVisibility string, gitlabProjectMembershipType string,
	ignoreFork bool,
) ([]*Repository, error) {

	if client == nil {
		log.Fatalf("Couldn't acquire a client to talk to %s", service)
	}

	var repositories []*Repository
	var cloneURL string

	ctx := context.Background()

	if githubRepoType == "starred" {
		options := github.ActivityListStarredOptions{}
		for {
			stars, resp, err := client.(*github.Client).Activity.ListStarred(ctx, "", &options)
			if err == nil {
				for _, star := range stars {
					if *star.Repository.Fork && ignoreFork {
						continue
					}
					namespace := strings.Split(*star.Repository.FullName, "/")[0]
					if useHTTPSClone != nil && *useHTTPSClone {
						cloneURL = *star.Repository.CloneURL
					} else {
						cloneURL = *star.Repository.SSHURL
					}
					repositories = append(repositories, &Repository{CloneURL: cloneURL, Name: *star.Repository.Name, Namespace: namespace, Private: *star.Repository.Private})
				}
			} else {
				return nil, err
			}
			if resp.NextPage == 0 {
				break
			}
			options.ListOptions.Page = resp.NextPage
		}
		return repositories, nil
	}

	options := github.RepositoryListOptions{Type: githubRepoType}
	githubNamespaceWhitelistLength := len(githubNamespaceWhitelist)

	for {
		repos, resp, err := client.(*github.Client).Repositories.List(ctx, "", &options)
		if err == nil {
			for _, repo := range repos {
				if *repo.Fork && ignoreFork {
					continue
				}
				namespace := strings.Split(*repo.FullName, "/")[0]

				if githubNamespaceWhitelistLength > 0 && !contains(githubNamespaceWhitelist, namespace) {
					continue
				}

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
	return repositories, nil
}
