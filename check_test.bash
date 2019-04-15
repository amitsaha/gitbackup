#!/usr/bin/env bash

set -e

export GO111MODULE=on
go env
golint
go build
go test