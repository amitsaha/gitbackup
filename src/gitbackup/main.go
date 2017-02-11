package main

import (
	"flag"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"log"
	"os"
	"os/exec"
	"path"
)

// TODO:
// 3. Other inline todos
// 1. Implement stats
// 2. Show progress
// 4. Don't shell out to git?

func backUp(backupDir string, repo *Repository) {
	// Check if we have a copy of the repo already, if
	// we do, we update the repo, else we do a fresh clone
	repoDir := path.Join(backupDir, repo.Name)
	_, err := os.Stat(repoDir)
	if err == nil {
		fmt.Printf("%v exists, updating. \n", repo.Name)
		cmd := exec.Command("git", "-C", repoDir, "pull")
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error pulling %v: %v\n", repo.GitURL, err)
		}
	} else {
		fmt.Printf("Cloning %v \n", repo.Name)
		cmd := exec.Command("git", "clone", repo.GitURL, repoDir)
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error cloning %v: %v", repo.Name, err)
		}
	}
}

func main() {
	// TODO: Make these configurable via config file
	service := flag.String("service", "", "Git Hosted Service Name (github/gitlab)")
	if len(*service) == 0 {
		log.Fatal("Please specify the git service type: github, gitlab")
	}
	// TODO:
	//serviceUrl := flag.String("gitlab-url", "", "DNS of the another GitLab service")
	username := flag.String("username", "", "GitHub username")
	homeDir, dirErr := homedir.Dir()
	backupDir := flag.String("backupdir", path.Join(homeDir, ".gitbackup", *service), "Backup directory")
	flag.Parse()

	if dirErr != nil && len(*backupDir) == 0 {
		log.Fatal("Couldn't retrieve your home directory. You must specify a backup directory")
	}

	// TODO: Check permissions for backup directory

	// Create an API client
	client := NewClient(nil, *service)
	if client == nil {
		log.Fatalf("Service %s not supported", *service)
	}

	// TODO: do away with username
	if *service == "github" && len(*username) == 0 {
		log.Fatal("Please specify your GitHub username")
	}

	// TODO: make these configurable via config file
	opt := ListRepositoriesOptions{repoType: "all", Sort: "created", Direction: "desc"}
	for {
		repos, resp, err := getRepositories(*service, client, *username, &opt)
		if err != nil {
			// TODO: Exit or continue
			log.Fatal(err)
		} else {
			_, err := os.Stat(*backupDir)
			if err != nil {
				fmt.Printf("%s doesn't exist, creating it\n", backupDir)
				err := os.Mkdir(*backupDir, 0771)
				if err != nil {
					log.Fatal(err)
				}
			}
			// Limit maximum concurrent operations to 20
			tokens := make(chan struct{}, 20)
			for _, repo := range repos {
				tokens <- struct{}{}
				go func(repo *Repository) {
					backUp(*backupDir, repo)
					<-tokens
				}(repo)
			}
			if resp.NextPage == 0 {
				break
			}
			opt.ListOptions.Page = resp.NextPage
		}
	}
}
