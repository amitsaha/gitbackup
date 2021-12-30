package main

import (
	"bytes"
	"context"
	"fmt"
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

func TestGetUserOwnedOrganizationsNone(t *testing.T) {

	mockedHTTPClient := githubmock.NewMockedHTTPClient(
		githubmock.WithRequestMatch(
			githubmock.GetUserMembershipsOrgs,
			[]github.Membership{},
		),
	)

	c := github.NewClient(mockedHTTPClient)
	ctx := context.Background()
	orgs, err := getGithubUserOwnedOrgs(ctx, c)
	if err != nil {
		t.Fatalf("Expected to query user organizations successfully, got: %v", err)
	}
	if len(orgs) != 0 {
		t.Fatalf("Expected slice of length 0, got %v", orgs)
	}
}

func TestGetUserOwnedOrganizations(t *testing.T) {

	testOrgNames := []string{
		"test-org-1",
		"test-org-2",
		"test-org-3",
	}

	mockedHTTPClient := githubmock.NewMockedHTTPClient(
		githubmock.WithRequestMatch(
			githubmock.GetUserMembershipsOrgs,
			[]github.Membership{
				{
					Organization: &github.Organization{
						Name: &testOrgNames[0],
					},
					Role: &orgRoleAdmin,
				},
				{
					Organization: &github.Organization{
						Name: &testOrgNames[1],
					},
					Role: &orgRoleMember,
				},
				{
					Organization: &github.Organization{
						Name: &testOrgNames[2],
					},
					Role: &orgRoleMaintainer,
				},
			},
		),
	)

	c := github.NewClient(mockedHTTPClient)
	ctx := context.Background()
	orgs, err := getGithubUserOwnedOrgs(ctx, c)
	if err != nil {
		t.Fatalf("Expected to query user organizations successfully, got: %v", err)
	}
	if len(orgs) != 1 {
		t.Fatalf("Expected slice of length 0, got %#v", orgs)
	}
	if *orgs[0].Name != testOrgNames[0] {
		t.Fatalf("Expected owned organization returned to be %s, got %s", testOrgNames[0], *orgs[0].Name)
	}
}

func TestGetGithubOrganizationRepositories(t *testing.T) {

	testOrgName := "test org 1"
	testOrgLogin := "test-org-1"
	testRepoName := "test-repo-1"
	testRepoFullname := fmt.Sprintf("%s/%s", testOrgLogin, testRepoName)
	testRepoHTTPSCloneURL := "https://github.com/test-org-1/test-repo-1.git"
	testRepoSSHCloneURL := "git@github.com:test-org-1/test-repo-1.git"
	testRepoTypePrivate := false

	testOrg := github.Organization{
		Name:  &testOrgName,
		Login: &testOrgLogin,
	}

	mockedHTTPClient := githubmock.NewMockedHTTPClient(
		githubmock.WithRequestMatch(
			githubmock.GetOrgsReposByOrg,
			[]github.Repository{
				{
					Name:     &testRepoName,
					FullName: &testRepoFullname,
					CloneURL: &testRepoHTTPSCloneURL,
					SSHURL:   &testRepoSSHCloneURL,
					Private:  &testRepoTypePrivate,
				},
			},
		),
	)

	c := github.NewClient(mockedHTTPClient)
	ctx := context.Background()
	repos, err := getGithubOrgRepositories(ctx, c, &testOrg)
	if err != nil {
		t.Fatalf("Expected to query user organization repositories successfully, got: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("Expected slice of length 0, got %#v", repos)
	}
	if repos[0].Name != testRepoName {
		t.Fatalf("Expected returned repo name to be %s, got %s", testRepoName, repos[0].Name)
	}
	if repos[0].Private != testRepoTypePrivate {
		t.Fatalf("Expected %v, got %v", testRepoTypePrivate, repos[0].Private)
	}
}
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
	var testMigrationID int64 = 10021
	backupDir := t.TempDir()
	testOrg := "TestOrg"

	mockedHTTPClient := githubmock.NewMockedHTTPClient(
		githubmock.WithRequestMatch(
			githubmock.GetOrgsMigrationsByOrgByMigrationId,
			github.Migration{
				ID:    &testMigrationID,
				State: &migrationStatePending,
			},
			github.Migration{
				ID:    &testMigrationID,
				State: &migrationStateFailed,
			},
		),
	)

	c := github.NewClient(mockedHTTPClient)
	ctx := context.Background()
	err := downloadGithubOrgMigrationData(ctx, c, testOrg, backupDir, &testMigrationID, 10*time.Millisecond)
	if err == nil {
		t.Fatalf("Expected migration download to fail.")
	}
}

func TestDownloadGithubUserOrgMigrationDataArchiveDownloadFail(t *testing.T) {
	var testMigrationID int64 = 10021
	backupDir := t.TempDir()
	testOrg := "TestOrg"

	mockedHTTPClient := githubmock.NewMockedHTTPClient(
		githubmock.WithRequestMatch(
			githubmock.GetOrgsMigrationsByOrgByMigrationId,
			github.Migration{
				ID:    &testMigrationID,
				State: &migrationStatePending,
			},
			github.Migration{
				ID:    &testMigrationID,
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
	err := downloadGithubOrgMigrationData(ctx, c, testOrg, backupDir, &testMigrationID, 10*time.Millisecond)
	if err == nil {
		t.Fatalf("Expected migration archive download to fail.")
	}
	if !strings.HasPrefix(err.Error(), "error downloading archive") {
		t.Fatalf("Expected error message to start with: error downloading archive, got: %v", err)
	}
}

func TestDownloadGithubUserOrgMigrationDataArchiveDownload(t *testing.T) {
	var testMigrationID int64 = 10021
	testOrg := "TestOrg"
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
			githubmock.GetOrgsMigrationsByOrgByMigrationId,
			github.Migration{
				ID:    &testMigrationID,
				State: &migrationStatePending,
			},
			github.Migration{
				ID:    &testMigrationID,
				State: &migrationStateExported,
			},
		),
		githubmock.WithRequestMatchHandler(
			githubmock.GetOrgsMigrationsArchiveByOrgByMigrationId,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, ts.URL+"/testarchive.tar.gz", http.StatusTemporaryRedirect)
			}),
		),
	)

	c := github.NewClient(mockedHTTPClient)
	ctx := context.Background()
	err := downloadGithubOrgMigrationData(ctx, c, testOrg, backupDir, &testMigrationID, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("Expected migration archive download to succeed. Got %v", err)
	}
	archiveFilepath := getLocalOrgMigrationFilepath(backupDir, testOrg, testMigrationID)
	_, err = os.Stat(archiveFilepath)
	if err != nil {
		t.Fatalf("Expected %s to exist", archiveFilepath)
	}
}
