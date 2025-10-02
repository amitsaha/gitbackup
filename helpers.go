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

	if service == "github" {
		ctx := context.Background()
		user, _, err := client.(*github.Client).Users.Get(ctx, "")
		if err != nil {
			log.Fatal("Error retrieving username", err.Error())
		}
		return *user.Login
	}

	if service == "gitlab" {
		user, _, err := client.(*gitlab.Client).Users.CurrentUser()
		if err != nil {
			log.Fatal("Error retrieving username", err.Error())
		}
		return user.Username
	}

	if service == "bitbucket" {
		user, err := client.(*bitbucket.Client).User.Profile()
		if err != nil {
			log.Fatal("Error retrieving username", err.Error())
		}
		return user.Username
	}

	return ""
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