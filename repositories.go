package main

import (
	"net/http"
)

// Response is derived from the following sources:
// https://github.com/google/go-github/blob/27c7c32b6d369610435bd2ad7b4d8554f235eb01/github/github.go#L301
// https://github.com/xanzy/go-gitlab/blob/3acf8d75e9de17ad4b41839a7cabbf2537760ab4/gitlab.go#L286
type Response struct {
	*http.Response

	// These fields provide the page values for paginating through a set of
	// results. Any or all of these may be set to the zero value for
	// responses that are not part of a paginated set, or for which there
	// are no additional pages.

	NextPage  int
	PrevPage  int
	FirstPage int
	LastPage  int
}

// Repository represents a git repository to be backed up
type Repository struct {
	CloneURL  string
	Name      string
	Namespace string
	Private   bool
}

// getRepositories retrieves all repositories from the specified git service
// that match the given criteria (repo type, visibility, membership, etc.)
func getRepositories(
	client interface{},
	service string, githubRepoType string, githubNamespaceWhitelist []string,
	gitlabProjectVisibility string, gitlabProjectMembershipType string,
	ignoreFork bool,
) ([]*Repository, error) {
	var repositories []*Repository
	var err error

	switch service {
	case "github":
		repositories, err = getGithubRepositories(
			client,
			service,
			githubRepoType,
			githubNamespaceWhitelist,
			gitlabProjectVisibility,
			gitlabProjectMembershipType,
			ignoreFork,
		)

	case "gitlab":
		repositories, err = getGitlabRepositories(
			client,
			service,
			githubRepoType,
			githubNamespaceWhitelist,
			gitlabProjectVisibility,
			gitlabProjectMembershipType,
			ignoreFork,
		)
	case "bitbucket":
		repositories, err = getBitbucketRepositories(
			client,
			service,
			githubRepoType,
			githubNamespaceWhitelist,
			gitlabProjectVisibility,
			gitlabProjectMembershipType,
			ignoreFork,
		)
	}
	return repositories, err
}
