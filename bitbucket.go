package main

import (
	"log"
	"strings"

	bitbucket "github.com/ktrysmt/go-bitbucket"
)

func getBitbucketRepositories(
	client interface{},
	service string, githubRepoType string, githubNamespaceWhitelist []string,
	gitlabProjectVisibility string, gitlabProjectMembershipType string,
	ignoreFork bool, forgejoRepoType string,
) ([]*Repository, error) {

	if client == nil {
		log.Fatalf("Couldn't acquire a client to talk to %s", service)
	}

	workspaces, err := client.(*bitbucket.Client).Workspaces.List()
	if err != nil {
		return nil, err
	}

	return fetchBitbucketRepositoriesFromWorkspaces(client.(*bitbucket.Client), workspaces.Workspaces)
}

// fetchBitbucketRepositoriesFromWorkspaces retrieves repositories from all workspaces
func fetchBitbucketRepositoriesFromWorkspaces(client *bitbucket.Client, workspaces []bitbucket.Workspace) ([]*Repository, error) {
	var repositories []*Repository

	for _, workspace := range workspaces {
		workspaceRepos, err := fetchBitbucketWorkspaceRepositories(client, workspace.Slug)
		if err != nil {
			return nil, err
		}
		repositories = append(repositories, workspaceRepos...)
	}

	return repositories, nil
}

// fetchBitbucketWorkspaceRepositories retrieves all repositories from a single workspace
func fetchBitbucketWorkspaceRepositories(client *bitbucket.Client, workspaceSlug string) ([]*Repository, error) {
	options := &bitbucket.RepositoriesOptions{Owner: workspaceSlug}
	
	resp, err := client.Repositories.ListForAccount(options)
	if err != nil {
		return nil, err
	}

	var repositories []*Repository
	for _, repo := range resp.Items {
		repositories = append(repositories, buildBitbucketRepository(repo))
	}

	return repositories, nil
}

// buildBitbucketRepository converts a Bitbucket repository to our Repository type
func buildBitbucketRepository(repo bitbucket.Repository) *Repository {
	namespace := strings.Split(repo.Full_name, "/")[0]
	httpsURL, sshURL := extractBitbucketCloneURLs(repo.Links)
	cloneURL := getCloneURL(httpsURL, sshURL)

	return &Repository{
		CloneURL:  cloneURL,
		Name:      repo.Slug,
		Namespace: namespace,
		Private:   repo.Is_private,
	}
}

func extractBitbucketCloneURLs(links map[string]interface{}) (httpsURL, sshURL string) {
	linkmaps, ok := links["clone"].([]interface{})
	if !ok {
		return "", ""
	}

	for _, linkmaps := range linkmaps {
		linkmap, ok := linkmaps.(map[string]interface{})
		if !ok {
			continue
		}

		if linkmap["name"] == "https" {
			httpsURL = linkmap["href"].(string)
		}

		if linkmap["name"] == "ssh" {
			sshURL = linkmap["href"].(string)
		}
	}
	return httpsURL, sshURL
}
