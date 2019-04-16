package main

import (
	"flag"
	"log"
	"sync"
)

// MaxConcurrentClones is the upper limit of the maximum number of
// concurrent git clones
var MaxConcurrentClones = 20

var gitHostToken string

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
	githostURL := flag.String("githost.url", "", "DNS of the custom Git host")
	backupDir := flag.String("backupdir", "", "Backup directory")

	// GitHub specific flags
	githubRepoType := flag.String("github.repoType", "all", "Repo types to backup (all, owner, member)")

	// Gitlab specific flags
	gitlabRepoVisibility := flag.String("gitlab.projectVisibility", "internal", "Visibility level of Projects to clone (internal, public, private)")
	gitlabProjectMembership := flag.String("gitlab.projectMembershipType", "all", "Project type to clone (all, owner, member)")

	flag.Parse()

	if len(*service) == 0 || !knownServices[*service] {
		log.Fatal("Please specify the git service type: github, gitlab")
	}
	*backupDir = setupBackupDir(*backupDir, *service, *githostURL)
	tokens := make(chan bool, MaxConcurrentClones)
	client := newClient(*service, *githostURL)
	repos, err := getRepositories(client, *service, *githubRepoType, *gitlabRepoVisibility, *gitlabProjectMembership)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Backing up %v repositories now..\n", len(repos))
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
