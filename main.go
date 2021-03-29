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
var useHTTPSClone *bool
var ignorePrivate *bool
var gitHostUsername string

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
	ignorePrivate = flag.Bool("ignore-private", false, "Ignore private repositories/projects")
	ignoreFork := flag.Bool("ignore-fork", false, "Ignore repositories which are forks")
	useHTTPSClone = flag.Bool("use-https-clone", false, "Use HTTPS for cloning instead of SSH")

	// GitHub specific flags
	githubRepoType := flag.String("github.repoType", "all", "Repo types to backup (all, owner, member)")
	githubUserData := flag.Bool("github.userData", false, "Download user data")

	// Gitlab specific flags
	gitlabRepoVisibility := flag.String("gitlab.projectVisibility", "internal", "Visibility level of Projects to clone (internal, public, private)")
	gitlabProjectMembership := flag.String("gitlab.projectMembershipType", "all", "Project type to clone (all, owner, member)")

	flag.Parse()

	if len(*service) == 0 || !knownServices[*service] {
		log.Fatal("Please specify the git service type: github, gitlab")
	}

	if !validGitlabProjectMembership(*gitlabProjectMembership) {
		log.Fatal("Please specify a valid gitlab project membership - all/owner/member")
	}

	*backupDir = setupBackupDir(*backupDir, *service, *githostURL)

	client := newClient(*service, *githostURL)

	if *githubUserData {
		repos, err := getRepositories(client, *service, *githubRepoType, *gitlabRepoVisibility, *gitlabProjectMembership, *ignoreFork)
		if err != nil {
			log.Fatalf("Error getting list of repositories: %v", err)
		}
		getGithubUserData(client, *backupDir, repos)
	} else {
		tokens := make(chan bool, MaxConcurrentClones)
		gitHostUsername = getUsername(client, *service)

		if len(gitHostUsername) == 0 && !*ignorePrivate && *useHTTPSClone {
			log.Fatal("Your Git host's username is needed for backing up private repositories via HTTPS")
		}
		repos, err := getRepositories(client, *service, *githubRepoType, *gitlabRepoVisibility, *gitlabProjectMembership, *ignoreFork)
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
}
