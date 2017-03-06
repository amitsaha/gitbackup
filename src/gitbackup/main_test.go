package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"testing"
)

// Help from https://npf.io/2015/06/testing-exec-command/

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestBackup(t *testing.T) {
	var wg sync.WaitGroup
	execCommand := fakeExecCommand
	defer func() {
		execCommand = exec.Command
		wg.Wait()
	}()
	backupDir := "/tmp/backupdir"
	repo := Repository{Name: "testrepo", GitURL: "git://foo.com/foo"}
	wg.Add(1)
	backUp(backupDir, &repo, &wg)
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// some code here to check arguments perhaps?
	fmt.Fprintf(os.Stdout, "cloned")
	os.Exit(0)
}
