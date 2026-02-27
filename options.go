package main

import (
	"errors"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

// appFlags returns the CLI flags for the backup command.
func appFlags() []cli.Flag {
	return []cli.Flag{
		// Config file flag
		&cli.StringFlag{
			Name:  "config",
			Usage: "Path to config file (default: OS config directory)",
		},

		// Generic flags
		&cli.StringFlag{
			Name:  "service",
			Usage: "Git Hosted Service Name (github/gitlab/bitbucket/forgejo)",
		},
		&cli.StringFlag{
			Name:  "githost.url",
			Usage: "DNS of the custom Git host",
		},
		&cli.StringFlag{
			Name:  "backupdir",
			Usage: "Backup directory",
		},
		&cli.BoolFlag{
			Name:  "ignore-private",
			Usage: "Ignore private repositories/projects",
		},
		&cli.BoolFlag{
			Name:  "ignore-fork",
			Usage: "Ignore repositories which are forks",
		},
		&cli.BoolFlag{
			Name:  "use-https-clone",
			Usage: "Use HTTPS for cloning instead of SSH",
		},
		&cli.BoolFlag{
			Name:  "bare",
			Usage: "Clone bare repositories",
		},

		// GitHub specific flags
		&cli.StringFlag{
			Name:        "github.repoType",
			Usage:       "Repo types to backup (all, owner, member, starred)",
			DefaultText: "all",
			Value:       "all",
		},
		&cli.StringFlag{
			Name:  "github.namespaceWhitelist",
			Usage: "Organizations/Users from where we should clone (separate each value by a comma: 'user1,org2')",
		},
		&cli.BoolFlag{
			Name:  "github.createUserMigration",
			Usage: "Download user data",
		},
		&cli.BoolFlag{
			Name:        "github.createUserMigrationRetry",
			Usage:       "Retry creating the GitHub user migration if we get an error",
			DefaultText: "true",
			Value:       true,
		},
		&cli.IntFlag{
			Name:        "github.createUserMigrationRetryMax",
			Usage:       "Number of retries to attempt for creating GitHub user migration",
			DefaultText: "5",
			Value:       defaultMaxUserMigrationRetry,
		},
		&cli.BoolFlag{
			Name:  "github.listUserMigrations",
			Usage: "List available user migrations",
		},
		&cli.BoolFlag{
			Name:        "github.waitForUserMigration",
			Usage:       "Wait for migration to complete",
			DefaultText: "true",
			Value:       true,
		},

		// Gitlab specific flags
		&cli.StringFlag{
			Name:        "gitlab.projectVisibility",
			Usage:       "Visibility level of Projects to clone (internal, public, private)",
			DefaultText: "internal",
			Value:       "internal",
		},
		&cli.StringFlag{
			Name:        "gitlab.projectMembershipType",
			Usage:       "Project type to clone (all, owner, member, starred)",
			DefaultText: "all",
			Value:       "all",
		},

		// Forgejo specific flags
		&cli.StringFlag{
			Name:        "forgejo.repoType",
			Usage:       "Repo types to backup (user, starred)",
			DefaultText: "user",
			Value:       "user",
		},
	}
}

// buildConfig builds an appConfig from the CLI context, respecting config file precedence.
// If a config file exists, its values are used as the base and only explicitly-set
// CLI flags override them.
func buildConfig(cCtx *cli.Context) (*appConfig, error) {
	configPath := cCtx.String("config")

	// Try to load config file as the base configuration
	var c appConfig
	configFileLoaded := false

	resolvedPath, pathErr := resolveConfigPath(configPath)
	if pathErr == nil {
		if _, err := os.Stat(resolvedPath); err == nil {
			fc, err := loadConfigFile(configPath)
			if err != nil {
				return nil, err
			}
			c = *fileConfigToAppConfig(fc)
			configFileLoaded = true
		}
	}

	if configFileLoaded {
		// Only override config file values with flags that were explicitly set
		if cCtx.IsSet("service") {
			c.service = cCtx.String("service")
		}
		if cCtx.IsSet("githost.url") {
			c.gitHostURL = cCtx.String("githost.url")
		}
		if cCtx.IsSet("backupdir") {
			c.backupDir = cCtx.String("backupdir")
		}
		if cCtx.IsSet("ignore-private") {
			c.ignorePrivate = cCtx.Bool("ignore-private")
		}
		if cCtx.IsSet("ignore-fork") {
			c.ignoreFork = cCtx.Bool("ignore-fork")
		}
		if cCtx.IsSet("use-https-clone") {
			c.useHTTPSClone = cCtx.Bool("use-https-clone")
		}
		if cCtx.IsSet("bare") {
			c.bare = cCtx.Bool("bare")
		}
		if cCtx.IsSet("github.repoType") {
			c.githubRepoType = cCtx.String("github.repoType")
		}
		if cCtx.IsSet("github.namespaceWhitelist") {
			ns := cCtx.String("github.namespaceWhitelist")
			if len(ns) > 0 {
				c.githubNamespaceWhitelist = strings.Split(ns, ",")
			}
		}
		if cCtx.IsSet("gitlab.projectVisibility") {
			c.gitlabProjectVisibility = cCtx.String("gitlab.projectVisibility")
		}
		if cCtx.IsSet("gitlab.projectMembershipType") {
			c.gitlabProjectMembershipType = cCtx.String("gitlab.projectMembershipType")
		}
		if cCtx.IsSet("forgejo.repoType") {
			c.forgejoRepoType = cCtx.String("forgejo.repoType")
		}

		// Migration flags are always from CLI (not in config file)
		c.githubCreateUserMigration = cCtx.Bool("github.createUserMigration")
		c.githubCreateUserMigrationRetry = cCtx.Bool("github.createUserMigrationRetry")
		c.githubCreateUserMigrationRetryMax = cCtx.Int("github.createUserMigrationRetryMax")
		c.githubListUserMigrations = cCtx.Bool("github.listUserMigrations")
		c.githubWaitForMigrationComplete = cCtx.Bool("github.waitForUserMigration")
	} else {
		// No config file — read all values from CLI context directly
		c.service = cCtx.String("service")
		c.gitHostURL = cCtx.String("githost.url")
		c.backupDir = cCtx.String("backupdir")
		c.ignorePrivate = cCtx.Bool("ignore-private")
		c.ignoreFork = cCtx.Bool("ignore-fork")
		c.useHTTPSClone = cCtx.Bool("use-https-clone")
		c.bare = cCtx.Bool("bare")
		c.githubRepoType = cCtx.String("github.repoType")
		c.gitlabProjectVisibility = cCtx.String("gitlab.projectVisibility")
		c.gitlabProjectMembershipType = cCtx.String("gitlab.projectMembershipType")
		c.forgejoRepoType = cCtx.String("forgejo.repoType")
		c.githubCreateUserMigration = cCtx.Bool("github.createUserMigration")
		c.githubCreateUserMigrationRetry = cCtx.Bool("github.createUserMigrationRetry")
		c.githubCreateUserMigrationRetryMax = cCtx.Int("github.createUserMigrationRetryMax")
		c.githubListUserMigrations = cCtx.Bool("github.listUserMigrations")
		c.githubWaitForMigrationComplete = cCtx.Bool("github.waitForUserMigration")

		ns := cCtx.String("github.namespaceWhitelist")
		if len(ns) > 0 {
			c.githubNamespaceWhitelist = strings.Split(ns, ",")
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
