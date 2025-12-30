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

		repos, err := paginateForgejoRepositories(func(page int) ([]*forgejo.Repository, *forgejo.Response, error) {
			return forgejoClient.SearchRepos(forgejo.SearchRepoOptions{
				ListOptions:     forgejo.ListOptions{Page: page},
				StarredByUserID: user.ID,
			})
		})
		if err != nil {
			return nil, fmt.Errorf("fetching starred repositories from %s: %v", service, err)
		}

		return repos, nil
	case "user", "":
		repos, err := paginateForgejoRepositories(func(page int) ([]*forgejo.Repository, *forgejo.Response, error) {
			return forgejoClient.ListMyRepos(forgejo.ListReposOptions{
				ListOptions: forgejo.ListOptions{Page: page},
			})
		})
		if err != nil {
			return nil, fmt.Errorf("fetching user repositories from %s: %v", service, err)
		}

		return repos, nil
	default:
		return nil, fmt.Errorf("unknown repo type: %s", forgejoRepoType)
	}
}

func paginateForgejoRepositories(fetch func(page int) ([]*forgejo.Repository, *forgejo.Response, error)) ([]*Repository, error) {
	var repositories []*Repository
	page := 1

	for {
		results, resp, err := fetch(page)
		if err != nil {
			return nil, err
		}

		for _, repo := range results {
			repositories = append(repositories, &Repository{
				CloneURL:  repo.CloneURL,
				Name:      repo.Name,
				Namespace: repo.Owner.UserName,
				Private:   repo.Private,
			})
		}

		if resp == nil || resp.NextPage == 0 {
			break
		}

		page = resp.NextPage
	}

	return repositories, nil
}
