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

func updateExistingClone(repoDir string, bare bool, repo *Repository) ([]byte, error) {
	log.Printf("%s exists, updating. \n", repo.Name)
	var cmd *exec.Cmd
	if bare {
		cmd = execCommand(gitCommand, "-C", repoDir, "remote", "update", "--prune")
	} else {
		cmd = execCommand(gitCommand, "-C", repoDir, "pull")
	}
	return cmd.CombinedOutput()
}

func newClone(repoDir string, bare bool, repo *Repository, useHTTPSClone *bool) ([]byte, error) {

	log.Printf("Cloning %s\n", repo.Name)
	log.Printf("%#v\n", repo)

	if repo.Private && ignorePrivate != nil && *ignorePrivate {
		log.Printf("Skipping %s as it is a private repo.\n", repo.Name)
		return nil, nil
	}

	if useHTTPSClone != nil && *useHTTPSClone {
		// Add username and token to the clone URL
		// https://gitlab.com/amitsaha/testproject1 => https://amitsaha:token@gitlab.com/amitsaha/testproject1
		u, err := url.Parse(repo.CloneURL)
		if err != nil {
			log.Fatalf("Invalid clone URL: %v\n", err)
		}
		repo.CloneURL = u.Scheme + "://" + gitHostUsername + ":" + gitHostToken + "@" + u.Host + u.Path
	}

	var cmd *exec.Cmd
	if bare {
		cmd = execCommand(gitCommand, "clone", "--mirror", repo.CloneURL, repoDir)
	} else {
		cmd = execCommand(gitCommand, "clone", repo.CloneURL, repoDir)
	}
	return cmd.CombinedOutput()
}

// Check if we have a copy of the repo already, if
// we do, we update the repo, else we do a fresh clone
func backUp(backupDir string, repo *Repository, bare bool, wg *sync.WaitGroup) ([]byte, error) {
	defer wg.Done()

	dirName := repo.Name
	if bare {
		dirName += ".git"
	}
	repoDir := path.Join(backupDir, repo.Namespace, dirName)
	_, err := appFS.Stat(repoDir)
	if err == nil {
		return updateExistingClone(repoDir, bare, repo)
	}
	return newClone(repoDir, bare, repo, useHTTPSClone)
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
