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

func main() {
	client := github.NewClient(nil)
	// Make these configurable
	username := flag.String("username", "", "GitHub username")
	homeDir, dirErr := homedir.Dir()
	backupdir := flag.String("backupdir", path.Join(homeDir, ".ghbackup"), "Backup directory")
	flag.Parse()

	if dirErr != nil && len(*backupdir) == 0 {
		log.Fatal("Couldn't retrieve your home directory. You must specify a backup directory")
	}

	// TODO: Check permissions for backup directory

	if len(*username) == 0 {
		log.Fatal("Please specify your GitHub username")
	}

	repoType := "all"
	BACKUP_DIR := *backupdir
	opt := &github.RepositoryListOptions{Type: repoType, Sort: "created", Direction: "desc"}
	for {
		repos, resp, err := client.Repositories.List(*username, opt)
		if err != nil {
			fmt.Println(err)
		} else {
			// default to ~/.ghbackup as the backup directory
			_, err := os.Stat(BACKUP_DIR)
			if err != nil {
				fmt.Printf("%s doesn't exist, creating it\n", BACKUP_DIR)
				err := os.Mkdir(BACKUP_DIR, 0771)
				if err != nil {
					log.Fatal(err)
				}
			}
			err = os.Chdir(BACKUP_DIR)
			if err != nil {
				log.Fatal(err)
			}
			for _, repo := range repos {
				// Check if we have a copy of the repo already, if
				// we do, we update the repo, else we do a fresh clone
				_, err := os.Stat(*repo.Name)
				if err == nil {
					fmt.Printf("%v exists, updating. \n", *repo.Name)
					err := os.Chdir(*repo.Name)
					cmd := exec.Command("git", "pull", *repo.GitURL)
					err = cmd.Run()
					if err != nil {
						fmt.Printf("Error pulling %v\n", *repo.Name)
					}
					// Go one level up
					os.Chdir("..")
				} else {
					fmt.Printf("Cloning %v \n", *repo.Name)
					cmd := exec.Command("git", "clone", *repo.GitURL)
					err := cmd.Run()
					if err != nil {
						fmt.Printf("Error cloning %v: ", *repo.Name)
						log.Fatal(err)
					}
				}
			}
		}

		// Learn about GitHub's pagination here:
		// https://developer.github.com/v3/
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}
}
