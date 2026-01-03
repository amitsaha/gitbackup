package main

import (
	"context"
	"log"

	"github.com/google/go-github/v34/github"
	"github.com/ktrysmt/go-bitbucket"
	gitlab "github.com/xanzy/go-gitlab"
)

// getUsername retrieves the username for the authenticated user from the git service
func getUsername(client interface{}, service string) string {
	if client == nil {
		log.Fatalf("Couldn't acquire a client to talk to %s", service)
	}

	switch service {
	case "github":
		return getGithubUsername(client.(*github.Client))
	case "gitlab":
		return getGitlabUsername(client.(*gitlab.Client))
	case "bitbucket":
		return getBitbucketUsername(client.(*bitbucket.Client))
	default:
		return ""
	}
}

// getGithubUsername retrieves the GitHub username
func getGithubUsername(client *github.Client) string {
	ctx := context.Background()
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		log.Fatal("Error retrieving username", err.Error())
	}
	return *user.Login
}

// getGitlabUsername retrieves the GitLab username
func getGitlabUsername(client *gitlab.Client) string {
	user, _, err := client.Users.CurrentUser()
	if err != nil {
		log.Fatal("Error retrieving username", err.Error())
	}
	return user.Username
}

// getBitbucketUsername retrieves the Bitbucket username
func getBitbucketUsername(client *bitbucket.Client) string {
	user, err := client.User.Profile()
	if err != nil {
		log.Fatal("Error retrieving username", err.Error())
	}
	return user.Username
}

// validGitlabProjectMembership checks if the given membership type is valid
func validGitlabProjectMembership(membership string) bool {
	validMemberships := []string{"all", "owner", "member", "starred"}
	for _, m := range validMemberships {
		if membership == m {
			return true
		}
	}
	return false
}

// contains checks if a string exists in a slice of strings
func contains(list []string, x string) bool {
	for _, item := range list {
		if item == x {
			return true
		}
	}
	return false
}

// getCloneURL returns the appropriate clone URL based on the useHTTPSClone setting
func getCloneURL(httpsURL, sshURL string) string {
	if useHTTPSClone != nil && *useHTTPSClone {
		return httpsURL
	}
	return sshURL
}