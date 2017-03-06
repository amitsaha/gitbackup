[![Build Status](https://travis-ci.org/amitsaha/gitbackup.svg?branch=master)](https://travis-ci.org/amitsaha/gitbackup)

# gitbackup - Backup your GitHub and GitLab repositories

``gitbackup`` is a tool to backup your git repositories from GitHub or GitLab (including custom GitLab installations).

## Building

Setup Golang 1.8 and [gb](https://getgb.io) following my blog post [here](http://echorand.me/setup-golang-18-and-gb-on-fedora-and-other-linux-distributions.html) and then:
```
$ gb build 
```

## Using

```
./bin/gitbackup -help
Usage of ./bin/gitbackup:
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

If you are following along my Linux Journal article, please obtain the version of the source tagged with [lj-0.1](https://github.com/amitsaha/gitbackup/releases/tag/lj-0.1).


