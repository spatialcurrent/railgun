#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LDFLAGS="-X main.gitBranch=$(git branch | grep \* | cut -d ' ' -f2) -X main.gitCommit=$(git rev-list -1 HEAD)"
go run -ldflags "$LDFLAGS" $DIR/../cmd/railgun/main.go version