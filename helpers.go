package main

import (
	"context"
	"log"

	"github.com/google/go-github/v34/github"
	gitlab "github.com/xanzy/go-gitlab"
)

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

	return ""
}

func validGitlabProjectMembership(membership string) bool {
	validMemberships := []string{"all", "owner", "member"}
	for _, m := range validMemberships {
		if membership == m {
			return true
		}
	}
	return false
}
