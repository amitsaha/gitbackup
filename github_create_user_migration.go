package main

import (
	"context"
	"log"
	"time"
)

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
		migrationStatePollingDuration := 60 * time.Second
		err = downloadGithubUserMigrationData(
			context.Background(),
			client, c.backupDir,
			m.ID,
			migrationStatePollingDuration,
		)
		if err != nil {
			log.Fatalf("Error querying/downloading migration: %v", err)
		}
	}

	orgs, err := getGithubUserOwnedOrgs(context.Background(), client)
	if err != nil {
		log.Fatal("Error getting user organizations", err)
	}
	for _, o := range orgs {
		orgRepos, err := getGithubOrgRepositories(context.Background(), client, o)
		if err != nil {
			log.Fatal("Error getting org repos", err)
		}
		if len(orgRepos) == 0 {
			log.Printf("No repos found in %s", *o.Login)
			continue
		}
		log.Printf("Creating a org migration (%s) for %d repos", *o.Login, len(orgRepos))
		oMigration, err := createGithubOrgMigration(context.Background(), client, *o.Login, orgRepos)
		if err != nil {
			log.Fatalf("Error creating migration: %v", err)
		}
		if c.githubWaitForMigrationComplete {
			migrationStatePollingDuration := 60 * time.Second
			downloadGithubOrgMigrationData(
				context.Background(),
				client,
				*o.Login,
				c.backupDir,
				oMigration.ID,
				migrationStatePollingDuration,
			)
		}
	}

}
