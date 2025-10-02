package main

import (
	"log"
	"net/url"
	"os/exec"
	"path"
	"sync"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
)

// We have them here so that we can override these in the tests
var execCommand = exec.Command
var appFS = afero.NewOsFs()
var gitCommand = "git"
var gethomeDir = homedir.Dir

// Check if we have a copy of the repo already, if
// we do, we update the repo, else we do a fresh clone
func backUp(backupDir string, repo *Repository, bare bool, wg *sync.WaitGroup) ([]byte, error) {
	defer wg.Done()

	repoDir := getRepoDir(backupDir, repo, bare)

	_, err := appFS.Stat(repoDir)

	var stdoutStderr []byte
	if err == nil {
		stdoutStderr, err = updateExistingRepo(repoDir, repo.Name, bare)
	} else {
		stdoutStderr, err = cloneNewRepo(repoDir, repo, bare)
	}
	return stdoutStderr, err
}

// getRepoDir returns the directory path for a repository
func getRepoDir(backupDir string, repo *Repository, bare bool) string {
	var dirName string
	if bare {
		dirName = repo.Name + ".git"
	} else {
		dirName = repo.Name
	}
	return path.Join(backupDir, repo.Namespace, dirName)
}

// updateExistingRepo updates an existing repository
func updateExistingRepo(repoDir, repoName string, bare bool) ([]byte, error) {
	log.Printf("%s exists, updating. \n", repoName)
	var cmd *exec.Cmd
	if bare {
		cmd = execCommand(gitCommand, "-C", repoDir, "remote", "update", "--prune")
	} else {
		cmd = execCommand(gitCommand, "-C", repoDir, "pull")
	}
	return cmd.CombinedOutput()
}

// cloneNewRepo clones a new repository
func cloneNewRepo(repoDir string, repo *Repository, bare bool) ([]byte, error) {
	log.Printf("Cloning %s\n", repo.Name)
	log.Printf("%#v\n", repo)

	if repo.Private && ignorePrivate != nil && *ignorePrivate {
		log.Printf("Skipping %s as it is a private repo.\n", repo.Name)
		return nil, nil
	}

	cloneURL := repo.CloneURL
	if useHTTPSClone != nil && *useHTTPSClone {
		// Add username and token to the clone URL
		// https://gitlab.com/amitsaha/testproject1 => https://amitsaha:token@gitlab.com/amitsaha/testproject1
		u, err := url.Parse(repo.CloneURL)
		if err != nil {
			log.Fatalf("Invalid clone URL: %v\n", err)
		}
		cloneURL = u.Scheme + "://" + gitHostUsername + ":" + gitHostToken + "@" + u.Host + u.Path
	}

	var cmd *exec.Cmd
	if bare {
		cmd = execCommand(gitCommand, "clone", "--mirror", cloneURL, repoDir)
	} else {
		cmd = execCommand(gitCommand, "clone", cloneURL, repoDir)
	}
	return cmd.CombinedOutput()
}

func setupBackupDir(backupDir, service, githostURL *string) string {
	var gitHost, backupPath string
	var err error

	if len(*githostURL) != 0 {
		u, err := url.Parse(*githostURL)
		if err != nil {
			panic(err)
		}
		gitHost = u.Host
	} else {
		gitHost = knownServices[*service]
	}

	if len(*backupDir) == 0 {
		homeDir, err := gethomeDir()
		if err == nil {
			backupPath = path.Join(homeDir, ".gitbackup", gitHost)
		} else {
			log.Fatal("Could not determine home directory and backup directory not specified")
		}
	} else {
		backupPath = path.Join(*backupDir, gitHost)
	}

	err = createBackupRootDirIfRequired(backupPath)
	if err != nil {
		log.Fatalf("Error creating backup directory: %s %v", backupPath, err)
	}
	return backupPath
}

func createBackupRootDirIfRequired(backupPath string) error {
	return appFS.MkdirAll(backupPath, 0771)
}
