package main

// appConfig holds the application configuration
type appConfig struct {
	service       string
	gitHostURL    string
	backupDir     string
	ignorePrivate bool
	ignoreFork    bool
	useHTTPSClone bool
	bare          bool

	// GitHub specific configuration
	githubRepoType                    string
	githubNamespaceWhitelist          []string
	githubCreateUserMigration         bool
	githubCreateUserMigrationRetry    bool
	githubCreateUserMigrationRetryMax int
	githubListUserMigrations          bool
	githubWaitForMigrationComplete    bool

	// GitLab specific configuration
	gitlabProjectVisibility     string
	gitlabProjectMembershipType string

	// Forgejo specific configuration
	forgejoRepoType string
}
