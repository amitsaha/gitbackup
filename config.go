package main

import (
	"errors"
)

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

func validateConfig(c *appConfig) error {
	if _, ok := knownServices[c.service]; !ok {
		return errors.New("Please specify the git service type: github, gitlab, bitbucket")
	}

	if !validGitlabProjectMembership(c.gitlabProjectMembershipType) {
		return errors.New("Please specify a valid gitlab project membership - all/owner/member")
	}

	return nil
}
