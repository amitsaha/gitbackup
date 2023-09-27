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

	c, err := initConfig(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	err = validateConfig(c)
	if err != nil {
		log.Fatal(err)
	}

	client := newClient(c.service, c)
	var executionErr error

	// TODO implement validation of options so that we don't
	// allow multiple operations at one go
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
