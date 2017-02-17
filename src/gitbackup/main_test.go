package main

import (
    "testing"
    "sync"
)

func TestBackup(t *testing.T) {
    backupDir := "/tmp/backupdir"
    repo := Repository{Name: "testrepo", GitURL: "git://foo.com/foo"}
	var wg sync.WaitGroup
	defer wg.Wait()
    wg.Add(1)
    //TODO: https://golang.org/src/os/exec/exec_test.go
    backUp(backupDir, &repo , &wg)
}
