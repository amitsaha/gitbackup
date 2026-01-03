package main

import (
	"context"
	"log"
	"time"

	"github.com/google/go-github/v34/github"
)

// defaultMigrationPollingInterval is the default time to wait between migration status checks
const defaultMigrationPollingInterval = 60 * time.Second

func handleGithubCreateUserMigration(client interface{}, c *appConfig) {
	repos, err := getRepositories(
		client,
		c.service,
		c.githubRepoType,
		c.githubNamespaceWhitelist,
		c.gitlabProjectVisibility,
		c.gitlabProjectMembershipType,
		c.ignoreFork,
		c.forgejoRepoType,
	)
	if err != nil {
		log.Fatalf("Error getting list of repositories: %v", err)
	}

	createUserMigration(client, c, repos)
	processOrganizationMigrations(client, c)
}

// createUserMigration creates and optionally downloads a user migration
func createUserMigration(client interface{}, c *appConfig, repos []*Repository) {
	log.Printf("Creating a user migration for %d repos", len(repos))
	m, err := createGithubUserMigration(
		context.Background(),
		client, repos,
		c.githubCreateUserMigrationRetry,
		c.githubCreateUserMigrationRetryMax,
	)
	if err != nil {
		log.Fatalf("Error creating migration: %v", err)
	}

	if c.githubWaitForMigrationComplete {
		err = downloadGithubUserMigrationData(
			context.Background(),
			client, c.backupDir,
			m.ID,
			defaultMigrationPollingInterval,
		)
		if err != nil {
			log.Fatalf("Error querying/downloading migration: %v", err)
		}
	}
}

// processOrganizationMigrations creates migrations for all user-owned organizations
func processOrganizationMigrations(client interface{}, c *appConfig) {
	orgs, err := getGithubUserOwnedOrgs(context.Background(), client)
	if err != nil {
		log.Fatal("Error getting user organizations", err)
	}
	
	for _, o := range orgs {
		createOrganizationMigration(client, c, o)
	}
}

// createOrganizationMigration creates and optionally downloads a migration for a single organization
func createOrganizationMigration(client interface{}, c *appConfig, org *github.Organization) {
	orgRepos, err := getGithubOrgRepositories(context.Background(), client, org)
	if err != nil {
		log.Fatal("Error getting org repos", err)
	}
	
	if len(orgRepos) == 0 {
		log.Printf("No repos found in %s", *org.Login)
		return
	}
	
	log.Printf("Creating a org migration (%s) for %d repos", *org.Login, len(orgRepos))
	oMigration, err := createGithubOrgMigration(context.Background(), client, *org.Login, orgRepos)
	if err != nil {
		log.Fatalf("Error creating migration: %v", err)
	}
	
	if c.githubWaitForMigrationComplete {
		downloadGithubOrgMigrationData(
			context.Background(),
			client,
			*org.Login,
			c.backupDir,
			oMigration.ID,
			defaultMigrationPollingInterval,
		)
	}
}
