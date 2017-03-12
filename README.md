# gitbackup - Backup your GitHub and GitLab repositories

[![Build Status](https://travis-ci.org/amitsaha/gitbackup.svg?branch=master)](https://travis-ci.org/amitsaha/gitbackup)


``gitbackup`` is a tool to backup your git repositories from GitHub or GitLab (including custom GitLab installations). ``gitbackup`` only creates a backup of the repository and does not currently support issues, pull requests or other data associated with a git repository.

If you are following along my Linux Journal article, please obtain the version of the source tagged with [lj-0.1](https://github.com/amitsaha/gitbackup/releases/tag/lj-0.1).

## Using ``gitbackup``

``gitbackup`` requires a [GitHub API access token](https://github.com/blog/1509-personal-api-tokens) for backing up GitHub repositories and [GitLab personal access token](https://gitlab.com/profile/personal_access_tokens) for GitLab. You can supply the token to ``gitbackup`` using ``GITHUB_TOKEN`` and ``GITLAB_TOKEN`` respectively.

Typing ``-help`` will display the command line options that ``gitbackup`` recognizes:

```
$ gitbackup -help
Usage of gitbackup:
  -backupdir string
    	Backup directory
  -github.repoType string
    	Repo types to backup (all, owner, member) (default "all")
  -gitlab.projectVisibility string
    	Visibility level of Projects to clone (default "internal")
  -gitlab.url string
    	DNS of the GitLab service
  -service string
    	Git Hosted Service Name (github/gitlab)
```
### Backing up your GitHub repositories

To backup all your GitHub repositories to the default backup directory (``$HOME/.gitbackup/github``):

```lang=bash
$ GITHUB_TOKEN=secret$token gitbackup -service github
```


## Building



Setup Golang 1.8 and [gb](https://getgb.io) following my blog post [here](http://echorand.me/setup-golang-18-and-gb-on-fedora-and-other-linux-distributions.html) and then:
```
$ gb build 
```


