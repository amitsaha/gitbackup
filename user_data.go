package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/go-github/v34/github"
)

func createGithubUserMigration(ctx context.Context, client interface{}, repos []*Repository, retry bool, maxNumRetries int) (*github.UserMigration, error) {
	var m *github.UserMigration
	var err error
	var resp *github.Response

	migrationOpts := github.UserMigrationOptions{
		LockRepositories:   false,
		ExcludeAttachments: false,
	}
	var repoPaths []string
	for _, repo := range repos {
		repoPaths = append(repoPaths, fmt.Sprintf("%s/%s", repo.Namespace, repo.Name))
	}

	numAttempts := 1
	if retry {
		numAttempts += maxNumRetries
	}

	var errResponse []byte
	for i := 1; i <= numAttempts; i++ {
		m, resp, err = client.(*github.Client).Migrations.StartUserMigration(ctx, repoPaths, &migrationOpts)
		if err == nil {
			return m, nil
		}
		if resp != nil {
			errResponse, _ = ioutil.ReadAll(resp.Body)
			resp.Body.Close()
		}
		log.Printf("Attempt #%d: Error creating user migration: %v", i, string(errResponse))
	}
	return m, err
}

func createGithubOrgMigration(ctx context.Context, client interface{}, org string, repos []*Repository) (*github.Migration, error) {
	migrationOpts := github.MigrationOptions{
		LockRepositories:   false,
		ExcludeAttachments: false,
	}
	var repoPaths []string
	for _, repo := range repos {
		repoPaths = append(repoPaths, fmt.Sprintf("%s/%s", repo.Namespace, repo.Name))
	}

	m, resp, err := client.(*github.Client).Migrations.StartMigration(ctx, org, repoPaths, &migrationOpts)
	if err != nil {
		defer resp.Body.Close()
		data, _ := ioutil.ReadAll(resp.Body)
		log.Printf("%v", string(data))
	}
	return m, err
}

func downloadGithubUserMigrationData(client interface{}, backupDir string, id *int64) {

	var ms *github.UserMigration
	ctx := context.Background()

	ms, _, err := client.(*github.Client).Migrations.UserMigrationStatus(ctx, *id)
	if err != nil {
		panic(err)
	}

	for {

		if *ms.State == "failed" {
			log.Fatal("Migration failed.")
		}
		if *ms.State == "exported" {
			archiveURL, err := client.(*github.Client).Migrations.UserMigrationArchiveURL(ctx, *ms.ID)
			if err != nil {
				panic(err)
			}

			archiveFilepath := path.Join(backupDir, fmt.Sprintf("user-migration-%d.tar.gz", *ms.ID))
			log.Printf("Downloading file to: %s\n", archiveFilepath)

			resp, err := http.Get(archiveURL)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			out, err := os.Create(archiveFilepath)
			if err != nil {
				log.Fatal(err)
			}
			defer out.Close()

			_, err = io.Copy(out, resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			break
		} else {
			log.Printf("Waiting for migration state to be exported: %v\n", ms.State)
			time.Sleep(60 * time.Second)

			ms, _, err = client.(*github.Client).Migrations.UserMigrationStatus(ctx, *ms.ID)
			if err != nil {
				panic(err)
			}
		}
	}
}

func downloadGithubOrgMigrationData(client interface{}, org string, backupDir string, id *int64) {

	var ms *github.Migration
	ctx := context.Background()

	ms, _, err := client.(*github.Client).Migrations.MigrationStatus(ctx, org, *id)
	if err != nil {
		panic(err)
	}

	for {

		if *ms.State == "failed" {
			log.Fatal("Migration failed.")
		}
		if *ms.State == "exported" {
			archiveURL, err := client.(*github.Client).Migrations.MigrationArchiveURL(ctx, org, *ms.ID)
			if err != nil {
				panic(err)
			}

			archiveFilepath := path.Join(backupDir, fmt.Sprintf("%s-migration-%d.tar.gz", org, *ms.ID))
			log.Printf("Downloading file to: %s\n", archiveFilepath)

			resp, err := http.Get(archiveURL)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			out, err := os.Create(archiveFilepath)
			if err != nil {
				log.Fatal(err)
			}
			defer out.Close()

			_, err = io.Copy(out, resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			break
		} else {
			log.Printf("Waiting for migration state to be exported: %v\n", ms.State)
			time.Sleep(60 * time.Second)

			ms, _, err = client.(*github.Client).Migrations.MigrationStatus(ctx, org, *ms.ID)
			if err != nil {
				panic(err)
			}
		}
	}
}

// ListGithubUserMigrationsResult type is for listing migration result
type ListGithubUserMigrationsResult struct {
	GUID  *string `json:"guid"`
	ID    *int64  `json:"id"`
	State *string `json:"state"`
}

// List Github user migrations
func getGithubUserMigrations(client interface{}) ([]ListGithubUserMigrationsResult, error) {

	ctx := context.Background()
	migrations, _, err := client.(*github.Client).Migrations.ListUserMigrations(ctx)

	if err != nil {
		return nil, err
	}

	var result []ListGithubUserMigrationsResult
	for _, m := range migrations {

		r := ListGithubUserMigrationsResult{}
		r.GUID = m.GUID
		r.ID = m.ID
		r.State = m.State

		result = append(result, r)
	}

	return result, nil
}

// GetGithubUserMigration to Get the status of a migration
func GetGithubUserMigration(client interface{}, id *int64) (*github.UserMigration, error) {
	ctx := context.Background()
	ms, _, err := client.(*github.Client).Migrations.UserMigrationStatus(ctx, *id)
	return ms, err
}

// GithubUserMigrationDeleteResult is a type for deletion result
type GithubUserMigrationDeleteResult struct {
	GhStatusCode   int    `json:"status_code"`
	GhResponseBody string `json:"mesage"`
}

// DeleteGithubUserMigration deletes an existing migration
func DeleteGithubUserMigration(id *int64) GithubUserMigrationDeleteResult {
	client := newClient("github", "https://github.com")
	ctx := context.Background()
	response, err := client.(*github.Client).Migrations.DeleteUserMigration(ctx, *id)

	result := GithubUserMigrationDeleteResult{}
	result.GhStatusCode = response.StatusCode

	if err != nil {
		result.GhResponseBody = err.Error()
	} else {

		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			panic(err)
		}
		result.GhResponseBody = string(data)
	}
	return result
}

func getUserOwnedOrgs(client interface{}) ([]*github.Organization, error) {

	var ownedOrgs []*github.Organization

	ctx := context.Background()
	opts := github.ListOrgMembershipsOptions{State: "active"}
	mShips, _, err := client.(*github.Client).Organizations.ListOrgMemberships(ctx, &opts)
	//TODO - if the user doesn't belong to any org, what happens?
	if err != nil {
		return nil, err
	}
	for _, m := range mShips {
		if *m.Role == "admin" {
			ownedOrgs = append(ownedOrgs, m.Organization)
		}
	}
	return ownedOrgs, nil
}

func getGithubOrgRepositories(client interface{}, o *github.Organization) ([]*Repository, error) {

	var repositories []*Repository
	var cloneURL string

	ctx := context.Background()
	// TODO: Allow customization for org repo types
	options := github.RepositoryListByOrgOptions{}

	for {
		// Login seems to be the safer attribute to use than organization Name
		repos, resp, err := client.(*github.Client).Repositories.ListByOrg(ctx, *o.Login, &options)
		if err == nil {
			for _, repo := range repos {
				namespace := strings.Split(*repo.FullName, "/")[0]
				if useHTTPSClone != nil && *useHTTPSClone {
					cloneURL = *repo.CloneURL
				} else {
					cloneURL = *repo.SSHURL
				}
				repositories = append(repositories, &Repository{CloneURL: cloneURL, Name: *repo.Name, Namespace: namespace, Private: *repo.Private})
			}
		} else {
			return nil, err
		}
		if resp.NextPage == 0 {
			break
		}
		options.ListOptions.Page = resp.NextPage

	}
	return repositories, nil
}
