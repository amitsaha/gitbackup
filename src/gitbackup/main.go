package main

import (
	"flag"
	"github.com/mitchellh/go-homedir"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"sync"
)

// TODO:
// 3. Other inline todos
// 1. Implement stats
// 2. Show progress
// 4. Don't shell out to git?

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
	var wg sync.WaitGroup
	defer wg.Wait()

	// Generic flags
	service := flag.String("service", "", "Git Hosted Service Name (github/gitlab)")
	homeDir, dirErr := homedir.Dir()
	backupDir := flag.String("backupdir", "", "Backup directory (~/.gitbackup)")

	// Gitlab specific flags
	gitlabUrl := flag.String("gitlab.url", "", "DNS of your the GitLab service")
	gitlabRepoVisibility := flag.String("gitlab.repoVisibility", "private", "Visibility level of Repositories to clone")

	// GitHub specific flags
	githubUser := flag.String("github.username", "", "Your GitHub username")
	githubRepoType := flag.String("github.repoType", "all", "Repo types to backup (all, owner, member)")
	flag.Parse()

	// Either service or gitlab.url should be specified
	if len(*service) == 0 && len(*gitlabUrl) == 0 {
		log.Fatal("Please specify the git service type: github, gitlab")
	}
	// If gitlab.url is specified, we know it's gitlab
	var gitlabUrlPath *url.URL
	// TODO: fix
	var err error
	if len(*gitlabUrl) != 0 {
		gitlabUrlPath, err = url.Parse(*gitlabUrl)
		if err != nil {
			log.Fatal("Invalid gitlab URL: %s", *gitlabUrl)
		}
		*service = "gitlab"
	} else {
		gitlabUrlPath = nil
	}
	// If service is github, we need a username
	if *service == "github" && len(*githubUser) == 0 {
		log.Fatal("Please specify your GitHub githubUser")
	}

	// Default backup directory, if none specified
	if dirErr == nil && len(*backupDir) == 0 {
		*backupDir = path.Join(homeDir, ".gitbackup", *service)
	}

	// TODO: Check permissions for backup directory

	// Create an API client
	client := NewClient(nil, *service)
	if client == nil {
		log.Fatalf("Service %s not supported", *service)
	}

	opt := ListRepositoriesOptions{repoType: *githubRepoType, repoVisibility: *gitlabRepoVisibility}
	for {
		repos, resp, err := getRepositories(*service, gitlabUrlPath, client, *githubUser, &opt)
		if err != nil {
			// TODO: Currently exits on a first error
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
			// Limit maximum concurrent operations to 20
			// TODO: verify
			tokens := make(chan struct{}, 20)
			for _, repo := range repos {
				tokens <- struct{}{}
				wg.Add(1)
				go func(repo *Repository) {
					backUp(*backupDir, repo, &wg)
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
