package main

import (
	"flag"
	"log"
	"sync"
)

// Maximum number of concurrent clones
var MAX_CONCURRENT_CLONES int = 20

func main() {

	// Used for waiting for all the goroutines to finish before exiting
	var wg sync.WaitGroup
	defer wg.Wait()

	// The services we know of
	knownServices := map[string]bool{
		"github": true,
		"gitlab": true,
	}

	// Generic flags
	service := flag.String("service", "", "Git Hosted Service Name (github/gitlab)")
	githostUrl := flag.String("githost.url", "", "DNS of the custom Git host")
	backupDir := flag.String("backupdir", "", "Backup directory")

	// GitHub specific flags
	githubRepoType := flag.String("github.repoType", "all", "Repo types to backup (all, owner, member)")

	// Gitlab specific flags
	gitlabRepoVisibility := flag.String("gitlab.projectVisibility", "internal", "Visibility level of Projects to clone")

	flag.Parse()

	if len(*service) == 0 || !knownServices[*service] {
		log.Fatal("Please specify the git service type: github, gitlab")
	}
	*backupDir = setupBackupDir(*backupDir, *service, *githostUrl)
	tokens := make(chan bool, MAX_CONCURRENT_CLONES)
	client := NewClient(*service, *githostUrl)
	repos, err := getRepositories(client, *service, *githubRepoType, *gitlabRepoVisibility)
	if err != nil {
		log.Fatal(err)
	} else {
		for _, repo := range repos {
			tokens <- true
			wg.Add(1)
			go func(repo *Repository) {
				stdoutStderr, err := backUp(*backupDir, repo, &wg)
				if err != nil {
					log.Printf("Error backing up %s: %s\n", repo.Name, stdoutStderr)
				}
				<-tokens
			}(repo)
		}
	}
}
