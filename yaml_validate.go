package main

import (
	"fmt"
	"os"
)

func validateYAMLConfig(path string) error {
	c, err := LoadYAMLConfig(path)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	// Validate required fields
	if _, ok := knownServices[c.service]; !ok {
		return fmt.Errorf("invalid service: %s", c.service)
	}
	// Check env vars for secrets (example: GITHUB_TOKEN)
	if c.service == "github" {
		if os.Getenv("GITHUB_TOKEN") == "" {
			return fmt.Errorf("GITHUB_TOKEN environment variable not set")
		}
	}
	if c.service == "gitlab" {
		if os.Getenv("GITLAB_TOKEN") == "" {
			return fmt.Errorf("GITLAB_TOKEN environment variable not set")
		}
	}
	if c.service == "bitbucket" {
		if os.Getenv("BITBUCKET_TOKEN") == "" {
			return fmt.Errorf("BITBUCKET_TOKEN environment variable not set")
		}
	}
	return nil
}
