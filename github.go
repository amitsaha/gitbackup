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

	ctx := context.Background()

	if githubRepoType == "starred" {
		return getGithubStarredRepositories(ctx, client.(*github.Client), ignoreFork)
	}

	options := github.RepositoryListOptions{Type: githubRepoType}

	for {
		repos, resp, err := client.(*github.Client).Repositories.List(ctx, "", &options)
		if err != nil {
			return nil, err
		}
		for _, repo := range repos {
			if *repo.Fork && ignoreFork {
				continue
			}
			namespace := strings.Split(*repo.FullName, "/")[0]

			if len(githubNamespaceWhitelist) > 0 && !contains(githubNamespaceWhitelist, namespace) {
				continue
			}

			cloneURL := getCloneURL(*repo.CloneURL, *repo.SSHURL)
			repositories = append(repositories, &Repository{
				CloneURL:  cloneURL,
				Name:      *repo.Name,
				Namespace: namespace,
				Private:   *repo.Private,
			})
		}
		if resp.NextPage == 0 {
			break
		}
		options.ListOptions.Page = resp.NextPage
	}
	return repositories, nil
}

func getGithubStarredRepositories(ctx context.Context, client *github.Client, ignoreFork bool) ([]*Repository, error) {
	var repositories []*Repository
	options := github.ActivityListStarredOptions{}

	for {
		stars, resp, err := client.Activity.ListStarred(ctx, "", &options)
		if err != nil {
			return nil, err
		}
		for _, star := range stars {
			if *star.Repository.Fork && ignoreFork {
				continue
			}
			namespace := strings.Split(*star.Repository.FullName, "/")[0]
			cloneURL := getCloneURL(*star.Repository.CloneURL, *star.Repository.SSHURL)
			repositories = append(repositories, &Repository{
				CloneURL:  cloneURL,
				Name:      *star.Repository.Name,
				Namespace: namespace,
				Private:   *star.Repository.Private,
			})
		}
		if resp.NextPage == 0 {
			break
		}
		options.ListOptions.Page = resp.NextPage
	}
	return repositories, nil
}
