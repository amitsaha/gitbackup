package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"sync"
	"testing"

	"github.com/spf13/afero"
)

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

func fakeRemoteUpdateCommand(command string, args ...string) (cmd *exec.Cmd) {
	cs := []string{"-test.run=TestHelperRemoteUpdateProcess", "--", command}
	cs = append(cs, args...)
	cmd = exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestBackup(t *testing.T) {
	var wg sync.WaitGroup
	repo := Repository{Name: "testrepo", CloneURL: "git://foo.com/foo"}
	backupDir := "/tmp/backupdir"

	// Memory FS
	appFS = afero.NewMemMapFs()
	appFS.MkdirAll(backupDir, 0771)

	defer func() {
		execCommand = exec.Command
		wg.Wait()
	}()

	// Test clone
	execCommand = fakeCloneCommand
	wg.Add(1)
	stdoutStderr, err := backUp(backupDir, &repo, false, &wg)
	if err != nil {
		t.Errorf("%s", stdoutStderr)
	}

	// Test pull
	repoDir := path.Join(backupDir, repo.Name)
	appFS.MkdirAll(repoDir, 0771)
	execCommand = fakePullCommand
	wg.Add(1)
	stdoutStderr, err = backUp(backupDir, &repo, false, &wg)
	if err != nil {
		t.Errorf("%s", stdoutStderr)
	}
}

func TestBareBackup(t *testing.T) {
	var wg sync.WaitGroup
	repo := Repository{Name: "testrepo", CloneURL: "git://foo.com/foo"}
	backupDir := "/tmp/backupdir"

	// Memory FS
	appFS = afero.NewMemMapFs()
	appFS.MkdirAll(backupDir, 0771)

	defer func() {
		execCommand = exec.Command
		wg.Wait()
	}()

	// Test clone
	execCommand = fakeCloneCommand
	wg.Add(1)
	stdoutStderr, err := backUp(backupDir, &repo, true, &wg)
	if err != nil {
		t.Errorf("%s", stdoutStderr)
	}

	// Test pull
	repoDir := path.Join(backupDir, repo.Name+".git")
	appFS.MkdirAll(repoDir, 0771)
	execCommand = fakeRemoteUpdateCommand
	wg.Add(1)
	stdoutStderr, err = backUp(backupDir, &repo, true, &wg)
	if err != nil {
		t.Errorf("%s", stdoutStderr)
	}
}

func TestHelperPullProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// Check that git command was executed
	if os.Args[3] != "git" || os.Args[6] != "pull" {
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

func TestHelperRemoteUpdateProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// Check that git command was executed
	if os.Args[3] != "git" || os.Args[6] != "remote" || os.Args[7] != "update" {
		fmt.Fprintf(os.Stdout, "Expected git remote update to be executed. Got %v", os.Args[3:])
		os.Exit(1)
	}
	os.Exit(0)
}

func TestSetupBackupDir(t *testing.T) {

	// test implementation of homedir.Dir()
	gethomeDir = func() (string, error) {
		return "/home/fakeuser", nil
	}
	appFS = afero.NewMemMapFs()

	backupRoot := "/my/backup/root"

	serviceGithubCustomUrl := "https://company.github.com"
	serviceGitlabCustomUrl := "https://company.gitlab.com"

	var testConfigs = []struct {
		backupRootDir  string
		gitService     string
		gitServiceUrl  string
		wantBackupPath string
	}{
		{
			"",
			"github",
			"",
			"/home/fakeuser/.gitbackup/github.com",
		},
		{
			backupRoot,
			"github",
			"",
			"/my/backup/root/github.com",
		},
		{
			backupRoot,
			"github",
			serviceGithubCustomUrl,
			"/my/backup/root/company.github.com",
		},
		{
			"",
			"gitlab",
			"",
			"/home/fakeuser/.gitbackup/gitlab.com",
		},

		{
			backupRoot,
			"gitlab",
			"",
			"/my/backup/root/gitlab.com",
		},
		{
			backupRoot,
			"gitlab",
			serviceGitlabCustomUrl,
			"/my/backup/root/company.gitlab.com",
		},
		{
			backupRoot,
			"bitbucket",
			"",
			"/my/backup/root/bitbucket.org",
		},
		{
			"",
			"bitbucket",
			"",
			"/home/fakeuser/.gitbackup/bitbucket.org",
		},
	}

	for _, tc := range testConfigs {
		backupdir := setupBackupDir(&tc.backupRootDir, &tc.gitService, &tc.gitServiceUrl)
		if backupdir != tc.wantBackupPath {
			t.Errorf("Expected %s, Got %s", tc.wantBackupPath, backupdir)
		}
		_, err := appFS.Stat(backupdir)
		if err != nil {
			t.Error(err)
		}
	}
}
