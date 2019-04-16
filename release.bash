#!/usr/bin/env bash
# Stolen from https://github.com/oklog/oklog/blob/master/release.fish
set -eu

VERSION=$1
git tag --annotate $VERSION -m "Release $VERSION"
git push origin $VERSION