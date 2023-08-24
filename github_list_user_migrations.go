package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/v34/github"
)

func handleGithubListUserMigrations(client interface{}, c *appConfig) {

	mList, err := getGithubUserMigrations(client)
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range mList {
		mData, err := GetGithubUserMigration(client, m.ID)
		if err != nil {
			fmt.Printf("Error getting migration data: %v", *m.ID)
			// FIXME
			continue
		}

		var archiveURL string
		_, err = client.(*github.Client).Migrations.UserMigrationArchiveURL(context.Background(), *m.ID)
		if err != nil {
			archiveURL = "No Longer Available"
		} else {
			archiveURL = "Available for Download"
		}
		fmt.Printf("%v - %v - %v - %v\n", *mData.ID, *mData.CreatedAt, *mData.State, archiveURL)
	}
}
