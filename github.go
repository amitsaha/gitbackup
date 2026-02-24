package main

import (
	"context"
	"strings"

	"github.com/google/go-github/v34/github"
)

func getGithubRepositories(
	client *github.Client,
	githubRepoType string, githubNamespaceWhitelist []string,
	ignoreFork bool,
) ([]*Repository, error) {

	var repositories []*Repository

	ctx := context.Background()

	if githubRepoType == "starred" {
		return getGithubStarredRepositories(ctx, client, ignoreFork)
	}

	options := github.RepositoryListOptions{Type: githubRepoType}

	for {
		repos, resp, err := client.Repositories.List(ctx, "", &options)
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

			var httpsCloneURL, sshCloneURL string
			if repo.CloneURL != nil {
				httpsCloneURL = *repo.CloneURL
			}
			if repo.SSHURL != nil {
				sshCloneURL = *repo.SSHURL
			}

			cloneURL := getCloneURL(httpsCloneURL, sshCloneURL)
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

			var httpsCloneURL, sshCloneURL string
			if star.Repository.CloneURL != nil {
				httpsCloneURL = *star.Repository.CloneURL
			}
			if star.Repository.SSHURL != nil {
				sshCloneURL = *star.Repository.SSHURL
			}

			cloneURL := getCloneURL(httpsCloneURL, sshCloneURL)
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
