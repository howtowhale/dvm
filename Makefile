SHELL := /bin/bash

COMMIT = $(shell git rev-parse --verify --short HEAD)
VERSION = $(shell git describe --tags --dirty='-dev' 2> /dev/null)
BREW_VERSION = $(shell git describe --tags --abbrev=0 2> /dev/null)
PERMALINK = $(shell if [[ $(VERSION) =~ [^-]*-([^.]+).* ]]; then echo $${BASH_REMATCH[1]}; else echo "latest"; fi)

GITHUB_ORG = howtowhale
GITHUB_REPO = dvm
PACKAGE = github.com/${GITHUB_ORG}/${GITHUB_REPO}/dvm-helper
UPGRADE_DISABLED = false

LDFLAGS = -w -X main.dvmCommit=${COMMIT} -X main.dvmVersion=${VERSION} -X main.upgradeDisabled=${UPGRADE_DISABLED}

GOCMD = go
GOBUILD = $(GOCMD) build -a -tags netgo -ldflags '$(LDFLAGS)'

BINDIR = bin/dvm/$(VERSION)
GOFILES = dvm-helper/*.go
GOFILES_NOVENDOR = $(shell go list ./... | grep -v /vendor/)

default: local

homebrew:
	brew bump-formula-pr --strict --url=https://github.com/howtowhale/dvm/archive/$(BREW_VERSION).tar.gz dvm

validate:
	go fmt $(GOFILES_NOVENDOR)
	go vet $(GOFILES_NOVENDOR)
	-go list ./... | grep -v /vendor/ | xargs -L1 golint --set_exit_status

test: local
	go test -v $(GOFILES_NOVENDOR)
	eval "$( ./dvm-helper --bash-completion )"
	./dvm-helper/dvm-helper --version

cross-build: local linux linux32 darwin windows windows32
	cp dvm.sh dvm.ps1 dvm.cmd install.sh install.ps1 README.md LICENSE bash_completion $(BINDIR)/
	find $(BINDIR) -maxdepth 1 -name "install.*" -exec sed -i -e 's/$(PERMALINK)/$(VERSION)/g' {} \;
	cp -R $(BINDIR) bin/dvm/$(PERMALINK)

local: $(GOFILES)
	CGO_ENABLED=0 $(GOBUILD) -o dvm-helper/dvm-helper $(PACKAGE)

linux: $(GOFILES)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINDIR)/Linux/x86_64/dvm-helper $(PACKAGE)
	cd $(BINDIR)/Linux/x86_64 && shasum -a 256 dvm-helper > dvm-helper.sha256

linux32: $(GOFILES)
	CGO_ENABLED=0 GOOS=linux GOARCH=386 $(GOBUILD) -o $(BINDIR)/Linux/i686/dvm-helper $(PACKAGE)
	cd $(BINDIR)/Linux/i686 && shasum -a 256 dvm-helper > dvm-helper.sha256

darwin: $(GOFILES)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINDIR)/Darwin/x86_64/dvm-helper $(PACKAGE)
	cd $(BINDIR)/Darwin/x86_64 && shasum -a 256 dvm-helper > dvm-helper.sha256

windows: $(GOFILES)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINDIR)/Windows/x86_64/dvm-helper.exe $(PACKAGE)
	cd $(BINDIR)/Windows/x86_64 && shasum -a 256 dvm-helper.exe > dvm-helper.exe.sha256

windows32: $(GOFILES)
	CGO_ENABLED=0 GOOS=windows GOARCH=386 $(GOBUILD) -o $(BINDIR)/Windows/i686/dvm-helper.exe $(PACKAGE)
	cd $(BINDIR)/Windows/i686 && shasum -a 256 dvm-helper.exe > dvm-helper.exe.sha256

# To make a release, push a tag to master, e.g. git tag 0.2.0 -a -m ""

.PHONY: clean deploy

clean:
	 -rm -fr bin/*
	 -rm -fr _deploy/
	 -rm dvm-helper/dvm-helper
	 -rm dvm-helper/dvm-helper.exe

deploy:
	VERSION=$(VERSION) PERMALINK=$(PERMALINK) ./script/deploy-gh-pages.sh
