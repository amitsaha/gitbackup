package main

import (
	"flag"
	"github.com/mitchellh/go-homedir"
	"log"
	"os"
	"os/exec"
	"path"
	"sync"
)

// Maximum number of concurrent clones
var MAX_CONCURRENT_CLONES int = 20

var execCommand = exec.Command

// Check if we have a copy of the repo already, if
// we do, we update the repo, else we do a fresh clone
func backUp(backupDir string, repo *Repository, wg *sync.WaitGroup) {
	defer wg.Done()

	repoDir := path.Join(backupDir, repo.Name)
	_, err := os.Stat(repoDir)

	if err == nil {
		log.Printf("%s exists, updating. \n", repo.Name)
		cmd := execCommand("git", "-C", repoDir, "pull")
		err = cmd.Run()
		if err != nil {
			log.Printf("Error pulling %s: %v\n", repo.GitURL, err)
		}
	} else {
		log.Printf("Cloning %s \n", repo.Name)
		cmd := execCommand("git", "clone", repo.GitURL, repoDir)
		err := cmd.Run()
		if err != nil {
			log.Printf("Error cloning %s: %v", repo.Name, err)
		}
	}
}

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
	backupDir := flag.String("backupdir", "", "Backup directory")

	// GitHub specific flags
	githubRepoType := flag.String("github.repoType", "all", "Repo types to backup (all, owner, member)")

	// Gitlab specific flags
	gitlabUrl := flag.String("gitlab.url", "", "DNS of the GitLab service")
	gitlabRepoVisibility := flag.String("gitlab.projectVisibility", "internal", "Visibility level of Projects to clone")

	flag.Parse()

	// If gitlab.url is specified, we know it's gitlab
	if len(*gitlabUrl) != 0 {
		*service = "gitlab"
	} else {
		// Either service or gitlab.url should be specified
		if len(*service) == 0 || !knownServices[*service] {
			log.Fatal("Please specify the git service type: github, gitlab")
		}
	}
	// Default backup directory, if none specified
	if len(*backupDir) == 0 {
		homeDir, err := homedir.Dir()
		if err == nil {
			*backupDir = path.Join(homeDir, ".gitbackup", *service)
		} else {
			log.Fatal("Could not determine home directory and backup directory not specified")
		}
	} else {
		*backupDir = path.Join(*backupDir, *service)
	}
	_, err := os.Stat(*backupDir)
	if err != nil {
		log.Printf("%s doesn't exist, creating it\n", *backupDir)
		err := os.MkdirAll(*backupDir, 0771)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Limit maximum concurrent clones to MAX_CONCURRENT_CLONES
	tokens := make(chan bool, MAX_CONCURRENT_CLONES)
	repos, err := getRepositories(*service, *gitlabUrl, *githubRepoType, *gitlabRepoVisibility)
	if err != nil {
		log.Fatal(err)
	} else {
		for _, repo := range repos {
			tokens <- true
			wg.Add(1)
			go func(repo *Repository) {
				backUp(*backupDir, repo, &wg)
				<-tokens
			}(repo)
		}
	}
}
