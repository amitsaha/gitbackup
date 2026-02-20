package main

import (
	"errors"
	"flag"
	"os"
	"strings"
)

// initConfig initializes and parses command-line flags into an appConfig struct.
// If a gitbackup.yml exists in the current directory, it is loaded first and
// CLI flags override any values from the config file.
func initConfig(args []string) (*appConfig, error) {

	// Try to load config file as the base configuration
	var c appConfig
	configFileLoaded := false
	if _, err := os.Stat(defaultConfigFile); err == nil {
		fc, err := loadConfigFile()
		if err != nil {
			return nil, err
		}
		c = *fileConfigToAppConfig(fc)
		configFileLoaded = true
	}

	var githubNamespaceWhitelistString string
	var flagConfig appConfig

	fs := flag.NewFlagSet("gitbackup", flag.ExitOnError)

	// Generic flags
	fs.StringVar(&flagConfig.service, "service", "", "Git Hosted Service Name (github/gitlab/bitbucket/forgejo)")
	fs.StringVar(&flagConfig.gitHostURL, "githost.url", "", "DNS of the custom Git host")
	fs.StringVar(&flagConfig.backupDir, "backupdir", "", "Backup directory")
	fs.BoolVar(&flagConfig.ignorePrivate, "ignore-private", false, "Ignore private repositories/projects")
	fs.BoolVar(&flagConfig.ignoreFork, "ignore-fork", false, "Ignore repositories which are forks")
	fs.BoolVar(&flagConfig.useHTTPSClone, "use-https-clone", false, "Use HTTPS for cloning instead of SSH")
	fs.BoolVar(&flagConfig.bare, "bare", false, "Clone bare repositories")

	// GitHub specific flags
	fs.StringVar(&flagConfig.githubRepoType, "github.repoType", "all", "Repo types to backup (all, owner, member, starred)")
	fs.StringVar(
		&githubNamespaceWhitelistString, "github.namespaceWhitelist",
		"", "Organizations/Users from where we should clone (separate each value by a comma: 'user1,org2')",
	)
	fs.BoolVar(&flagConfig.githubCreateUserMigration, "github.createUserMigration", false, "Download user data")
	fs.BoolVar(
		&flagConfig.githubCreateUserMigrationRetry, "github.createUserMigrationRetry", true,
		"Retry creating the GitHub user migration if we get an error",
	)
	fs.IntVar(
		&flagConfig.githubCreateUserMigrationRetryMax, "github.createUserMigrationRetryMax",
		defaultMaxUserMigrationRetry,
		"Number of retries to attempt for creating GitHub user migration",
	)
	fs.BoolVar(
		&flagConfig.githubListUserMigrations,
		"github.listUserMigrations",
		false,
		"List available user migrations",
	)
	fs.BoolVar(
		&flagConfig.githubWaitForMigrationComplete,
		"github.waitForUserMigration",
		true,
		"Wait for migration to complete",
	)

	// Gitlab specific flags
	fs.StringVar(
		&flagConfig.gitlabProjectVisibility,
		"gitlab.projectVisibility",
		"internal",
		"Visibility level of Projects to clone (internal, public, private)",
	)
	fs.StringVar(
		&flagConfig.gitlabProjectMembershipType,
		"gitlab.projectMembershipType", "all",
		"Project type to clone (all, owner, member, starred)",
	)

	// Forgejo specific flags
	fs.StringVar(&flagConfig.forgejoRepoType, "forgejo.repoType", "user", "Repo types to backup (user, starred)")

	err := fs.Parse(args)
	if err != nil && !errors.Is(err, flag.ErrHelp) {
		return nil, err
	}

	if configFileLoaded {
		// Only override config file values with flags that were explicitly set
		fs.Visit(func(f *flag.Flag) {
			switch f.Name {
			case "service":
				c.service = flagConfig.service
			case "githost.url":
				c.gitHostURL = flagConfig.gitHostURL
			case "backupdir":
				c.backupDir = flagConfig.backupDir
			case "ignore-private":
				c.ignorePrivate = flagConfig.ignorePrivate
			case "ignore-fork":
				c.ignoreFork = flagConfig.ignoreFork
			case "use-https-clone":
				c.useHTTPSClone = flagConfig.useHTTPSClone
			case "bare":
				c.bare = flagConfig.bare
			case "github.repoType":
				c.githubRepoType = flagConfig.githubRepoType
			case "github.namespaceWhitelist":
				// handled below
			case "gitlab.projectVisibility":
				c.gitlabProjectVisibility = flagConfig.gitlabProjectVisibility
			case "gitlab.projectMembershipType":
				c.gitlabProjectMembershipType = flagConfig.gitlabProjectMembershipType
			case "forgejo.repoType":
				c.forgejoRepoType = flagConfig.forgejoRepoType
			}
		})

		// Migration flags are always from CLI (not in config file)
		c.githubCreateUserMigration = flagConfig.githubCreateUserMigration
		c.githubCreateUserMigrationRetry = flagConfig.githubCreateUserMigrationRetry
		c.githubCreateUserMigrationRetryMax = flagConfig.githubCreateUserMigrationRetryMax
		c.githubListUserMigrations = flagConfig.githubListUserMigrations
		c.githubWaitForMigrationComplete = flagConfig.githubWaitForMigrationComplete

		// Parse namespace whitelist if explicitly set
		if len(githubNamespaceWhitelistString) > 0 {
			c.githubNamespaceWhitelist = strings.Split(githubNamespaceWhitelistString, ",")
		}
	} else {
		// No config file — use flags directly (original behavior)
		c = flagConfig
		if len(githubNamespaceWhitelistString) > 0 {
			c.githubNamespaceWhitelist = strings.Split(githubNamespaceWhitelistString, ",")
		}
	}

	c.backupDir = setupBackupDir(&c.backupDir, &c.service, &c.gitHostURL)
	return &c, nil
}

// validateConfig validates the configuration and returns an error if invalid
func validateConfig(c *appConfig) error {
	if _, ok := knownServices[c.service]; !ok {
		return errors.New("please specify the git service type: github, gitlab, bitbucket, forgejo")
	}

	if !validGitlabProjectMembership(c.gitlabProjectMembershipType) {
		return errors.New("please specify a valid gitlab project membership - all/owner/member/starred")
	}
	return nil
}
