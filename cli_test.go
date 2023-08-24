package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestCliUsage(t *testing.T) {
	binaryFilename := "gitbackup_test_bin"
	goldenFilepath := path.Join("testdata", t.Name()+".golden")
	if runtime.GOOS == "windows" {
		binaryFilename = "gitbackup_test_bin.exe"
		goldenFilepath = goldenFilepath + ".windows"
	}

	cmd := exec.Command("go", "build", "-o", binaryFilename)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Error building test binary: %v - %v", err, string(stdoutStderr))
	}
	defer func() {
		err := os.Remove(binaryFilename)
		if err != nil {
			t.Fatal(err)
		}
	}()

	var stdout, stderr bytes.Buffer
	goldenFilepathNew := goldenFilepath + ".expected"

	cmd = exec.Command("./"+binaryFilename, "-h")
	t.Log(cmd.String())
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
	expectedUsageString := string(expectedUsage)
	// For windows
	expectedUsage = []byte(strings.ReplaceAll(expectedUsageString, "\r\n", "\n"))

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
