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
	ignoreFork bool,
) ([]*Repository, error) {

	if client == nil {
		log.Fatalf("Couldn't acquire a client to talk to %s", service)
	}

	var repositories []*Repository
	var cloneURL string

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
	return repositories, nil
}
