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
	ignoreFork bool, forgejoRepoType string,
) ([]*Repository, error) {

	if client == nil {
		log.Fatalf("Couldn't acquire a client to talk to %s", service)
	}

	ctx := context.Background()

	if githubRepoType == "starred" {
		return getGithubStarredRepositories(ctx, client.(*github.Client), ignoreFork)
	}

	return getGithubUserRepositories(ctx, client.(*github.Client), githubRepoType, githubNamespaceWhitelist, ignoreFork)
}

// getGithubUserRepositories retrieves user repositories (not starred) from GitHub
func getGithubUserRepositories(ctx context.Context, client *github.Client, repoType string, namespaceWhitelist []string, ignoreFork bool) ([]*Repository, error) {
	var repositories []*Repository
	options := github.RepositoryListOptions{Type: repoType}

	for {
		repos, resp, err := client.Repositories.List(ctx, "", &options)
		if err != nil {
			return nil, err
		}
		for _, repo := range repos {
			if shouldSkipGithubRepo(repo, namespaceWhitelist, ignoreFork) {
				continue
			}
			repositories = append(repositories, buildGithubRepository(repo))
		}
		if resp.NextPage == 0 {
			break
		}
		options.ListOptions.Page = resp.NextPage
	}
	return repositories, nil
}

// shouldSkipGithubRepo determines if a repository should be skipped based on filters
func shouldSkipGithubRepo(repo *github.Repository, namespaceWhitelist []string, ignoreFork bool) bool {
	if *repo.Fork && ignoreFork {
		return true
	}
	namespace := strings.Split(*repo.FullName, "/")[0]
	return len(namespaceWhitelist) > 0 && !contains(namespaceWhitelist, namespace)
}

// buildGithubRepository converts a GitHub repository to our Repository type
func buildGithubRepository(repo *github.Repository) *Repository {
	namespace := strings.Split(*repo.FullName, "/")[0]
	
	var httpsCloneURL, sshCloneURL string
	if repo.CloneURL != nil {
		httpsCloneURL = *repo.CloneURL
	}
	if repo.SSHURL != nil {
		sshCloneURL = *repo.SSHURL
	}

	return &Repository{
		CloneURL:  getCloneURL(httpsCloneURL, sshCloneURL),
		Name:      *repo.Name,
		Namespace: namespace,
		Private:   *repo.Private,
	}
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
			repositories = append(repositories, buildGithubRepository(star.Repository))
		}
		if resp.NextPage == 0 {
			break
		}
		options.ListOptions.Page = resp.NextPage
	}
	return repositories, nil
}
