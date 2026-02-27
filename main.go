package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
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
	"forgejo":   "codeberg.org",
}

func main() {
	app := &cli.App{
		Name:  "gitbackup",
		Usage: "Backup your Git repositories from GitHub, GitLab, Bitbucket, or Forgejo",
		Flags: appFlags(),
		Action: func(cCtx *cli.Context) error {
			c, err := buildConfig(cCtx)
			if err != nil {
				return err
			}
			err = validateConfig(c)
			if err != nil {
				return err
			}

			client := newClient(c.service, c.gitHostURL)

			if c.githubListUserMigrations {
				handleGithubListUserMigrations(client, c)
			} else if c.githubCreateUserMigration {
				handleGithubCreateUserMigration(client, c)
			} else {
				if err := handleGitRepositoryClone(client, c); err != nil {
					return err
				}
			}
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "Create a default gitbackup.yml configuration file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "config",
						Usage: "Path to config file (default: OS config directory)",
					},
				},
				Action: func(cCtx *cli.Context) error {
					return handleInitConfig(cCtx.String("config"))
				},
			},
			{
				Name:  "validate",
				Usage: "Validate the gitbackup.yml configuration file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "config",
						Usage: "Path to config file (default: OS config directory)",
					},
				},
				Action: func(cCtx *cli.Context) error {
					return handleValidateConfig(cCtx.String("config"))
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		log.Fatal(err)
	}
}
