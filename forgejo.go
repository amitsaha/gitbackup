package main

import (
	"fmt"
	"log"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
)

func getForgejoRepositories(
	client *forgejo.Client,
	forgejoRepoType string,
) ([]*Repository, error) {

	switch forgejoRepoType {
	case "starred":
		user, _, err := client.GetMyUserInfo()
		if err != nil {
			log.Fatalf("Error fetching user info from forgejo: %v", err)
		}

		log.Printf("Found user %s with ID %d", user.UserName, user.ID)

		repos, err := paginateForgejoRepositories(func(page int) ([]*forgejo.Repository, *forgejo.Response, error) {
			return client.SearchRepos(forgejo.SearchRepoOptions{
				ListOptions:     forgejo.ListOptions{Page: page},
				StarredByUserID: user.ID,
			})
		})
		if err != nil {
			return nil, fmt.Errorf("fetching starred repositories from forgejo: %v", err)
		}

		return repos, nil
	case "user", "":
		repos, err := paginateForgejoRepositories(func(page int) ([]*forgejo.Repository, *forgejo.Response, error) {
			return client.ListMyRepos(forgejo.ListReposOptions{
				ListOptions: forgejo.ListOptions{Page: page},
			})
		})
		if err != nil {
			return nil, fmt.Errorf("fetching user repositories from forgejo: %v", err)
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
