package main

import (
	"flag"
	"fmt"
	"github.com/google/go-github/github"
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

func backUp(backupDir string, repo *github.Repository) {
	// Check if we have a copy of the repo already, if
	// we do, we update the repo, else we do a fresh clone
	repoDir := path.Join(backupDir, *repo.Name)
	_, err := os.Stat(repoDir)
	if err == nil {
		fmt.Printf("%v exists, updating. \n", *repo.Name)
		cmd := exec.Command("git", "-C", repoDir, "pull")
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error pulling %v: %v\n", *repo.GitURL, err)
		}
	} else {
		fmt.Printf("Cloning %v \n", *repo.Name)
		cmd := exec.Command("git", "clone", *repo.GitURL, repoDir)
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error cloning %v: %v", *repo.Name, err)
		}
	}
}

func main() {
	client := github.NewClient(nil)
	// TODO: Make these configurable via config file
	username := flag.String("username", "", "GitHub username")
	homeDir, dirErr := homedir.Dir()
	backupDir := flag.String("backupdir", path.Join(homeDir, ".ghbackup"), "Backup directory")
	flag.Parse()

	if dirErr != nil && len(*backupDir) == 0 {
		log.Fatal("Couldn't retrieve your home directory. You must specify a backup directory")
	}

	// TODO: Check permissions for backup directory

	if len(*username) == 0 {
		log.Fatal("Please specify your GitHub username")
	}

	// TODO: make these configurable via config file
	repoType := "all"
	opt := &github.RepositoryListOptions{Type: repoType, Sort: "created", Direction: "desc"}
	for {
		repos, resp, err := client.Repositories.List(*username, opt)
		if err != nil {
			fmt.Println(err)
		} else {
			// default to ~/.ghbackup as the backup directory
			_, err := os.Stat(*backupDir)
			if err != nil {
				fmt.Printf("%s doesn't exist, creating it\n", backupDir)
				err := os.Mkdir(*backupDir, 0771)
				if err != nil {
					log.Fatal(err)
				}
			}
			tokens := make(chan struct{}, 20)
			for _, repo := range repos {
				//TODO: Modify this to allow certain number of max concurrent
				// operations
				tokens <- struct{}{}
				go func(repo *github.Repository) {
					backUp(*backupDir, repo)
					<-tokens
				}(repo)
			}
			// Learn about GitHub's pagination here:
			// https://developer.github.com/v3/
			if resp.NextPage == 0 {
				break
			}
			opt.ListOptions.Page = resp.NextPage
		}
	}
}
