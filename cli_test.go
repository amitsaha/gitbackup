package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestCli(t *testing.T) {
	cmd := exec.Command("go", "build", "-o", "gitbackup_test_bin")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := os.Remove("gitbackup_test_bin")
		if err != nil {
			t.Fatal(err)
		}
	}()
	cmd = exec.Command("./gitbackup_test_bin", "-h")
	stdoutStderr, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(stdoutStderr))
}
