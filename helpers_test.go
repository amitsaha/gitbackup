package main

import (
	"testing"
)

func TestGetCloneURL(t *testing.T) {
	httpsURL := "https://github.com/user/repo.git"
	sshURL := "git@github.com:user/repo.git"

	// Test when useHTTPSClone is nil (default to SSH)
	useHTTPSClone = nil
	result := getCloneURL(httpsURL, sshURL)
	if result != sshURL {
		t.Errorf("Expected SSH URL when useHTTPSClone is nil, got: %s", result)
	}

	// Test when useHTTPSClone is false
	falseVal := false
	useHTTPSClone = &falseVal
	result = getCloneURL(httpsURL, sshURL)
	if result != sshURL {
		t.Errorf("Expected SSH URL when useHTTPSClone is false, got: %s", result)
	}

	// Test when useHTTPSClone is true
	trueVal := true
	useHTTPSClone = &trueVal
	result = getCloneURL(httpsURL, sshURL)
	if result != httpsURL {
		t.Errorf("Expected HTTPS URL when useHTTPSClone is true, got: %s", result)
	}

	// Reset to nil for other tests
	useHTTPSClone = nil
}

func TestContains(t *testing.T) {
	list := []string{"apple", "banana", "cherry"}

	if !contains(list, "banana") {
		t.Error("Expected contains to return true for 'banana'")
	}

	if contains(list, "grape") {
		t.Error("Expected contains to return false for 'grape'")
	}

	if contains([]string{}, "anything") {
		t.Error("Expected contains to return false for empty list")
	}
}

func TestValidGitlabProjectMembership(t *testing.T) {
	validMemberships := []string{"all", "owner", "member", "starred"}

	for _, m := range validMemberships {
		if !validGitlabProjectMembership(m) {
			t.Errorf("Expected validGitlabProjectMembership to return true for '%s'", m)
		}
	}

	invalidMemberships := []string{"invalid", "", "ALL", "Owner"}
	for _, m := range invalidMemberships {
		if validGitlabProjectMembership(m) {
			t.Errorf("Expected validGitlabProjectMembership to return false for '%s'", m)
		}
	}
}
