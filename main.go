package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
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
	githostURL    string
	backupDir     string
	ignorePrivate bool
	ignoreFork    bool
	useHTTPSClone bool
	bareClone     bool

	githubRepoType                    string
	githubCreateUserMigration         bool
	githubCreateUserMigrationRetry    bool
	githubCreateUserMigrationRetryMax int
	githubListUserMigrations          bool
	githubWaitForMigrationComplete    bool

	gitlabProjectVisibility     string
	gitlabProjectMembershipType string
}

func parseArgs(w io.Writer, args []string) (*appConfig, error) {
	c := appConfig{}

	fs := flag.NewFlagSet("gitbackup", flag.ContinueOnError)
	fs.SetOutput(w)

	// Generic flags
	fs.StringVar(&c.service, "service", "", "Git Hosted Service Name (github/gitlab/bitbucket)")
	fs.StringVar(&c.githostURL, "githost.url", "", "DNS of the custom Git host")
	fs.StringVar(&c.backupDir, "backupdir", "", "Backup directory")
	fs.BoolVar(&c.ignorePrivate, "ignore-private", false, "Ignore private repositories/projects")
	fs.BoolVar(&c.ignoreFork, "ignore-fork", false, "Ignore repositories which are forks")
	fs.BoolVar(&c.useHTTPSClone, "use-https-clone", false, "Use HTTPS for cloning instead of SSH")
	fs.BoolVar(&c.bareClone, "bare", false, "Clone bare repositories")

	// GitHub specific flags
	fs.StringVar(&c.githubRepoType, "github.repoType", "all", "Repo types to backup (all, owner, member, starred)")
	fs.BoolVar(&c.githubCreateUserMigration, "github.createUserMigration", false, "Download user data")
	fs.BoolVar(
		&c.githubCreateUserMigrationRetry, "github.createUserMigrationRetry",
		true, "Retry creating the GitHub user migration if we get an error",
	)
	fs.IntVar(
		&c.githubCreateUserMigrationRetryMax, "github.createUserMigrationRetryMax",
		defaultMaxUserMigrationRetry, "Number of retries to attempt for creating GitHub user migration",
	)
	fs.BoolVar(
		&c.githubWaitForMigrationComplete, "github.waitForUserMigration",
		false, "Wait for migration to complete",
	)

	fs.BoolVar(
		&c.githubListUserMigrations, "github.listUserMigrations",
		false, "List available user migrations",
	)

	// Gitlab specific flags
	fs.StringVar(
		&c.gitlabProjectVisibility, "gitlab.projectVisibility",
		"internal", "Visibility level of Projects to clone (internal, public, private)",
	)
	fs.StringVar(
		&c.gitlabProjectMembershipType,
		"gitlab.projectMembershipType", "all",
		"Project type to clone (all, owner, member, starred)",
	)

	err := fs.Parse(args)
	if err != nil {
		return &c, err
	}
	if fs.NArg() != 0 {
		return &c, errors.New("Positional arguments specified")
	}
	return &c, nil
}

func validateArgs(conf *appConfig) error {

	if _, ok := knownServices[conf.service]; !ok {
		return errors.New("Please specify the git service type: github, gitlab, bitbucket")
	}

	if !validGitlabProjectMembership(conf.gitlabProjectMembershipType) {
		return errors.New("Please specify a valid gitlab project membership - all/owner/member")
	}

}

func main() {
	var backupDir string

	// Used for waiting for all the goroutines to finish before exiting
	var wg sync.WaitGroup
	defer wg.Wait()

	conf, err := parseArgs(os.Stderr, os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}

	err := validateArgs(&conf)
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}

	backupDir = setupBackupDir(&conf.backupDir, &conf.service, &conf.githostURL)
	client := newClient(conf.service, conf.githostURL)

	if conf.githubListUserMigrations {
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

	} else if conf.githubCreateUserMigration {

		repos, err := getRepositories(
			client, conf.service, conf.githubRepoType,
			conf.gitlabProjectVisibility, conf.gitlabProjectMembershipType,
			conf.ignoreFork,
		)
		if err != nil {
			log.Fatalf("Error getting list of repositories: %v", err)
		}

		log.Printf("Creating a user migration for %d repos", len(repos))
		m, err := createGithubUserMigration(
			context.Background(),
			client, repos, conf.githubCreateUserMigrationRetry,
			conf.githubCreateUserMigrationRetryMax,
		)
		if err != nil {
			log.Fatalf("Error creating migration: %v", err)
		}

		if conf.githubWaitForMigrationComplete {
			migrationStatePollingDuration := 60 * time.Second
			err = downloadGithubUserMigrationData(
				context.Background(), client,
				backupDir, m.ID, migrationStatePollingDuration,
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
			if conf.githubWaitForMigrationComplete {
				migrationStatePollingDuration := 60 * time.Second
				downloadGithubOrgMigrationData(
					context.Background(), client,
					*o.Login, backupDir,
					oMigration.ID,
					migrationStatePollingDuration,
				)
			}
		}

	} else {
		tokens := make(chan bool, MaxConcurrentClones)
		gitHostUsername = getUsername(client, conf.service)

		if len(gitHostUsername) == 0 && !conf.ignorePrivate && conf.useHTTPSClone {
			log.Fatal("Your Git host's username is needed for backing up private repositories via HTTPS")
		}
		repos, err := getRepositories(
			client, conf.service, conf.githubRepoType,
			conf.gitlabProjectVisibility, conf.gitlabProjectMembershipType, conf.ignoreFork,
		)
		if err != nil {
			log.Fatal(err)
		} else {
			log.Printf("Backing up %v repositories now..\n", len(repos))
			for _, repo := range repos {
				tokens <- true
				wg.Add(1)
				go func(repo *Repository) {
					stdoutStderr, err := backUp(backupDir, repo, conf.bareClone, &wg)
					if err != nil {
						log.Printf("Error backing up %s: %s\n", repo.Name, stdoutStderr)
					}
					<-tokens
				}(repo)
			}
		}
	}
}
