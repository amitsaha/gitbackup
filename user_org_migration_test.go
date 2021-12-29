package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v34/github"
	githubmock "github.com/migueleliasweb/go-github-mock/src/mock"
)

func TestCreateGithubUserOrgMigration(t *testing.T) {
	testOrg := "TestOrg"
	testRepoName := "test-repo-1"
	orgRepos := []*Repository{
		{
			Name: testRepoName,
		},
	}

	mockedHTTPClient := githubmock.NewMockedHTTPClient(
		githubmock.WithRequestMatch(
			githubmock.PostOrgsMigrationsByOrg,
			github.Migration{
				Repositories: []*github.Repository{
					{
						Name: &testRepoName,
					},
				},
			},
		),
	)

	c := github.NewClient(mockedHTTPClient)
	ctx := context.Background()
	_, err := createGithubOrgMigration(ctx, c, testOrg, orgRepos)
	if err != nil {
		t.Fatalf("Expected org migration to be successfully created, got: %v", err)
	}
}

func TestDownloadGithubUserOrgMigrationDataFailed(t *testing.T) {
	var mockMigrationID int64 = 10021
	backupDir := t.TempDir()
	testOrg := "TestOrg"

	mockedHTTPClient := githubmock.NewMockedHTTPClient(
		githubmock.WithRequestMatch(
			githubmock.GetOrgsMigrationsByOrgByMigrationId,
			github.Migration{
				ID:    &mockMigrationID,
				State: &migrationStatePending,
			},
			github.Migration{
				ID:    &mockMigrationID,
				State: &migrationStateFailed,
			},
		),
	)

	c := github.NewClient(mockedHTTPClient)
	ctx := context.Background()
	err := downloadGithubOrgMigrationData(ctx, c, testOrg, backupDir, &mockMigrationID, 10*time.Millisecond)
	if err == nil {
		t.Fatalf("Expected migration download to fail.")
	}
}

func TestDownloadGithubUserOrgMigrationDataArchiveDownloadFail(t *testing.T) {
	var mockMigrationID int64 = 10021
	backupDir := t.TempDir()
	testOrg := "TestOrg"

	mockedHTTPClient := githubmock.NewMockedHTTPClient(
		githubmock.WithRequestMatch(
			githubmock.GetOrgsMigrationsByOrgByMigrationId,
			github.Migration{
				ID:    &mockMigrationID,
				State: &migrationStatePending,
			},
			github.Migration{
				ID:    &mockMigrationID,
				State: &migrationStateExported,
			},
		),
		githubmock.WithRequestMatchHandler(
			githubmock.GetOrgsMigrationsArchiveByOrgByMigrationId,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "http://127.0.0.1:8080/testarchive.tar.gz", http.StatusTemporaryRedirect)
			}),
		),
	)

	c := github.NewClient(mockedHTTPClient)
	ctx := context.Background()
	err := downloadGithubOrgMigrationData(ctx, c, testOrg, backupDir, &mockMigrationID, 10*time.Millisecond)
	if err == nil {
		t.Fatalf("Expected migration archive download to fail.")
	}
	if !strings.HasPrefix(err.Error(), "error downloading archive") {
		t.Fatalf("Expected error message to start with: error downloading archive, got: %v", err)
	}
}

func TestDownloadGithubUserOrgMigrationDataArchiveDownload(t *testing.T) {
	var mockMigrationID int64 = 10021
	backupDir := t.TempDir()

	mux := http.NewServeMux()
	mux.HandleFunc("/testarchive.tar.gz", func(w http.ResponseWriter, r *http.Request) {
		b := bytes.NewBuffer([]byte("testdata"))
		r.Header.Set("Content-Type", "application/gzip")
		io.Copy(w, b)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	mockedHTTPClient := githubmock.NewMockedHTTPClient(
		githubmock.WithRequestMatch(
			githubmock.GetUserMigrationsByMigrationId,
			github.UserMigration{
				ID:    &mockMigrationID,
				State: &migrationStatePending,
			},
			github.UserMigration{
				ID:    &mockMigrationID,
				State: &migrationStateExported,
			},
		),
		githubmock.WithRequestMatchHandler(
			githubmock.GetUserMigrationsArchiveByMigrationId,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, ts.URL+"/testarchive.tar.gz", http.StatusTemporaryRedirect)
			}),
		),
	)

	c := github.NewClient(mockedHTTPClient)
	ctx := context.Background()
	err := downloadGithubUserMigrationData(ctx, c, backupDir, &mockMigrationID, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("Expected migration archive download to succeed.")
	}
	archiveFilepath := getLocalMigrationFilepath(backupDir, mockMigrationID)
	_, err = os.Stat(archiveFilepath)
	if err != nil {
		t.Fatalf("Expected %s to exist", archiveFilepath)
	}
}
