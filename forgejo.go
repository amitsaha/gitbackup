package main

import (
	"log"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
)

func getForgejoRepositories(
	client interface{},
	service string, githubRepoType string, githubNamespaceWhitelist []string,
	gitlabProjectVisibility string, gitlabProjectMembershipType string,
	ignoreFork bool,
) ([]*Repository, error) {
	if client == nil {
		log.Fatalf("Couldn't acquire a client to talk to %s", service)
	}

	var repositories []*Repository

	repos, _, err := client.(*forgejo.Client).ListMyRepos(forgejo.ListReposOptions{})
	if err != nil {
		log.Fatalf("Error fetching repositories from %s: %v", service, err)
	}

	for _, repo := range repos {
		repositories = append(repositories, &Repository{
			CloneURL:  repo.CloneURL,
			Name:      repo.Name,
			Namespace: repo.Owner.UserName,
			Private:   repo.Private,
		})
	}

	return repositories, nil
}
