package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/google/go-github/v34/github"
)

func createGithubUserMigration(ctx context.Context, client *github.Client, repos []string) (*github.UserMigration, error) {
	migrationOpts := github.UserMigrationOptions{
		LockRepositories:   false,
		ExcludeAttachments: false,
	}
	m, _, err := client.Migrations.StartUserMigration(ctx, repos, &migrationOpts)
	return m, err
}

func getGithubUserData(client interface{}, backupDir string, repos []*Repository) {

	var ms *github.UserMigration
	ctx := context.Background()

	var repoPaths []string
	for _, repo := range repos {
		repoPaths = append(repoPaths, fmt.Sprintf("%s/%s", repo.Namespace, repo.Name))
	}
	m, err := createGithubUserMigration(ctx, client.(*github.Client), repoPaths)
	if err != nil {
		log.Fatal(err)
	}

	ms, _, err = client.(*github.Client).Migrations.UserMigrationStatus(ctx, *m.ID)
	if err != nil {
		panic(err)
	}

	var result GithubUserMigrationState
	result.ID = ms.ID
	result.State = ms.State
	result.CreatedAt = ms.CreatedAt
	result.UpdatedAt = ms.UpdatedAt

	for {

		// URL can be null since GitHub only keeps the archive URL for 7 days
		if *ms.State == "exported" {
			log.Printf("Migration state: %v\n", ms.State)
			archiveURL, err := client.(*github.Client).Migrations.UserMigrationArchiveURL(ctx, *ms.ID)
			if err != nil {
				panic(err)
			}
			result.ArchiveURL = &archiveURL
			parsedURL, _ := url.Parse(archiveURL)
			archiveFilepath := path.Join(backupDir, parsedURL.EscapedPath())
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
			time.Sleep(30 * time.Second)

			ms, _, err = client.(*github.Client).Migrations.UserMigrationStatus(ctx, *m.ID)
			if err != nil {
				panic(err)
			}
		}
	}
}

type ListGithubUserMigrationsResult struct {
	GUID  *string `json:"guid"`
	ID    *int64  `json:"id"`
	State *string `json:"state"`
}

// List Github user migrations
func getGithubUserMigrations(client *github.Client) []ListGithubUserMigrationsResult {

	ctx := context.Background()
	migrations, _, err := client.Migrations.ListUserMigrations(ctx)

	if err != nil {
		panic(err)
	}

	var result []ListGithubUserMigrationsResult
	for _, m := range migrations {

		r := ListGithubUserMigrationsResult{}
		r.GUID = m.GUID
		r.ID = m.ID
		r.State = m.State

		result = append(result, r)
	}

	return result
}

type GithubUserMigrationState struct {
	ID         *int64  `json:"id"`
	CreatedAt  *string `json:"created_at"`
	UpdatedAt  *string `json:"updated_at"`
	State      *string `json:"state"`
	ArchiveURL *string `json:"archive_url"`
}

// Get the status of a migration
func GetGithubUserMigration(id *int64) GithubUserMigrationState {
	client := newClient("github", "https://github.com")
	ctx := context.Background()
	ms, _, err := client.(*github.Client).Migrations.UserMigrationStatus(ctx, *id)

	if err != nil {
		panic(err)
	}
	var result GithubUserMigrationState
	result.ID = ms.ID
	result.State = ms.State
	result.CreatedAt = ms.CreatedAt
	result.UpdatedAt = ms.UpdatedAt

	// URL can be null since GitHub only keeps the archive URL for 7 days
	if *ms.State == "exported" {
		url, err := client.(*github.Client).Migrations.UserMigrationArchiveURL(ctx, *id)
		if err != nil {
			panic(err)
		}
		result.ArchiveURL = &url
	}

	return result
}

type GithubUserMigrationDeleteResult struct {
	GhStatusCode   int    `json:"status_code"`
	GhResponseBody string `json:"mesage"`
}

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
