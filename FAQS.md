# Frequently Asked Questions (and Answers)

## How does the "backup" work?

`gitbackup` currently has two operation modes.

The first and original operating mode is to create clones of only your git repository. 
This is supported for Bitbucket, GitHub and Gitlab. It runs `git clone` (the first time) and then
subsequently `git pull`. You can also configure it to perform a bare clone instead. This means that
if you mess around your original repo, `gitbackup` will introduce the mess into your backup as well.

The second operating mode is only available for GitHub where you can create a user migration 
(including orgs) which you get back as a .tar.gz file containing all the artefacts that GitHub 
supports via their Migration API.

## Why do I need to provide credentials for public repositories?

This is to make the API call to get the list of repositories.
