#!/usr/bin/env bash
# Stolen from https://github.com/oklog/oklog/blob/master/release.fish

VERSION=$1
DISTDIR="dist/$VERSION"
mkdir -p $DISTDIR
git tag --annotate v$VERSION -m "Release v$VERSION"

for pair in linux/386 linux/amd64 linux/arm linux/arm64 darwin/amd64 dragonfly/amd64 freebsd/amd64 netbsd/amd64 openbsd/amd64 windows/amd64; do
	GOOS=`echo $pair | cut -d'/' -f1`
    GOARCH=`echo $pair | cut -d'/' -f2` 
    GOOS=$GOOS GOARCH=$GOARCH GOPATH=`pwd`/vendor/  go build -o $DISTDIR/gitbackup-$VERSION-$GOOS-$GOARCH src/gitbackup/main.go src/gitbackup/client.go src/gitbackup/backup.go src/gitbackup/repositories.go
done
