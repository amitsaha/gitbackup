package main

import (
	"log"
	"os"
	"strings"

	bitbucket "github.com/ktrysmt/go-bitbucket"
)

func getBitbucketRepositories(
	client *bitbucket.Client,
	ignoreFork bool,
) ([]*Repository, error) {

	// As of April 14, 2026 Atlassian removed the cross-workspace listing
	// endpoints (/2.0/workspaces and /2.0/user/permissions/workspaces) under
	// changelog entries CHANGE-2770 / CHANGE-3022. There is no supported way
	// to enumerate the workspaces a user belongs to programmatically. The
	// caller must supply the workspace slugs via BITBUCKET_WORKSPACES
	// (comma-separated).
	workspacesEnv := os.Getenv("BITBUCKET_WORKSPACES")
	if workspacesEnv == "" {
		log.Fatal("BITBUCKET_WORKSPACES environment variable not set (comma-separated workspace slugs)")
	}

	var repositories []*Repository

	for _, slug := range strings.Split(workspacesEnv, ",") {
		slug = strings.TrimSpace(slug)
		if slug == "" {
			continue
		}

		options := &bitbucket.RepositoriesOptions{Owner: slug}

		resp, err := client.Repositories.ListForAccount(options)
		if err != nil {
			return nil, err
		}

		for _, repo := range resp.Items {
			if repo.Parent != nil && ignoreFork {
				continue
			}
			namespace := strings.Split(repo.Full_name, "/")[0]

			httpsURL, sshURL := extractBitbucketCloneURLs(repo.Links)
			cloneURL := getCloneURL(httpsURL, sshURL)

			repositories = append(repositories, &Repository{
				CloneURL:  cloneURL,
				Name:      repo.Slug,
				Namespace: namespace,
				Private:   repo.Is_private,
			})
		}
	}
	return repositories, nil
}

func extractBitbucketCloneURLs(links map[string]interface{}) (httpsURL, sshURL string) {
	linkmaps, ok := links["clone"].([]interface{})
	if !ok {
		return "", ""
	}

	for _, linkmaps := range linkmaps {
		linkmap, ok := linkmaps.(map[string]interface{})
		if !ok {
			continue
		}

		if linkmap["name"] == "https" {
			httpsURL = linkmap["href"].(string)
		}

		if linkmap["name"] == "ssh" {
			sshURL = linkmap["href"].(string)
		}
	}
	return httpsURL, sshURL
}
