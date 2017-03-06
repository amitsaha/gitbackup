package main

import (
	"log"
	"os"
	"os/exec"
	"sync"
	"testing"
)

// Help from https://npf.io/2015/06/testing-exec-command/

func fakeExecCommand(command string, args ...string) (cmd *exec.Cmd) {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd = exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestBackup(t *testing.T) {
	var wg sync.WaitGroup
	execCommand = fakeExecCommand
	defer func() {
		execCommand = exec.Command
		wg.Wait()
	}()
	backupDir := "/tmp/backupdir"
	repo := Repository{Name: "testrepo", GitURL: "git://foo.com/foo"}
	wg.Add(1)
	err := backUp(backupDir, &repo, &wg)
	if err != nil {
		log.Fatal(err)
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// Check that git command was executed
	if os.Args[3] != "git" {
		log.Fatal("Expected git to be executed. Got %v", os.Args[3:])
	}
	os.Exit(0)
}
