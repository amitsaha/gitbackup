package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-github/v34/github"
	githubmock "github.com/migueleliasweb/go-github-mock/src/mock"
)

type requestCounter struct {
	mutex             sync.Mutex
	cnt               int
	originalTransport http.RoundTripper
}

func (c *requestCounter) RoundTrip(r *http.Request) (*http.Response, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cnt += 1
	resp, err := c.originalTransport.RoundTrip(r)
	return resp, err
}

func TestCreateGitHubUserMigrationRetryMax(t *testing.T) {
	expectedNumAttempts := defaultMaxUserMigrationRetry + 1

	mockedHTTPClient := githubmock.NewMockedHTTPClient(
		githubmock.WithRequestMatchHandler(
			githubmock.PostUserMigrations,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				githubmock.WriteError(
					w,
					http.StatusBadGateway,
					"github 502",
				)
			}),
		),
	)
	requestCounter := requestCounter{}
	requestCounter.originalTransport = mockedHTTPClient.Transport
	mockedHTTPClient.Transport = &requestCounter

	c := github.NewClient(mockedHTTPClient)

	ctx := context.Background()

	_, _ = createGithubUserMigration(ctx, c, nil, true, defaultMaxUserMigrationRetry)
	if requestCounter.cnt != expectedNumAttempts {
		t.Fatalf("Expected:%d attempts, got: %d\n", expectedNumAttempts, requestCounter.cnt)
	}
}

func TestCreateGitHubUserMigrationFailOnceThenSucceed(t *testing.T) {
	expectedNumAttempts := 2
	mockRepoName := "mock-repo-1"

	mockedHTTPClient := githubmock.NewMockedHTTPClient(
		githubmock.WithRequestMatch(
			githubmock.PostUserMigrations,
			"rubbish_1",
			github.UserMigration{
				Repositories: []*github.Repository{
					{
						Name: &mockRepoName,
					},
				},
			},
		),
	)
	requestCounter := requestCounter{}
	requestCounter.originalTransport = mockedHTTPClient.Transport
	mockedHTTPClient.Transport = &requestCounter

	c := github.NewClient(mockedHTTPClient)
	ctx := context.Background()

	reposToMigrate := []*Repository{
		{
			Name:      "mock-repo-1",
			Namespace: "test-user-1",
		},
	}

	m, err := createGithubUserMigration(ctx, c, reposToMigrate, true, defaultMaxUserMigrationRetry)
	if err != nil {
		t.Fatal(err)
	}

	if len(m.Repositories) != len(reposToMigrate) {
		t.Fatalf("Expected %d repos in the migration. Got: %d", len(reposToMigrate), len(m.Repositories))
	}
	if requestCounter.cnt != expectedNumAttempts {
		t.Fatalf("Expected to send %d requests, sent: %d\n", defaultMaxUserMigrationRetry+1, requestCounter.cnt)
	}
}

func TestDownloadGithubUserMigrationDataFailed(t *testing.T) {
	var mockMigrationID int64 = 10021
	backUpDir := t.TempDir()

	mockedHTTPClient := githubmock.NewMockedHTTPClient(
		githubmock.WithRequestMatch(
			githubmock.GetUserMigrationsByMigrationId,
			github.UserMigration{
				ID:    &mockMigrationID,
				State: &migrationStatePending,
			},
			github.UserMigration{
				ID:    &mockMigrationID,
				State: &migrationStateFailed,
			},
		),
	)

	c := github.NewClient(mockedHTTPClient)
	ctx := context.Background()
	err := downloadGithubUserMigrationData(ctx, c, backUpDir, &mockMigrationID, 10*time.Millisecond)
	if err == nil {
		t.Fatalf("Expected migration download to fail.")
	}
}

func TestDownloadGithubUserMigrationDataArchiveDownloadFail(t *testing.T) {
	var mockMigrationID int64 = 10021
	backupDir := t.TempDir()

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
				http.Redirect(w, r, "http://127.0.0.1:8080/testarchive.tar.gz", http.StatusTemporaryRedirect)
			}),
		),
	)

	c := github.NewClient(mockedHTTPClient)
	ctx := context.Background()
	err := downloadGithubUserMigrationData(ctx, c, backupDir, &mockMigrationID, 10*time.Millisecond)
	if err == nil {
		t.Fatalf("Expected migration archive download to fail.")
	}
	if !strings.HasPrefix(err.Error(), "error downloading archive") {
		t.Fatalf("Expected error message to start with: error downloading archive, got: %v", err)
	}
}

func TestDownloadGithubUserMigrationDataArchiveDownload(t *testing.T) {
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
