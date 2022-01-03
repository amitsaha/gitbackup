package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
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

func main() {

	// Git host
	var gitHost string

	// Used for waiting for all the goroutines to finish before exiting
	var wg sync.WaitGroup
	defer wg.Wait()

	// The services we know of and their default public host names
	knownServices := map[string]string{
		"github":    "github.com",
		"gitlab":    "gitlab.com",
		"bitbucket": "bitbucker.org",
	}

	// Generic flags
	service := flag.String("service", "", "Git Hosted Service Name (github/gitlab/bitbucket)")
	githostURL := flag.String("githost.url", "", "DNS of the custom Git host")
	backupDir := flag.String("backupdir", "", "Backup directory")
	ignorePrivate = flag.Bool("ignore-private", false, "Ignore private repositories/projects")
	ignoreFork := flag.Bool("ignore-fork", false, "Ignore repositories which are forks")
	useHTTPSClone = flag.Bool("use-https-clone", false, "Use HTTPS for cloning instead of SSH")
	bare := flag.Bool("bare", false, "Clone bare repositories")

	// GitHub specific flags
	githubRepoType := flag.String("github.repoType", "all", "Repo types to backup (all, owner, member, starred)")

	githubCreateUserMigration := flag.Bool("github.createUserMigration", false, "Download user data")
	githubCreateUserMigrationRetry := flag.Bool("github.createUserMigrationRetry", true, "Retry creating the GitHub user migration if we get an error")
	githubCreateUserMigrationRetryMax := flag.Int("github.createUserMigrationRetryMax", defaultMaxUserMigrationRetry, "Number of retries to attempt for creating GitHub user migration")
	githubListUserMigrations := flag.Bool("github.listUserMigrations", false, "List available user migrations")
	githubWaitForMigrationComplete := flag.Bool("github.waitForUserMigration", true, "Wait for migration to complete")

	// Gitlab specific flags
	gitlabProjectVisibility := flag.String("gitlab.projectVisibility", "internal", "Visibility level of Projects to clone (internal, public, private)")
	gitlabProjectMembershipType := flag.String("gitlab.projectMembershipType", "all", "Project type to clone (all, owner, member, starred)")

	flag.Parse()

	if _, ok := knownServices[*service]; !ok {
		log.Fatal("Please specify the git service type: github, gitlab, bitbucket")
	}

	if !validGitlabProjectMembership(*gitlabProjectMembershipType) {
		log.Fatal("Please specify a valid gitlab project membership - all/owner/member")
	}

	if len(*githostURL) != 0 {
		u, err := url.Parse(*githostURL)
		if err != nil {
			panic(err)
		}
		gitHost = u.Host
	} else {
		gitHost = knownServices[*service]
	}

	*backupDir = setupBackupDir(*backupDir, gitHost)

	client := newClient(*service, *githostURL)

	if *githubListUserMigrations {
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

	} else if *githubCreateUserMigration {

		repos, err := getRepositories(client, *service, *githubRepoType, *gitlabProjectVisibility, *gitlabProjectMembershipType, *ignoreFork)
		if err != nil {
			log.Fatalf("Error getting list of repositories: %v", err)
		}

		log.Printf("Creating a user migration for %d repos", len(repos))
		m, err := createGithubUserMigration(context.Background(), client, repos, *githubCreateUserMigrationRetry, *githubCreateUserMigrationRetryMax)
		if err != nil {
			log.Fatalf("Error creating migration: %v", err)
		}

		if *githubWaitForMigrationComplete {
			migrationStatePollingDuration := 60 * time.Second
			err = downloadGithubUserMigrationData(context.Background(), client, *backupDir, m.ID, migrationStatePollingDuration)
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
			if *githubWaitForMigrationComplete {
				migrationStatePollingDuration := 60 * time.Second
				downloadGithubOrgMigrationData(context.Background(), client, *o.Login, *backupDir, oMigration.ID, migrationStatePollingDuration)
			}
		}

	} else {
		tokens := make(chan bool, MaxConcurrentClones)
		gitHostUsername = getUsername(client, *service)

		if len(gitHostUsername) == 0 && !*ignorePrivate && *useHTTPSClone {
			log.Fatal("Your Git host's username is needed for backing up private repositories via HTTPS")
		}
		repos, err := getRepositories(
			client, *service, *githubRepoType,
			*gitlabProjectVisibility, *gitlabProjectMembershipType, *ignoreFork,
		)
		if err != nil {
			log.Fatal(err)
		} else {
			log.Printf("Backing up %v repositories now..\n", len(repos))
			for _, repo := range repos {
				tokens <- true
				wg.Add(1)
				go func(repo *Repository) {
					stdoutStderr, err := backUp(*backupDir, repo, *bare, &wg)
					if err != nil {
						log.Printf("Error backing up %s: %s\n", repo.Name, stdoutStderr)
					}
					<-tokens
				}(repo)
			}
		}
	}
}
