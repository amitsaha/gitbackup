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

// Check if we have a copy of the repo already, if
// we do, we update the repo, else we do a fresh clone
func backUp(backupDir string, repo *Repository, wg *sync.WaitGroup) {
	defer wg.Done()

	repoDir := path.Join(backupDir, repo.Name)
	_, err := os.Stat(repoDir)

	if err == nil {
		log.Printf("%v exists, updating. \n", repo.Name)
		cmd := exec.Command("git", "-C", repoDir, "pull")
		err = cmd.Run()
		if err != nil {
			log.Printf("Error pulling %v: %v\n", repo.GitURL, err)
		}
	} else {
		log.Printf("Cloning %v \n", repo.Name)
		cmd := exec.Command("git", "clone", repo.GitURL, repoDir)
		err := cmd.Run()
		if err != nil {
			log.Printf("Error cloning %v: %v", repo.Name, err)
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
	gitlabUrl := flag.String("gitlab.url", "", "DNS of your the GitLab service")
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
	}

	opt := ListRepositoriesOptions{repoType: *githubRepoType, repoVisibility: *gitlabRepoVisibility}
	// Limit maximum concurrent clones to 20
	tokens := make(chan bool, 20)
	repos, err := getRepositories(*service, *gitlabUrl, &opt)
	if err != nil {
		log.Fatal(err)
	} else {
		_, err := os.Stat(*backupDir)
		if err != nil {
			log.Printf("%s doesn't exist, creating it\n", *backupDir)
			err := os.MkdirAll(*backupDir, 0771)
			if err != nil {
				log.Fatal(err)
			}
		}
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
