COMMIT = $(shell git rev-parse --verify --short HEAD)
VERSION = $(shell git describe --tags --dirty='-dev' 2> /dev/null)
GITHUB_ORG = getcarina
GITHUB_REPO = dvm
PACKAGE = github.com/${GITHUB_ORG}/${GITHUB_REPO}/dvm-helper
UPGRADE_DISABLED = false

LDFLAGS = -w -X main.dvmCommit=${COMMIT} -X main.dvmVersion=${VERSION} -X main.upgradeDisabled=${UPGRADE_DISABLED}

GOCMD = go
GOBUILD = $(GOCMD) build -a -tags netgo -ldflags '$(LDFLAGS)'

BINDIR = bin/dvm/$(VERSION)
GOFILES = dvm-helper/*.go

default: get-deps local

get-deps:
	go get github.com/Masterminds/glide
	glide install

#test: local
#	go test -v
#	eval "$( ./dvm-helper --bash-completion )"
#	./dvm-helper --version

cross-build: local linux linux32 darwin windows windows32
	cp dvm.sh dvm.ps1 dvm.cmd install.sh install.ps1 README.md LICENSE bash_completion $(BINDIR)/
	find $(BINDIR) -maxdepth 1 -name "install.*" -exec sed -i -e 's/latest/$(VERSION)/g' {} \;
	cp -R $(BINDIR) bin/dvm/latest

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

.PHONY: clean

clean:
	 -rm -fr bin/*
	 -rm dvm-helper/dvm-helper
	 -rm dvm-helper/dvm-helper.exe
