package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const defaultConfigFile = "gitbackup.yml"

// fileConfig represents the YAML configuration file structure.
// Migration-related flags are intentionally excluded as they
// are one-off operations better suited to CLI flags.
type fileConfig struct {
	Service       string       `yaml:"service"`
	GitHostURL    string       `yaml:"githost_url"`
	BackupDir     string       `yaml:"backup_dir"`
	IgnorePrivate bool         `yaml:"ignore_private"`
	IgnoreFork    bool         `yaml:"ignore_fork"`
	UseHTTPSClone bool         `yaml:"use_https_clone"`
	Bare          bool         `yaml:"bare"`
	GitHub        githubConfig `yaml:"github"`
	GitLab        gitlabConfig `yaml:"gitlab"`
	Forgejo       forgejoConfig `yaml:"forgejo"`
}

type githubConfig struct {
	RepoType           string   `yaml:"repo_type"`
	NamespaceWhitelist []string `yaml:"namespace_whitelist"`
}

type gitlabConfig struct {
	ProjectVisibility     string `yaml:"project_visibility"`
	ProjectMembershipType string `yaml:"project_membership_type"`
}

type forgejoConfig struct {
	RepoType string `yaml:"repo_type"`
}

// defaultFileConfig returns a fileConfig with the same defaults as the CLI flags
func defaultFileConfig() fileConfig {
	return fileConfig{
		Service:       "github",
		GitHostURL:    "",
		BackupDir:     "",
		IgnorePrivate: false,
		IgnoreFork:    false,
		UseHTTPSClone: false,
		Bare:          false,
		GitHub: githubConfig{
			RepoType:           "all",
			NamespaceWhitelist: []string{},
		},
		GitLab: gitlabConfig{
			ProjectVisibility:     "internal",
			ProjectMembershipType: "all",
		},
		Forgejo: forgejoConfig{
			RepoType: "user",
		},
	}
}

// handleInitConfig creates a default gitbackup.yml in the current directory
func handleInitConfig() error {
	if _, err := os.Stat(defaultConfigFile); err == nil {
		return fmt.Errorf("%s already exists", defaultConfigFile)
	}

	cfg := defaultFileConfig()
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("error generating config: %v", err)
	}

	err = os.WriteFile(defaultConfigFile, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing %s: %v", defaultConfigFile, err)
	}

	fmt.Printf("Created %s\n", defaultConfigFile)
	return nil
}

// fileConfigToAppConfig converts a fileConfig into an appConfig.
// Migration-related fields are left at their zero values since they
// are CLI-only flags.
func fileConfigToAppConfig(fc *fileConfig) *appConfig {
	return &appConfig{
		service:                     fc.Service,
		gitHostURL:                  fc.GitHostURL,
		backupDir:                   fc.BackupDir,
		ignorePrivate:               fc.IgnorePrivate,
		ignoreFork:                  fc.IgnoreFork,
		useHTTPSClone:               fc.UseHTTPSClone,
		bare:                        fc.Bare,
		githubRepoType:              fc.GitHub.RepoType,
		githubNamespaceWhitelist:    fc.GitHub.NamespaceWhitelist,
		gitlabProjectVisibility:     fc.GitLab.ProjectVisibility,
		gitlabProjectMembershipType: fc.GitLab.ProjectMembershipType,
		forgejoRepoType:             fc.Forgejo.RepoType,
	}
}

// loadConfigFile reads and parses gitbackup.yml from the current directory
func loadConfigFile() (*fileConfig, error) {
	data, err := os.ReadFile(defaultConfigFile)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %v", defaultConfigFile, err)
	}

	var cfg fileConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s: %v", defaultConfigFile, err)
	}
	return &cfg, nil
}

// handleValidateConfig reads gitbackup.yml and validates its contents
func handleValidateConfig() error {
	cfg, err := loadConfigFile()
	if err != nil {
		return err
	}

	var errors []string

	// Validate service
	if _, ok := knownServices[cfg.Service]; !ok {
		errors = append(errors, fmt.Sprintf("invalid service: %q (must be github, gitlab, bitbucket, or forgejo)", cfg.Service))
	}

	// Validate service-specific field values
	switch cfg.Service {
	case "github":
		if !contains([]string{"all", "owner", "member", "starred"}, cfg.GitHub.RepoType) {
			errors = append(errors, fmt.Sprintf("invalid github.repo_type: %q (must be all, owner, member, or starred)", cfg.GitHub.RepoType))
		}
	case "gitlab":
		if !contains([]string{"internal", "public", "private"}, cfg.GitLab.ProjectVisibility) {
			errors = append(errors, fmt.Sprintf("invalid gitlab.project_visibility: %q (must be internal, public, or private)", cfg.GitLab.ProjectVisibility))
		}
		if !validGitlabProjectMembership(cfg.GitLab.ProjectMembershipType) {
			errors = append(errors, fmt.Sprintf("invalid gitlab.project_membership_type: %q (must be all, owner, member, or starred)", cfg.GitLab.ProjectMembershipType))
		}
	case "forgejo":
		if !contains([]string{"user", "starred"}, cfg.Forgejo.RepoType) {
			errors = append(errors, fmt.Sprintf("invalid forgejo.repo_type: %q (must be user or starred)", cfg.Forgejo.RepoType))
		}
	}

	// Validate required environment variables
	switch cfg.Service {
	case "github":
		if os.Getenv("GITHUB_TOKEN") == "" {
			errors = append(errors, "GITHUB_TOKEN environment variable not set")
		}
	case "gitlab":
		if os.Getenv("GITLAB_TOKEN") == "" {
			errors = append(errors, "GITLAB_TOKEN environment variable not set")
		}
	case "bitbucket":
		if os.Getenv("BITBUCKET_USERNAME") == "" {
			errors = append(errors, "BITBUCKET_USERNAME environment variable not set")
		}
		if os.Getenv("BITBUCKET_TOKEN") == "" && os.Getenv("BITBUCKET_PASSWORD") == "" {
			errors = append(errors, "BITBUCKET_TOKEN or BITBUCKET_PASSWORD environment variable must be set")
		}
	case "forgejo":
		if os.Getenv("FORGEJO_TOKEN") == "" {
			errors = append(errors, "FORGEJO_TOKEN environment variable not set")
		}
	}

	if len(errors) > 0 {
		fmt.Println("Validation errors:")
		for _, e := range errors {
			fmt.Printf("  - %s\n", e)
		}
		return fmt.Errorf("config validation failed")
	}

	fmt.Printf("%s is valid\n", defaultConfigFile)
	return nil
}
