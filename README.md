# gitbackup - Backup your GitHub and GitLab repositories

[![Build Status](https://travis-ci.org/amitsaha/gitbackup.svg?branch=master)](https://travis-ci.org/amitsaha/gitbackup)


``gitbackup`` is a tool to backup your git repositories from GitHub (including GitHub enterprise) or 
GitLab (including custom GitLab installations). 

``gitbackup`` only creates a backup of the repository and does not currently support issues, 
pull requests or other data associated with a git repository.

If you are following along my Linux Journal article, please obtain the version of the source tagged 
with [lj-0.1](https://github.com/amitsaha/gitbackup/releases/tag/lj-0.1).

## Using ``gitbackup``

``gitbackup`` requires a [GitHub API access token](https://github.com/blog/1509-personal-api-tokens) for backing up GitHub repositories and [GitLab personal access token](https://gitlab.com/profile/personal_access_tokens) for GitLab. You can supply the token to ``gitbackup`` using ``GITHUB_TOKEN`` and ``GITLAB_TOKEN`` respectively.

Typing ``-help`` will display the command line options that ``gitbackup`` recognizes:

```
./bin/gitbackup -help
Usage of ./bin/gitbackup:
  -backupdir string
    	Backup directory
  -githost.url string
    	DNS of the custom Git host
  -github.repoType string
    	Repo types to backup (all, owner, member) (default "all")
  -gitlab.projectVisibility string
    	Visibility level of Projects to clone (default "internal")
  -service string
    	Git Hosted Service Name (github/gitlab)
```
### Backing up your GitHub repositories

To backup all your GitHub repositories to the default backup directory (``$HOME/.gitbackup/github``):

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

### Backing up your GitLab repositories

To backup all projects which have their [visibility](https://docs.gitlab.com/ce/api/projects.html#project-visibility-level) set to "internal" on ``https://gitlab.com`` to the default backup directory (``$HOME/.gitbackup/gitlab``):

```lang=bash
$ GITLAB_TOKEN=secret$token gitbackup -service gitlab
```

To backup only the GitLab projects which are "public"

```lang=bash
$ GITLAB_TOKEN=secret$token gitbackup -service gitlab -gitlab.projectVisibility public
```

To backup only the private repositories:

```lang=bash
$ GITLAB_TOKEN=secret$token gitbackup -service gitlab -gitlab.projectVisibility private
```


### GitHub Enterprise or custom GitLab installation

To specify a custom GitHub enterprise or GitLab location, use the ``githost.url`` flag, like so:

```lang=bash
$ GITLAB_TOKEN=secret$token gitbackup -service gitlab -githost.url https://git.yourhost.com -gitlab.projectVisibility private
```


### Specifying a backup location

To specify a custom backup directory, we can use the ``backupdir`` flag:

```lang=bash
$ GITHUB_TOKEN=secret$token gitbackup -service github -backupdir /data/
```

This will create a ``github`` directory in ``/data`` and backup all your repositories there instead.
Similarly, it will create a ``gitlab`` directory, if you are backing up repositories from ``gitlab``.


## Building


Setup Golang 1.8 and [gb](https://getgb.io) following my blog post [here](http://echorand.me/setup-golang-18-and-gb-on-fedora-and-other-linux-distributions.html) and then:
```
$ gb build 
```

The built binary will be in ``bin/gitbackup``.
