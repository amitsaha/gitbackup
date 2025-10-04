package main

import (
	"log"
	"os"
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

func main() {

	args := os.Args[1:]
	if len(args) > 0 {
		switch args[0] {
		case "create-config":
			handleCreateConfig(args[1:])
			return
		case "validate-config":
			handleValidateConfig(args[1:])
			return
		case "help", "--help", "-h":
			printCreateConfigUsage()
			return
		}
	}

	var c *appConfig
	if _, err := os.Stat("gitbackup.yml"); err == nil {
		c, err = LoadYAMLConfig("gitbackup.yml")
		if err != nil {
			log.Fatalf("Failed to load gitbackup.yml: %v", err)
		}
		if err := validateYAMLConfig("gitbackup.yml"); err != nil {
			log.Fatalf("Config validation failed: %v", err)
		}
	} else {
		c, err = initConfig(args)
		if err != nil {
			log.Fatal(err)
		}
		err = validateConfig(c)
		if err != nil {
			log.Fatal(err)
		}
	}

	client := newClient(c.service, c.gitHostURL)
	var executionErr error

	if c.githubListUserMigrations {
		handleGithubListUserMigrations(client, c)
	} else if c.githubCreateUserMigration {
		handleGithubCreateUserMigration(client, c)
	} else {
		executionErr = handleGitRepositoryClone(client, c)
	}
	if executionErr != nil {
		log.Fatal(executionErr)
	}
}
