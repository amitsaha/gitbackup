package main

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestHandleInitConfig(t *testing.T) {
	// Work in a temp directory
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// First call should create the file
	err := handleInitConfig()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	configPath := filepath.Join(tmpDir, defaultConfigFile)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Expected gitbackup.yml to be created")
	}

	// Verify the file is valid YAML that parses into fileConfig
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Error reading config file: %v", err)
	}

	var cfg fileConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		t.Fatalf("Error parsing config file: %v", err)
	}

	// Verify defaults match
	if cfg.Service != "github" {
		t.Errorf("Expected service to be 'github', got: %v", cfg.Service)
	}
	if cfg.GitHub.RepoType != "all" {
		t.Errorf("Expected github.repo_type to be 'all', got: %v", cfg.GitHub.RepoType)
	}
	if cfg.GitLab.ProjectVisibility != "internal" {
		t.Errorf("Expected gitlab.project_visibility to be 'internal', got: %v", cfg.GitLab.ProjectVisibility)
	}

	// Second call should error because file already exists
	err = handleInitConfig()
	if err == nil {
		t.Fatal("Expected error when config file already exists")
	}
}

func TestHandleValidateConfig(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Validate should fail if no config file exists
	err := handleValidateConfig()
	if err == nil {
		t.Fatal("Expected error when config file doesn't exist")
	}

	// Create a valid config and set the required env var
	err = handleInitConfig()
	if err != nil {
		t.Fatalf("Expected no error creating config, got: %v", err)
	}

	os.Setenv("GITHUB_TOKEN", "testtoken")
	defer os.Unsetenv("GITHUB_TOKEN")

	err = handleValidateConfig()
	if err != nil {
		t.Fatalf("Expected valid config, got: %v", err)
	}
}

func TestHandleValidateConfigInvalidService(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Write a config with an invalid service
	os.WriteFile(defaultConfigFile, []byte("service: notaservice\n"), 0644)

	err := handleValidateConfig()
	if err == nil {
		t.Fatal("Expected validation error for invalid service")
	}
}

func TestHandleValidateConfigMissingEnvVar(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Write a valid gitlab config but don't set GITLAB_TOKEN
	os.WriteFile(defaultConfigFile, []byte("service: gitlab\ngitlab:\n  project_visibility: internal\n  project_membership_type: all\n"), 0644)
	os.Unsetenv("GITLAB_TOKEN")

	err := handleValidateConfig()
	if err == nil {
		t.Fatal("Expected validation error for missing GITLAB_TOKEN")
	}
}

func TestInitConfigWithConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Write a config file with specific values
	os.WriteFile(defaultConfigFile, []byte("service: gitlab\nignore_fork: true\nuse_https_clone: true\ngitlab:\n  project_visibility: private\n  project_membership_type: owner\n"), 0644)
	os.Setenv("GITLAB_TOKEN", "testtoken")
	defer os.Unsetenv("GITLAB_TOKEN")

	// initConfig with no CLI flags should use config file values
	c, err := initConfig([]string{})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if c.service != "gitlab" {
		t.Errorf("Expected service 'gitlab', got: %v", c.service)
	}
	if !c.ignoreFork {
		t.Error("Expected ignore_fork to be true from config file")
	}
	if !c.useHTTPSClone {
		t.Error("Expected use_https_clone to be true from config file")
	}
	if c.gitlabProjectVisibility != "private" {
		t.Errorf("Expected project_visibility 'private', got: %v", c.gitlabProjectVisibility)
	}
	if c.gitlabProjectMembershipType != "owner" {
		t.Errorf("Expected project_membership_type 'owner', got: %v", c.gitlabProjectMembershipType)
	}
}

func TestInitConfigCLIOverridesConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Config file says gitlab with ignore_fork true
	os.WriteFile(defaultConfigFile, []byte("service: gitlab\nignore_fork: true\ngitlab:\n  project_visibility: private\n  project_membership_type: all\n"), 0644)
	os.Setenv("GITLAB_TOKEN", "testtoken")
	defer os.Unsetenv("GITLAB_TOKEN")

	// CLI flag overrides service to github
	c, err := initConfig([]string{"-service", "github"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// service should be overridden by CLI flag
	if c.service != "github" {
		t.Errorf("Expected service 'github' from CLI flag, got: %v", c.service)
	}
	// ignore_fork should still be true from config file (not overridden)
	if !c.ignoreFork {
		t.Error("Expected ignore_fork to be true from config file")
	}
}

func TestInitConfigNoConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// No config file — should behave exactly as before
	c, err := initConfig([]string{"-service", "github"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if c.service != "github" {
		t.Errorf("Expected service 'github', got: %v", c.service)
	}
	// Defaults should be the flag defaults
	if c.ignoreFork {
		t.Error("Expected ignore_fork to be false by default")
	}
	if c.githubRepoType != "all" {
		t.Errorf("Expected github.repoType 'all', got: %v", c.githubRepoType)
	}
}

func TestHandleValidateConfigInvalidRepoType(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.WriteFile(defaultConfigFile, []byte("service: github\ngithub:\n  repo_type: badvalue\n"), 0644)
	os.Setenv("GITHUB_TOKEN", "testtoken")
	defer os.Unsetenv("GITHUB_TOKEN")

	err := handleValidateConfig()
	if err == nil {
		t.Fatal("Expected validation error for invalid repo_type")
	}
}
