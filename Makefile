COMMIT = $(shell git rev-parse --verify --short HEAD)
VERSION = $(shell git describe --tags --dirty='-dev' 2> /dev/null)
GITHUB_ORG = getcarina
GITHUB_REPO = dvm
PACKAGE = github.com/${GITHUB_ORG}/${GITHUB_REPO}/dvm-helper

LDFLAGS = -w -X main.dvmCommit=${COMMIT} -X main.dvmVersion=${VERSION}

GOCMD = go
GOBUILD = $(GOCMD) build -a -tags netgo -ldflags '$(LDFLAGS)'

GOFILES = dvm-helper/*.go

default: dvm-helper

get-deps:
	go get ./...

#test: dvm-helper
#	go test -v
#	eval "$( ./dvm-helper --bash-completion )"
#	./dvm-helper --version

cross-build: get-deps dvm-helper linux darwin windows

dvm-helper: get-deps $(GOFILES)
	CGO_ENABLED=0 $(GOBUILD) -o dvm-helper/dvm-helper $(PACKAGE)

linux: $(GOFILES)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/dvm-helper-linux-amd64 $(PACKAGE)
	cd bin && shasum -a 256 dvm-helper-linux-amd64 > dvm-helper-linux-amd64.sha256

darwin: $(GOFILES)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o bin/dvm-helper-darwin-amd64 $(PACKAGE)
	cd bin && shasum -a 256 dvm-helper-darwin-amd64 > dvm-helper-darwin-amd64.sha256

windows: $(GOFILES)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o bin/dvm-helper-windows-amd64.exe $(PACKAGE)
	cd bin && shasum -a 256 dvm-helper-windows-amd64.exe > dvm-helper-windows-amd64.exe.sha256

############################ RELEASE TARGETS ############################

build-tagged-for-release: clean
	-docker rm -fv dvm-helper-build
	docker build -f Dockerfile.build -t dvm-helper-cli-build --no-cache=true .
	docker run --name dvm-helper-build dvm-helper-cli-build make tagged-build TAG=$(TAG)
	mkdir -p bin/
	docker cp dvm-helper-build:/built/bin .

checkout-tag:
	git checkout $(TAG)

# This one is intended to be run inside the accompanying Docker container
tagged-build: checkout-tag cross-build
	dvm-helper/dvm-helper --version
	mkdir -p /built/
	cp -r bin /built/bin

.PHONY: clean build-tagged-for-release checkout tagged-build

clean:
	 -rm -f bin/*
	 -rm dvm-helper/dvm-helper
	 -rm dvm-helper/dvm-helper.exe
