package main

import (
	"strings"

	gitlab "github.com/xanzy/go-gitlab"
)

func getGitlabRepositories(
	client *gitlab.Client,
	gitlabProjectVisibility string, gitlabProjectMembershipType string,
	ignoreFork bool,
) ([]*Repository, error) {

	var repositories []*Repository

	var visibility gitlab.VisibilityValue
	var boolTrue bool = true

	gitlabListOptions := gitlab.ListProjectsOptions{}

	switch gitlabProjectMembershipType {
	case "owner":
		gitlabListOptions.Owned = &boolTrue
	case "member":
		gitlabListOptions.Membership = &boolTrue
	case "starred":
		gitlabListOptions.Starred = &boolTrue
	case "all":
		gitlabListOptions.Owned = &boolTrue
		gitlabListOptions.Membership = &boolTrue
		gitlabListOptions.Starred = &boolTrue
	}

	if gitlabProjectVisibility != "all" {
		switch gitlabProjectVisibility {
		case "public":
			visibility = gitlab.PublicVisibility
		case "private":
			visibility = gitlab.PrivateVisibility
		case "internal":
			fallthrough
		case "default":
			visibility = gitlab.InternalVisibility
		}
		gitlabListOptions.Visibility = &visibility
	}

	for {
		repos, resp, err := client.Projects.ListProjects(&gitlabListOptions)
		if err != nil {
			return nil, err
		}
		for _, repo := range repos {
			if repo.ForkedFromProject != nil && ignoreFork {
				continue
			}
			namespace := strings.Split(repo.PathWithNamespace, "/")[0]
			cloneURL := getCloneURL(repo.WebURL, repo.SSHURLToRepo)
			repositories = append(repositories, &Repository{
				CloneURL:  cloneURL,
				Name:      repo.Name,
				Namespace: namespace,
				Private:   repo.Visibility == "private",
			})
		}
		if resp.NextPage == 0 {
			break
		}
		gitlabListOptions.ListOptions.Page = resp.NextPage
	}
	return repositories, nil
}
