#!/usr/bin/env bash

set -e

export GO111MODULE=on
go env
golint -set_exit_status
go build
go test