#!/usr/bin/env bash

VERSION=$(git describe --abbrev=0 --tags)
DISTDIR="artifacts/"
export GO111MODULE=on

for pair in linux/386 linux/amd64 linux/arm linux/arm64 darwin/amd64 dragonfly/amd64 freebsd/amd64 netbsd/amd64 openbsd/amd64 windows/amd64; do
    GOOS=`echo $pair | cut -d'/' -f1`
    GOARCH=`echo $pair | cut -d'/' -f2` 
    OBJECT_FILE="gitbackup-$VERSION-$GOOS-$GOARCH"
    GOOS=$GOOS GOARCH=$GOARCH go build -o "$DISTDIR/$OBJECT_FILE" 
    pushd $DISTDIR
    echo $OBJECT_FILE
    zip "$OBJECT_FILE.zip" $OBJECT_FILE
    popd
done

ls $DISTDIR
