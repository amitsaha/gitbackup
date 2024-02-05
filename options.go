package main

import (
	"errors"
	"flag"
	"strings"
)

func initConfig(args []string) (*appConfig, error) {

	var githubNamespaceWhitelistString string
	var c appConfig

	fs := flag.NewFlagSet("gitbackup", flag.ExitOnError)

	// Generic flags
	fs.StringVar(&c.service, "service", "", "Git Hosted Service Name (github/gitlab/bitbucket)")
	fs.StringVar(&c.gitHostURL, "githost.url", "", "DNS of the custom Git host")
	fs.StringVar(&c.backupDir, "backupdir", "", "Backup directory")
	fs.BoolVar(&c.ignorePrivate, "ignore-private", false, "Ignore private repositories/projects")
	fs.BoolVar(&c.ignoreFork, "ignore-fork", false, "Ignore repositories which are forks")
	fs.BoolVar(&c.useHTTPSClone, "use-https-clone", false, "Use HTTPS for cloning instead of SSH")
	fs.BoolVar(&c.bare, "bare", false, "Clone bare repositories")

	// GitHub specific flags
	fs.StringVar(&c.githubRepoType, "github.repoType", "all", "Repo types to backup (all, owner, member, starred)")
	fs.StringVar(
		&githubNamespaceWhitelistString, "github.namespaceWhitelist",
		"", "Organizations/Users from where we should clone (separate each value by a comma: 'user1,org2')",
	)
	fs.BoolVar(&c.githubCreateUserMigration, "github.createUserMigration", false, "Download user data")
	fs.BoolVar(
		&c.githubCreateUserMigrationRetry, "github.createUserMigrationRetry", true,
		"Retry creating the GitHub user migration if we get an error",
	)
	fs.IntVar(
		&c.githubCreateUserMigrationRetryMax, "github.createUserMigrationRetryMax",
		defaultMaxUserMigrationRetry,
		"Number of retries to attempt for creating GitHub user migration",
	)
	fs.BoolVar(
		&c.githubListUserMigrations,
		"github.listUserMigrations",
		false,
		"List available user migrations",
	)
	fs.BoolVar(
		&c.githubWaitForMigrationComplete,
		"github.waitForUserMigration",
		true,
		"Wait for migration to complete",
	)

	// Gitlab specific flags
	fs.StringVar(
		&c.gitlabProjectVisibility,
		"gitlab.projectVisibility",
		"internal",
		"Visibility level of Projects to clone (internal, public, private)",
	)
	fs.StringVar(
		&c.gitlabProjectMembershipType,
		"gitlab.projectMembershipType", "all",
		"Project type to clone (all, owner, member, starred)",
	)

	err := fs.Parse(args)
	if err != nil && !errors.Is(err, flag.ErrHelp) {
		return nil, err
	}

	// Split namespaces
	if len(c.githubNamespaceWhitelist) > 0 {
		c.githubNamespaceWhitelist = strings.Split(githubNamespaceWhitelistString, ",")
	}
	c.backupDir = setupBackupDir(&c.backupDir, &c.service, &c.gitHostURL)
	return &c, nil
}
