package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

// buildTestConfig creates a minimal cli.App with appFlags(), runs it with the
// given args, and returns the appConfig produced by buildConfig.
func buildTestConfig(args []string) (*appConfig, error) {
	var result *appConfig
	var buildErr error

	app := &cli.App{
		Name:  "gitbackup",
		Flags: appFlags(),
		Action: func(cCtx *cli.Context) error {
			result, buildErr = buildConfig(cCtx)
			return buildErr
		},
	}

	// urfave/cli expects the program name as args[0]
	fullArgs := append([]string{"gitbackup"}, args...)
	if err := app.Run(fullArgs); err != nil {
		return nil, fmt.Errorf("app.Run: %w", err)
	}
	return result, buildErr
}

func TestHandleInitConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, defaultConfigFile)

	// First call should create the file
	err := handleInitConfig(configPath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

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
	err = handleInitConfig(configPath)
	if err == nil {
		t.Fatal("Expected error when config file already exists")
	}
}

func TestHandleInitConfigCreatesParentDir(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "nested", defaultConfigFile)

	err := handleInitConfig(configPath)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Expected config file to be created in nested directory")
	}
}

func TestHandleValidateConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, defaultConfigFile)

	// Validate should fail if no config file exists
	err := handleValidateConfig(configPath)
	if err == nil {
		t.Fatal("Expected error when config file doesn't exist")
	}

	// Create a valid config and set the required env var
	err = handleInitConfig(configPath)
	if err != nil {
		t.Fatalf("Expected no error creating config, got: %v", err)
	}

	os.Setenv("GITHUB_TOKEN", "testtoken")
	defer os.Unsetenv("GITHUB_TOKEN")

	err = handleValidateConfig(configPath)
	if err != nil {
		t.Fatalf("Expected valid config, got: %v", err)
	}
}

func TestHandleValidateConfigInvalidService(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, defaultConfigFile)

	// Write a config with an invalid service
	os.WriteFile(configPath, []byte("service: notaservice\n"), 0644)

	err := handleValidateConfig(configPath)
	if err == nil {
		t.Fatal("Expected validation error for invalid service")
	}
}

func TestHandleValidateConfigMissingEnvVar(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, defaultConfigFile)

	// Write a valid gitlab config but don't set GITLAB_TOKEN
	os.WriteFile(configPath, []byte("service: gitlab\ngitlab:\n  project_visibility: internal\n  project_membership_type: all\n"), 0644)
	os.Unsetenv("GITLAB_TOKEN")

	err := handleValidateConfig(configPath)
	if err == nil {
		t.Fatal("Expected validation error for missing GITLAB_TOKEN")
	}
}

func TestInitConfigWithConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, defaultConfigFile)

	// Write a config file with specific values
	os.WriteFile(configPath, []byte("service: gitlab\nignore_fork: true\nuse_https_clone: true\ngitlab:\n  project_visibility: private\n  project_membership_type: owner\n"), 0644)
	os.Setenv("GITLAB_TOKEN", "testtoken")
	defer os.Unsetenv("GITLAB_TOKEN")

	// buildTestConfig with --config flag should use config file values
	c, err := buildTestConfig([]string{"-config", configPath})
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
	configPath := filepath.Join(tmpDir, defaultConfigFile)

	// Config file says gitlab with ignore_fork true
	os.WriteFile(configPath, []byte("service: gitlab\nignore_fork: true\ngitlab:\n  project_visibility: private\n  project_membership_type: all\n"), 0644)
	os.Setenv("GITLAB_TOKEN", "testtoken")
	defer os.Unsetenv("GITLAB_TOKEN")

	// CLI flag overrides service to github
	c, err := buildTestConfig([]string{"-config", configPath, "-service", "github"})
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
	// No config file — should behave exactly as before
	c, err := buildTestConfig([]string{"-service", "github"})
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
	configPath := filepath.Join(tmpDir, defaultConfigFile)

	os.WriteFile(configPath, []byte("service: github\ngithub:\n  repo_type: badvalue\n"), 0644)
	os.Setenv("GITHUB_TOKEN", "testtoken")
	defer os.Unsetenv("GITHUB_TOKEN")

	err := handleValidateConfig(configPath)
	if err == nil {
		t.Fatal("Expected validation error for invalid repo_type")
	}
}
