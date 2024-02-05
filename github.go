package main

import (
	"context"
	"strings"

	"github.com/google/go-github/v34/github"
)

func getGithubRepositories(
	client interface{},
	service string, githubRepoType string, githubNamespaceWhitelist []string,
	gitlabProjectVisibility string, gitlabProjectMembershipType string,
	ignoreFork bool,
) ([]*Repository, error) {

	var repositories []*Repository
	var cloneURL string

	var ghRepository []*github.Repository

	ctx := context.TODO()

	switch githubRepoType {
	case "starred":
		options := github.ActivityListStarredOptions{}
		for {
			stars, resp, err := client.(*github.Client).Activity.ListStarred(ctx, "", &options)
			if err != nil {
				return nil, err
			}
			for _, star := range stars {
				ghRepository = append(ghRepository, star.Repository)
			}
			if resp.NextPage == 0 {
				break
			}
			options.ListOptions.Page = resp.NextPage
		}
	default:
		options := github.RepositoryListOptions{Type: githubRepoType}

		for {
			repos, resp, err := client.(*github.Client).Repositories.List(ctx, "", &options)
			if err != nil {
				return nil, err
			}
			for _, repo := range repos {
				ghRepository = append(ghRepository, repo)
			}
			if resp.NextPage == 0 {
				break
			}
			options.ListOptions.Page = resp.NextPage
		}
	}

	githubNamespaceWhitelistLength := len(githubNamespaceWhitelist)
	for _, repo := range ghRepository {
		if *repo.Fork && ignoreFork {
			continue
		}

		namespace := strings.Split(*repo.FullName, "/")[0]
		if githubNamespaceWhitelistLength > 0 && !contains(githubNamespaceWhitelist, namespace) {
			continue
		}
		ghRepository = append(ghRepository, repo)
		cloneURL = *repo.SSHURL
		if useHTTPSClone != nil && *useHTTPSClone {
			cloneURL = *repo.CloneURL
		}
		repositories = append(
			repositories,
			&Repository{
				CloneURL:  cloneURL,
				Name:      *repo.Name,
				Namespace: namespace,
				Private:   *repo.Private,
			},
		)
	}
	return repositories, nil
}
