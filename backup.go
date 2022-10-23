package main

import (
	"log"
	"net/url"
	"os/exec"
	"path"
	"sync"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
)

// We have them here so that we can override these in the tests
var execCommand = exec.Command
var appFS = afero.NewOsFs()
var gitCommand = "git"

// Check if we have a copy of the repo already, if
// we do, we update the repo, else we do a fresh clone
func backUp(backupDir string, repo *Repository, bare bool, wg *sync.WaitGroup) ([]byte, error) {
	defer wg.Done()

	var dirName string
	if bare {
		dirName = repo.Name + ".git"
	} else {
		dirName = repo.Name
	}
	repoDir := path.Join(backupDir, repo.Namespace, dirName)

	_, err := appFS.Stat(repoDir)

	var stdoutStderr []byte
	if err == nil {
		log.Printf("%s exists, updating. \n", repo.Name)
		var cmd *exec.Cmd
		if bare {
			cmd = execCommand(gitCommand, "-C", repoDir, "remote", "update", "--prune")
		} else {
			cmd = execCommand(gitCommand, "-C", repoDir, "pull")
		}
		stdoutStderr, err = cmd.CombinedOutput()
	} else {
		log.Printf("Cloning %s\n", repo.Name)
		log.Printf("%#v\n", repo)

		if repo.Private && ignorePrivate != nil && *ignorePrivate {
			log.Printf("Skipping %s as it is a private repo.\n", repo.Name)
			return stdoutStderr, nil
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
		stdoutStderr, err = cmd.CombinedOutput()
	}
	return stdoutStderr, err
}

func setupBackupDir(backupDir, service, githostURL *string) string {
	var gitHost, backupPath string
	var err error

	if githostURL != nil {
		u, err := url.Parse(*githostURL)
		if err != nil {
			panic(err)
		}
		gitHost = u.Host
	} else {
		gitHost = knownServices[*service]
	}

	if backupDir == nil {
		homeDir, err := homedir.Dir()
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
		log.Fatal(err)
	}
	return backupPath
}

func createBackupRootDirIfRequired(backupPath string) error {
	return appFS.MkdirAll(backupPath, 0771)
}
