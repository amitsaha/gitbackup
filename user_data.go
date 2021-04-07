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

func createGithubUserMigration(ctx context.Context, client interface{}, repos []*Repository) (*github.UserMigration, error) {
	migrationOpts := github.UserMigrationOptions{
		LockRepositories:   false,
		ExcludeAttachments: false,
	}
	var repoPaths []string
	for _, repo := range repos {
		repoPaths = append(repoPaths, fmt.Sprintf("%s/%s", repo.Namespace, repo.Name))
	}

	m, resp, err := client.(*github.Client).Migrations.StartUserMigration(ctx, repoPaths, &migrationOpts)
	if err != nil {
		defer resp.Body.Close()
		data, _ := ioutil.ReadAll(resp.Body)
		log.Printf("%v", string(data))
	}
	return m, err
}

func downloadGithubUserData(client interface{}, backupDir string, id *int64) {

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

func getUserOrgs(client interface{}) ([]*github.Organization, error) {
	ctx := context.Background()
	orgs, _, err := client.(*github.Client).Organizations.List(ctx, "", nil)
	return orgs, err
}

func getGithubOrgRepositories(client interface{}, orgs []*github.Organization) ([]*Repository, error) {

	var repositories []*Repository
	var cloneURL string

	githubRepoType := "owner"
	ignoreFork := true

	ctx := context.Background()
	options := github.RepositoryListOptions{Type: githubRepoType}
	for {
		repos, resp, err := client.(*github.Client).Repositories.List(ctx, "", &options)
		if err == nil {
			for _, repo := range repos {
				if *repo.Fork && ignoreFork {
					continue
				}
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
