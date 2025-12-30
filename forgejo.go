package main

import (
	"fmt"
	"log"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
)

func getForgejoRepositories(
	client interface{},
	service string, githubRepoType string, githubNamespaceWhitelist []string,
	gitlabProjectVisibility string, gitlabProjectMembershipType string,
	ignoreFork bool, forgejoRepoType string,
) ([]*Repository, error) {
	if client == nil {
		log.Fatalf("Couldn't acquire a client to talk to %s", service)
	}

	forgejoClient := client.(*forgejo.Client)

	switch forgejoRepoType {
	case "starred":
		user, _, err := forgejoClient.GetMyUserInfo()
		if err != nil {
			log.Fatalf("Error fetching user info from %s: %v", service, err)
		}

		log.Printf("Found user %s with ID %d", user.UserName, user.ID)

		repos, _, err := forgejoClient.SearchRepos(forgejo.SearchRepoOptions{
			StarredByUserID: user.ID,
		})
		if err != nil {
			log.Fatalf("Error fetching starred repositories from %s: %v", service, err)
		}

		return hydrateForgejoRepositories(repos), nil
	case "user", "":
		repos, _, err := forgejoClient.ListMyRepos(forgejo.ListReposOptions{})
		if err != nil {
			log.Fatalf("Error fetching user repositories from %s: %v", service, err)
		}

		return hydrateForgejoRepositories(repos), nil
	default:
		return nil, fmt.Errorf("unknown repo type: %s", forgejoRepoType)
	}
}

func hydrateForgejoRepositories(repos []*forgejo.Repository) []*Repository {
	var repositories []*Repository

	for _, repo := range repos {
		repositories = append(repositories, &Repository{
			CloneURL:  repo.CloneURL,
			Name:      repo.Name,
			Namespace: repo.Owner.UserName,
			Private:   repo.Private,
		})
	}

	return repositories
}
