package main

import (
	"log"
	"strings"

	gitlab "github.com/xanzy/go-gitlab"
)

func getGitlabRepositories(
	client interface{},
	service string, githubRepoType string, githubNamespaceWhitelist []string,
	gitlabProjectVisibility string, gitlabProjectMembershipType string,
	ignoreFork bool, forgejoRepoType string,
) ([]*Repository, error) {

	if client == nil {
		log.Fatalf("Couldn't acquire a client to talk to %s", service)
	}

	gitlabListOptions := buildGitlabListOptions(gitlabProjectMembershipType, gitlabProjectVisibility)

	return fetchGitlabProjects(client.(*gitlab.Client), gitlabListOptions)
}

// buildGitlabListOptions constructs the list options for GitLab project queries
func buildGitlabListOptions(membershipType, visibility string) gitlab.ListProjectsOptions {
	var boolTrue = true
	options := gitlab.ListProjectsOptions{}

	// Set membership type filters
	switch membershipType {
	case "owner":
		options.Owned = &boolTrue
	case "member":
		options.Membership = &boolTrue
	case "starred":
		options.Starred = &boolTrue
	case "all":
		options.Owned = &boolTrue
		options.Membership = &boolTrue
		options.Starred = &boolTrue
	}

	// Set visibility filter
	if visibility != "all" {
		visibilityValue := getGitlabVisibility(visibility)
		options.Visibility = &visibilityValue
	}

	return options
}

// getGitlabVisibility converts a visibility string to GitLab's VisibilityValue type
func getGitlabVisibility(visibility string) gitlab.VisibilityValue {
	switch visibility {
	case "public":
		return gitlab.PublicVisibility
	case "private":
		return gitlab.PrivateVisibility
	case "internal", "default":
		return gitlab.InternalVisibility
	default:
		return gitlab.InternalVisibility
	}
}

// fetchGitlabProjects retrieves all projects from GitLab with pagination
func fetchGitlabProjects(client *gitlab.Client, options gitlab.ListProjectsOptions) ([]*Repository, error) {
	var repositories []*Repository

	for {
		repos, resp, err := client.Projects.ListProjects(&options)
		if err != nil {
			return nil, err
		}
		for _, repo := range repos {
			repositories = append(repositories, buildGitlabRepository(repo))
		}
		if resp.NextPage == 0 {
			break
		}
		options.ListOptions.Page = resp.NextPage
	}
	return repositories, nil
}

// buildGitlabRepository converts a GitLab project to our Repository type
func buildGitlabRepository(repo *gitlab.Project) *Repository {
	namespace := strings.Split(repo.PathWithNamespace, "/")[0]
	cloneURL := getCloneURL(repo.WebURL, repo.SSHURLToRepo)
	
	return &Repository{
		CloneURL:  cloneURL,
		Name:      repo.Name,
		Namespace: namespace,
		Private:   repo.Visibility == "private",
	}
}
