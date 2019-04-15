#!/usr/bin/env bash

export GO111MODULE=on
go env
golint
go build
go test