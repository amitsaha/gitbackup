package main

import (
	"fmt"
	"log"
	"sync"
)

func handleGitRepositoryClone(client interface{}, c *appConfig) error {

	// Used for waiting for all the goroutines to finish before exiting
	var wg sync.WaitGroup
	defer wg.Wait()

	tokens := make(chan bool, MaxConcurrentClones)
	gitHostUsername = getUsername(client, c.service)

	if len(gitHostUsername) == 0 && !*ignorePrivate && *useHTTPSClone {
		log.Fatal("Your Git host's username is needed for backing up private repositories via HTTPS")
	}

	repositories, err := getRepositories(
		client,
		c.service,
		c.githubRepoType,
		c.githubNamespaceWhitelist,
		c.gitlabProjectVisibility,
		c.gitlabProjectMembershipType,
		c.ignoreFork,
	)
	if err != nil {
		return err
	}
	if len(repositories) == 0 {
		return fmt.Errorf("no repositories retrieved")
	}

	log.Printf("Backing up %v repositories now..\n", len(repositories))
	for _, repo := range repositories {
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
	return nil
}
