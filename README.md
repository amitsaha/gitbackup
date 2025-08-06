# gitbackup - Backup your GitHub, GitLab, and Bitbucket repositories
Code Quality [![Go Report Card](https://goreportcard.com/badge/github.com/amitsaha/gitbackup)](https://goreportcard.com/report/github.com/amitsaha/gitbackup)
[![.github/workflows/ci.yml](https://github.com/amitsaha/gitbackup/actions/workflows/ci.yml/badge.svg)](https://github.com/amitsaha/gitbackup/actions/workflows/ci.yml)

- [gitbackup - Backup your GitHub, GitLab, and Bitbucket repositories](#gitbackup---backup-your-github-gitlab-and-bitbucket-repositories)
  - [Introduction](#introduction)
  - [Installing `gitbackup`](#installing-gitbackup)
  - [Using `gitbackup`](#using-gitbackup)
    - [GitHub Specific oAuth App Flow](#github-specific-oauth-app-flow)
    - [OAuth Scopes/Permissions required](#oauth-scopespermissions-required)
      - [Bitbucket](#bitbucket)
      - [GitHub](#github)
      - [GitLab](#gitlab)
    - [Security and credentials](#security-and-credentials)
    - [Examples](#examples)
      - [Backing up your GitHub repositories](#backing-up-your-github-repositories)
      - [Backing up your GitLab repositories](#backing-up-your-gitlab-repositories)
      - [GitHub Enterprise or custom GitLab installation](#github-enterprise-or-custom-gitlab-installation)
      - [Backing up your Bitbucket repositories](#backing-up-your-bitbucket-repositories)
      - [Specifying a backup location](#specifying-a-backup-location)
      - [Cloning bare repositories](#cloning-bare-repositories)
      - [GitHub Migrations](#github-migrations)
  - [Building](#building)
  
## Introduction

``gitbackup`` is a tool to backup your git repositories from GitHub (including GitHub enterprise),
GitLab (including custom GitLab installations), or Bitbucket.

``gitbackup`` currently has two operation modes:

- The first and original operating mode is to create clones of only your git repository. This is supported for Bitbucket, GitHub and Gitlab.
- The second operating mode is only available for GitHub where you can create a user migration (including orgs) which you get back as a .tar.gz
  file containing all the artefacts that GitHub supports via their Migration API.
  
If you are following along my [Linux Journal article](https://www.linuxjournal.com/content/back-github-and-gitlab-repositories-using-golang) (published in 2017), please obtain the version of the 
source tagged with [lj-0.1](https://github.com/amitsaha/gitbackup/releases/tag/lj-0.1).

## Installing `gitbackup`

Binary releases are available from the [Releases](https://github.com/amitsaha/gitbackup/releases/) page. Please download the binary corresponding to your OS
and architecture and copy the binary somewhere in your ``$PATH``. It is recommended to rename the binary to `gitbackup` or `gitbackup.exe` (on Windows).

If you are on MacOS, a community member has created a [Homebrew formula](https://formulae.brew.sh/formula/gitbackup).

## Using `gitbackup`

``gitbackup`` requires a [GitHub API access token](https://github.com/blog/1509-personal-api-tokens) for
backing up GitHub repositories, a [GitLab personal access token](https://gitlab.com/profile/personal_access_tokens)
for GitLab repositories, and a username and [app password](https://bitbucket.org/account/settings/app-passwords/) for
Bitbucket repositories.

You can supply the tokens to ``gitbackup`` using ``GITHUB_TOKEN`` and ``GITLAB_TOKEN`` environment variables
respectively, and the Bitbucket credentials with ``BITBUCKET_USERNAME`` and ``BITBUCKET_PASSWORD``.

### GitHub Specific oAuth App Flow

Starting with the 0.6 release, if you run `gitbackup` without specifying `GITHUB_TOKEN`, it will prompt you to complete
a oAuth flow to grant the necessary access:

```
$ ./gitbackup -service github -github.repoType starred
Copy code: <some code>
then open: https://github.com/login/device
```
Once your authorize the app, `gitbackup` will retrieve the token, and also store it in your operating system's
keychain/keyring (using the [99designs/keyring](https://github.com/99designs/keyring) package - thanks!). Next
time you run it, it will ask you for the keyring password and retrieve the token automatically.


### OAuth Scopes/Permissions required

#### Bitbucket

For the App password, the following permissions are required:

- `Account:Read`
- `Repositories:Read`

#### GitHub

- `repo`: Reading repositories, including private repositories
- `user` and `admin:org`: Basically, this gives `gitbackup` a lot of permissions than you may be comfortable with. 
   However, these are required for the user migration and org migration operations.

#### GitLab

- `api`: Grants complete read/write access to the API, including all groups and projects.
For some reason, `read_user` and `read_repository` is not sufficient.

### Security and credentials

When you provide the tokens via environment variables, they remain accessible in your shell history 
and via the processes' environment for the lifetime of the process. By default, SSH authentication
is used to clone your repositories. If `use-https-clone` is specified, private repositories
are cloned via `https` basic auth and the token provided will be stored  in the repositories' 
`.git/config`.

### Examples

Typing ``-help`` will display the command line options that `gitbackup` recognizes:

```
$ gitbackup -help
Usage of ./gitbackup:
  -backupdir string
        Backup directory
  -bare
        Clone bare repositories
  -githost.url string
        DNS of the custom Git host
  -github.createUserMigration
        Download user data
  -github.createUserMigrationRetry
        Retry creating the GitHub user migration if we get an error (default true)
  -github.createUserMigrationRetryMax int
        Number of retries to attempt for creating GitHub user migration (default 5)
  -github.listUserMigrations
        List available user migrations
  -github.namespaceWhitelist string
        Organizations/Users from where we should clone (separate each value by a comma: 'user1,org2')
  -github.repoType string
        Repo types to backup (all, owner, member, starred) (default "all")
  -github.waitForUserMigration
        Wait for migration to complete (default true)
  -gitlab.projectMembershipType string
        Project type to clone (all, owner, member, starred) (default "all")
  -gitlab.projectVisibility string
        Visibility level of Projects to clone (internal, public, private) (default "internal")
  -ignore-fork
        Ignore repositories which are forks
  -ignore-private
        Ignore private repositories/projects
  -service string
        Git Hosted Service Name (github/gitlab/bitbucket)
  -use-https-clone
        Use HTTPS for cloning instead of SSH
```

#### Backing up your GitHub repositories

To backup all your own GitHub repositories to the default backup directory (``$HOME/.gitbackup/``):

```lang=bash
$ GITHUB_TOKEN=secret$token gitbackup -service github
```

To backup only the GitHub repositories which you are the "owner" of:

```lang=bash
$ GITHUB_TOKEN=secret$token gitbackup -service github -github.repoType owner
```

To backup only the GitHub repositories which you are the "member" of:

```lang=bash
$ GITHUB_TOKEN=secret$token gitbackup -service github -github.repoType member
```

Separately, to backup GitHub repositories you have starred:

```lang=bash
$ GITHUB_TOKEN=secret$token gitbackup -service github -github.repoType starred
```

Additionally, to backup only the GitHub repositories under 'user1' and 'org3':

```lang=bash
$ GITHUB_TOKEN=secret$token gitbackup -service github -github.namespaceWhitelist "user1,org3"
```

#### Backing up your GitLab repositories

To backup all projects you either own or are a member of which have their [visibility](https://docs.gitlab.com/ce/api/projects.html#project-visibility-level) set to
"internal" on ``https://gitlab.com`` to the default backup directory (``$HOME/.gitbackup/``):

```lang=bash
$ GITLAB_TOKEN=secret$token gitbackup -service gitlab
```

To backup only the GitLab projects (either you are an owner or member of) which are "public"

```lang=bash
$ GITLAB_TOKEN=secret$token gitbackup -service gitlab -gitlab.projectVisibility public
```

To backup only the private repositories (either you are an owner or member of):

```lang=bash
$ GITLAB_TOKEN=secret$token gitbackup -service gitlab -gitlab.projectVisibility private
```

To backup public repositories which you are an owner of:

```lang=bash
$ GITLAB_TOKEN=secret$token gitbackup \
    -service gitlab \
    -gitlab.projectVisibility public \
    -gitlab.projectMembershipType owner
```

To backup public repositories which you are an member of:

```lang=bash
$ GITLAB_TOKEN=secret$token gitbackup \
    -service gitlab \
    -gitlab.projectVisibility public \
    -gitlab.projectMembershipType member
```

To backup GitLub repositories you have starred:

```lang=bash
$ GITLAB_TOKEN=secret$token gitbackup -service gitlab \
  -gitlab.projectMembershipType starred \
  -gitlab.projectVisibility public
```

#### GitHub Enterprise or custom GitLab installation

To specify a custom GitHub enterprise or GitLab location, specify the ``service`` as well as the
the ``githost.url`` flag, like so

```lang=bash
$ GITLAB_TOKEN=secret$token gitbackup -service gitlab -githost.url https://git.yourhost.com
```

#### Backing up your Bitbucket repositories

To backup all your Bitbucket repositories to the default backup directory (``$HOME/.gitbackup/``):

```lang=bash
$ BITBUCKET_USERNAME=username BITBUCKET_PASSWORD=password gitbackup -service bitbucket
```

#### Specifying a backup location

To specify a custom backup directory, we can use the ``backupdir`` flag:

```lang=bash
$ GITHUB_TOKEN=secret$token gitbackup -service github -backupdir /data/
```

This will create a ``github.com`` directory in ``/data`` and backup all your repositories there instead.
Similarly, it will create a ``gitlab.com`` directory, if you are backing up repositories from ``gitlab``, and a
``bitbucket.com`` directory if you are backing up from Bitbucket.
If you have specified a Git Host URL, it will create a directory structure ``data/host-url/``.


#### Cloning bare repositories

To clone bare repositories, we can use the ``bare`` flag:

```lang=bash
$ GITHUB_TOKEN=secret$token gitbackup -service github -bare
```

This will create a directory structure like ``github.com/org/repo.git`` containing bare repositories.

#### GitHub Migrations

`gitbackup` starting from the 0.6 release includes support for downloading your user data/organization data as 
made available via the [Migrations API](https://docs.github.com/en/rest/reference/migrations). As of this
release, you can create an user migration (including your owned organizations data) and download the migration
artefact using the following command:

```
$ ./gitbackup -service github -github.createUserMigration -ignore-fork -github.repoType owner

2021/05/14 05:05:27 /home/runner/.gitbackup/github.com doesn't exist, creating it
2021/05/14 05:05:35 Creating a user migration for 129 repos
2021/05/14 05:05:46 Waiting for migration state to be exported: 0xc0002a6260
2021/05/14 05:06:48 Waiting for migration state to be exported: 0xc000290070
..
2021/05/14 05:33:44 Waiting for migration state to be exported: 0xc0001c2020

2021/05/14 05:34:46 Downloading file to: /home/runner/.gitbackup/github.com/user-migration-571089.tar.gz

2021/05/14 05:35:00 Creating a org migration (FedoraScientific) for 19 repos
2021/05/14 05:35:03 Waiting for migration state to be exported: 0xc000144050
..
2021/05/14 05:39:05 Downloading file to: /home/runner/.gitbackup/github.com/FedoraScientific-migration-571098.tar.gz
..
2021/05/14 05:46:16 Downloading file to: /home/runner/.gitbackup/github.com/practicalgo-migration-571103.tar.gz
```
You can then integrate this with your own scripting to push the data to S3 for example (See an example
workflow via scheduled github actions [here](https://github.com/amitsaha/gitbackup/actions/workflows/backup.yml)).

## Building

If you have Go 1.24.x installed, you can clone the repository and:

```
$ go build
```

The built binary will be ``gitbackup``.
