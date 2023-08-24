package main

type appConfig struct {
	service       string
	gitHostURL    string
	backupDir     string
	ignorePrivate bool
	ignoreFork    bool
	useHTTPSClone bool
	bare          bool

	githubRepoType                    string
	githubNamespaceWhitelist          []string
	githubCreateUserMigration         bool
	githubCreateUserMigrationRetry    bool
	githubCreateUserMigrationRetryMax int
	githubListUserMigrations          bool
	githubWaitForMigrationComplete    bool

	gitlabProjectVisibility     string
	gitlabProjectMembershipType string
}
