package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/go-github/v34/github"
)

// MaxConcurrentClones is the upper limit of the maximum number of
// concurrent git clones
const MaxConcurrentClones = 20

const defaultMaxUserMigrationRetry = 5

var gitHostToken string
var useHTTPSClone *bool
var ignorePrivate *bool
var gitHostUsername string

// The services we know of and their default public host names
var knownServices = map[string]string{
	"github":    "github.com",
	"gitlab":    "gitlab.com",
	"bitbucket": "bitbucket.org",
}

type appConfig struct {
	service       string
	gitHostURL    string
	backupDir     string
	ignorePrivate bool
	ignoreFork    bool
	useHTTPSClone bool
	bare          bool

	githubRepoType                    string
	githubNamespaceWhitelist          []string
	githubCreateUserMigration         bool
	githubCreateUserMigrationRetry    bool
	githubCreateUserMigrationRetryMax int
	githubListUserMigrations          bool
	githubWaitForMigrationComplete    bool

	gitlabProjectVisibility     string
	gitlabProjectMembershipType string
}

func main() {

	// Used for waiting for all the goroutines to finish before exiting
	var wg sync.WaitGroup
	defer wg.Wait()

	c, err := initConfig(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	err = validateConfig(c)
	if err != nil {
		log.Fatal(err)
	}

	client := newClient(c.service, c.gitHostURL)

	if c.githubListUserMigrations {
		mList, err := getGithubUserMigrations(client)
		if err != nil {
			log.Fatal(err)
		}

		for _, m := range mList {
			mData, err := GetGithubUserMigration(client, m.ID)
			if err != nil {
				fmt.Printf("Error getting migration data: %v", *m.ID)
				// FIXME
				continue
			}

			var archiveURL string
			_, err = client.(*github.Client).Migrations.UserMigrationArchiveURL(context.Background(), *m.ID)
			if err != nil {
				archiveURL = "No Longer Available"
			} else {
				archiveURL = "Available for Download"
			}
			fmt.Printf("%v - %v - %v - %v\n", *mData.ID, *mData.CreatedAt, *mData.State, archiveURL)
		}

	} else if c.githubCreateUserMigration {

		repos, err := getRepositories(
			client,
			c.service,
			c.githubRepoType,
			c.githubNamespaceWhitelist,
			c.gitlabProjectVisibility,
			c.gitlabProjectMembershipType,
			c.ignoreFork,
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

	} else {
		tokens := make(chan bool, MaxConcurrentClones)
		gitHostUsername = getUsername(client, c.service)

		if len(gitHostUsername) == 0 && !*ignorePrivate && *useHTTPSClone {
			log.Fatal("Your Git host's username is needed for backing up private repositories via HTTPS")
		}
		repos, err := getRepositories(
			client,
			c.service,
			c.githubRepoType,
			c.githubNamespaceWhitelist,
			c.gitlabProjectVisibility,
			c.gitlabProjectMembershipType,
			c.ignoreFork,
		)
		if err != nil {
			log.Fatal(err)
		} else {
			log.Printf("Backing up %v repositories now..\n", len(repos))
			for _, repo := range repos {
				tokens <- true
				wg.Add(1)
				go func(repo *Repository) {
					stdoutStderr, err := backUp(c.backupDir, repo, c.bare, &wg)
					if err != nil {
						log.Printf("Error backing up %s: %s\n", repo.Name, stdoutStderr)
					}
					<-tokens
				}(repo)
			}
		}
	}
}
