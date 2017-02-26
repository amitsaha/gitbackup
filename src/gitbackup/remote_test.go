package main

import (
	"github.com/google/go-github/github"
	"github.com/xanzy/go-gitlab"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("github")
	client = client.(*github.Client)

	client = NewClient("gitlab")
	client = client.(*gitlab.Client)
}
