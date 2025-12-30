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

	var repositories []*Repository

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

			httpsURL, sshURL := extractBitbucketCloneURLs(repo.Links)
			cloneURL := getCloneURL(httpsURL, sshURL)

			repositories = append(repositories, &Repository{
				CloneURL:  cloneURL,
				Name:      repo.Slug,
				Namespace: namespace,
				Private:   repo.Is_private,
			})
		}
	}
	return repositories, nil
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
