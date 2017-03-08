package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"sync"
	"testing"
)

// Help from https://npf.io/2015/06/testing-exec-command/

func fakePullCommand(command string, args ...string) (cmd *exec.Cmd) {
	cs := []string{"-test.run=TestHelperPullProcess", "--", command}
	cs = append(cs, args...)
	cmd = exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func fakeCloneCommand(command string, args ...string) (cmd *exec.Cmd) {
	cs := []string{"-test.run=TestHelperCloneProcess", "--", command}
	cs = append(cs, args...)
	cmd = exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestBackup(t *testing.T) {
	var wg sync.WaitGroup
	backupDir := "/tmp/backupdir"
	defer func() {
		execCommand = exec.Command
		wg.Wait()
		// Cleanup backupDir
		os.RemoveAll(backupDir)
	}()

	os.MkdirAll(backupDir, 0771)

	// Test clone
	execCommand = fakeCloneCommand
	repo := Repository{Name: "testrepo", GitURL: "git://foo.com/foo"}
	wg.Add(1)
	stdoutStderr, err := backUp(backupDir, &repo, &wg)
	if err != nil {
		log.Fatal("%s", stdoutStderr)
	}

	// Test pull
	repoDir := path.Join(backupDir, repo.Name)
	os.Mkdir(repoDir, 0771)
	execCommand = fakePullCommand
	wg.Add(1)
	stdoutStderr, err = backUp(backupDir, &repo, &wg)
	if err != nil {
		fmt.Printf("%s", stdoutStderr)
		os.Exit(1)
	}

}

func TestHelperPullProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// Check that git command was executed
	if os.Args[3] != "git" || os.Args[4] != "pull" {
		fmt.Fprintf(os.Stdout, "Expected git pull to be executed. Got %v", os.Args[3:])
		os.Exit(1)
	}
	os.Exit(0)
}

func TestHelperCloneProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// Check that git command was executed
	if os.Args[3] != "git" || os.Args[4] != "clone" {
		fmt.Fprintf(os.Stdout, "Expected git clone to be executed. Got %v", os.Args[3:])
		os.Exit(1)
	}
	os.Exit(0)
}
