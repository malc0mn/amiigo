SOURCEDIR=cmd
BINARY=amiigo
VERSION := 0 #$(shell git describe --tags)
BUILD_TIME := $(shell date +%FT%T%z)

LDFLAGS=-ldflags "-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"
# TODO use trimpath: go build -gcflags=-trimpath=$GOPATH -asmflags=-trimpath=$GOPATH
.DEFAULT_GOAL: all

.PHONY: all
all: amiigo

amiigo:
	cd cmd; go build ${LDFLAGS} -o ../${BINARY}

.PHONY: test
test:
	go test ./...

.PHONY: install
install:
	cd cmd; GOBIN=/usr/local/bin/ go install ${LDFLAGS}

.PHONY: clean
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
