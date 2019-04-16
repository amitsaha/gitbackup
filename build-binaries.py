import subprocess
import os

VERSION=subprocess.check_output(["git", "describe", "--abbrev=0", "--tags"]).decode("utf-8").rstrip("\n")
DISTDIR="./artifacts"

distpairs = [
    "linux/386",
    "linux/amd64",
    "linux/arm",
    "linux/arm64",
    "darwin/amd64",
    "dragonfly/amd64",
    "freebsd/amd64",
    "netbsd/amd64",
    "openbsd/amd64",
    "windows/amd64"
]

for distpair in distpairs:
    GOOS = distpair.split("/")[0]
    GOARCH = distpair.split("/")[1]
    OBJECT_FILE="gitbackup-{0}-{1}-{2}".format(VERSION, GOOS, GOARCH)
    subprocess.check_output(
        ["go", "build", "-o", "{0}/{1}".format(DISTDIR, OBJECT_FILE)],
        env = {"GOOS": GOOS, "GOARCH": GOARCH, "GOPATH": os.environ.get("GOPATH"), "GOCACHE": "on", "GOROOT": os.environ.get("GOROOT")})