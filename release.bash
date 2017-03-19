#!/usr/bin/env bash
# Stolen from https://github.com/oklog/oklog/blob/master/release.fish

VERSION=$1
DISTDIR="dist/$VERSION"
mkdir -p $DISTDIR
git tag --annotate v$VERSION -m "Release v$VERSION"

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
