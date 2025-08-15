package main

import (
	"os"
	"gopkg.in/yaml.v3"
)

func LoadYAMLConfig(path string) (*appConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var c appConfig
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

func WriteSampleYAMLConfig(path string) error {
	c := appConfig{
		service: "github",
		gitHostURL: "github.com",
		backupDir: "./backup",
		ignorePrivate: false,
		ignoreFork: false,
		useHTTPSClone: true,
		bare: false,
		githubRepoType: "all",
		githubNamespaceWhitelist: []string{"user1", "org2"},
		githubCreateUserMigration: false,
		githubCreateUserMigrationRetry: true,
		githubCreateUserMigrationRetryMax: 5,
		githubListUserMigrations: false,
		githubWaitForMigrationComplete: true,
		gitlabProjectVisibility: "internal",
		gitlabProjectMembershipType: "all",
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := yaml.NewEncoder(f)
	defer enc.Close()
	return enc.Encode(c)
}
