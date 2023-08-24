package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path"
	"reflect"
	"testing"
)

func TestCliUsage(t *testing.T) {
	cmd := exec.Command("go", "build", "-o", "gitbackup_test_bin")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Error building test binary: %v - %v", err, string(stdoutStderr))
	}
	defer func() {
		err := os.Remove("gitbackup_test_bin")
		if err != nil {
			t.Fatal(err)
		}
	}()

	var stdout, stderr bytes.Buffer
	goldenFilepath := path.Join("testdata", t.Name()+".golden")
	goldenFilepathNew := goldenFilepath + ".expected"

	cmd = exec.Command("./gitbackup_test_bin", "-h")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		log.Println(string(stderr.Bytes()))
		t.Fatal(err)
	}

	gotUsage := stderr.Bytes()

	expectedUsage, err := os.ReadFile(goldenFilepath)
	if err != nil {
		t.Errorf("couldn't read %[1]s..writing expected output to %[1]s", goldenFilepath)
		if err := writeExpectedGoldenFile(goldenFilepath, gotUsage); err != nil {
			t.Fatal("Error writing file", err)
		}
		t.FailNow()
	}
	if !reflect.DeepEqual(expectedUsage, gotUsage) {
		t.Errorf("expected and got data mismatch. ..writing expected output to %[1]s", goldenFilepathNew)
		if err := writeExpectedGoldenFile(goldenFilepathNew, gotUsage); err != nil {
			t.Fatal("Error writing file", err)
		}
		t.FailNow()
	}
}

func writeExpectedGoldenFile(filePath string, data []byte) error {
	return os.WriteFile(filePath, data, 0644)
}
