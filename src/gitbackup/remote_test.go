package main

import (
	"github.com/google/go-github/github"
	"github.com/xanzy/go-gitlab"
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	os.Setenv("GITHUB_TOKEN", "$$$randome")
	client := NewClient("github")
	// Type assertion
	client = client.(*github.Client)

	os.Setenv("GITLAB_TOKEN", "$$$randome")
	client = NewClient("gitlab")
	// Type assertion
	client = client.(*gitlab.Client)

	client = NewClient("notyetsupported")
	if client != nil {
		t.Errorf("Expected nil")
	}

}
