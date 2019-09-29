# =================================================================
#
# Copyright (C) 2019 Spatial Current, Inc. - All Rights Reserved
# Released as open source under the MIT License.  See LICENSE file.
#
# =================================================================

ifdef GOPATH
GCFLAGS=-trimpath=$(shell printenv GOPATH)/src
else
GCFLAGS=-trimpath=$(shell go env GOPATH)/src
endif

LDFLAGS=-X main.gitBranch=$(shell git branch | grep \* | cut -d ' ' -f2) -X main.gitCommit=$(shell git rev-list -1 HEAD)

ifndef DEST
DEST=bin
endif

.PHONY: help
help:  ## Print the help documentation
	@grep -E '^[a-zA-Z_-\]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

#
# Dependencies
#

deps_go:  ## Install Go dependencies
	go get -d -t ./...

.PHONY: deps_go_test
deps_go_test: ## Download Go dependencies for tests
	go get golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow # download shadow
	go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow # install shadow
	go get -u github.com/kisielk/errcheck # download and install errcheck
	go get -u github.com/client9/misspell/cmd/misspell # download and install misspell
	go get -u github.com/gordonklaus/ineffassign # download and install ineffassign
	go get -u honnef.co/go/tools/cmd/staticcheck # download and instal staticcheck
	go get -u golang.org/x/tools/cmd/goimports # download and install goimports

deps_arm:  ## Install dependencies to cross-compile to ARM
	# ARMv7
	apt-get install -y libc6-armel-cross libc6-dev-armel-cross binutils-arm-linux-gnueabi libncurses5-dev gcc-arm-linux-gnueabi g++-arm-linux-gnueabi
  # ARMv8
	apt-get install gcc-aarch64-linux-gnu g++-aarch64-linux-gnu

deps_gopherjs:  ## Install GopherJS with jsbuiltin
	go get -u github.com/gopherjs/gopherjs
	go get -u -d -tags=js github.com/gopherjs/jsbuiltin

deps_javascript:  ## Install dependencies for JavaScript tests
	npm install .

bin/chamber_linux_amd64:
	GOOS=linux go build -o bin/chamber_linux_amd64 -ldflags "-linkmode external -extldflags -static" github.com/segmentio/chamber

#
# Go building, formatting, testing, and installing
#

.PHONY: fmt
fmt:  ## Format Go source code
	go fmt $$(go list ./... )

.PHONY: imports
imports: ## Update imports in Go source code
	goimports -w -local github.com/spatialcurrent/railgun,github.com/spatialcurrent/ $$(find . -iname '*.go')

.PHONY: vet
vet: ## Vet Go source code
	go vet $$(go list ./... )

.PHONY: test_pkg
test_pkg: ## Run Go tests
	CGO_ENABLED=0 go test -count 1 github.com/spatialcurrent/railgun/pkg/...

.PHONY: test_cmd
test_cmd: ## Run Go tests
	CGO_ENABLED=0 go test -count 1 github.com/spatialcurrent/railgun/cmd/...

.PHONY: test_cli
test_cli: ## Run Go tests
	bash scripts/test-cli.sh

.PHONY: test_go
test_go: ## Run Go tests
	CGO_ENABLED=0 bash scripts/test.sh

build: build_cli build_javascript  ## Build CLI and JavaScript

install:  ## Install railgun CLI on current platform
	go install -gcflags="$(GCFLAGS)" -ldflags="$(LDFLAGS)" github.com/spatialcurrent/railgun/cmd/railgun

#
# Command line Programs
#

bin/railgun:
	go build -o bin/railgun -gcflags="$(GCFLAGS)" -ldflags="$(LDFLAGS)" github.com/spatialcurrent/railgun/cmd/railgun

bin/railgun_darwin_amd64:
	GOOS=darwin GOARCH=amd64 go build -o bin/railgun_darwin_amd64 -gcflags="$(GCFLAGS)" -ldflags="$(LDFLAGS)" github.com/spatialcurrent/railgun/cmd/railgun

bin/railgun_linux_amd64:
	# GOTAGS= CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build --tags "linux" -o "railgun_linux_amd64" -ldflags "$LDFLAGS" github.com/spatialcurrent/railgun/cmd/railgun
	GOOS=linux GOARCH=amd64 go build -o bin/railgun_linux_amd64 -gcflags="$(GCFLAGS)" -ldflags="$(LDFLAGS)" github.com/spatialcurrent/railgun/cmd/railgun

bin/railgun_windows_amd64.exe:
	# GOTAGS= CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc go build -o "railgun_windows_amd64.exe" -ldflags "$LDFLAGS" github.com/spatialcurrent/railgun/cmd/railgun
	GOOS=windows GOARCH=amd64 go build -o bin/railgun_windows_amd64.exe -gcflags="$(GCFLAGS)" -ldflags="$(LDFLAGS)" github.com/spatialcurrent/railgun/cmd/railgun

bin/railgun_linux_arm64:
	GOOS=linux GOARCH=arm64 go build -o bin/railgun_linux_arm64 -gcflags="$(GCFLAGS)" -ldflags="$(LDFLAGS)" github.com/spatialcurrent/railgun/cmd/railgun

build_cli: bin/railgun_darwin_amd64 bin/railgun_linux_amd64 bin/railgun_windows_amd64.exe bin/railgun_linux_arm64  ## Build command line programs

build_cli_container:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o bin/railgun_linux_amd64 -gcflags="$(GCFLAGS)" -ldflags="$(LDFLAGS)" github.com/spatialcurrent/railgun/cmd/railgun

#
# JavaScript
#

dist/railgun.mod.js:  ## Build JavaScript module
	gopherjs build -o dist/railgun.mod.js github.com/spatialcurrent/gotmpl/cmd/railgun.mod.js

dist/railgun.mod.min.js:  ## Build minified JavaScript module
	gopherjs build -m -o dist/railgun.mod.min.js github.com/spatialcurrent/gotmpl/cmd/railgun.mod.js

dist/railgun.global.js:  ## Build JavaScript library that attaches to global or window.
	gopherjs build -o dist/railgun.global.js github.com/spatialcurrent/gotmpl/cmd/railgun.global.js

dist/railgun.global.min.js:  ## Build minified JavaScript library that attaches to global or window.
	gopherjs build -m -o dist/railgun.global.min.js github.com/spatialcurrent/railgun/cmd/railgun.global.js

build_javascript: dist/railgun.mod.js dist/railgun.mod.min.js dist/railgun.global.js dist/railgun.global.min.js  ## Build artifacts for JavaScript

test_javascript:  ## Run JavaScript tests
	npm run test

lint:  ## Lint JavaScript source code
	npm run lint

#
# Local Development
#

secret/keypair:
	@mkdir -p secret
	@ssh-keygen -t rsa -b 4096 -m PEM -f secret/keypair
	@openssl rsa -in secret/keypair -pubout -outform PEM -out secret/keypair.pub
	@echo "Keypair created at secret/keypair"

serve: secret/keypair  ## Start Railgun server
	@bash scripts/serve.sh

#
# Clean
#

clean:
	rm -fr bin
	rm -fr dist
	rm -fr secret/keypair
	rm -fr secret/keypair.pub
