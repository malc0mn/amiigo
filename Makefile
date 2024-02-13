SOURCEDIR=cmd
BINARY=amiigo
VERSION := 0 #$(shell git describe --tags)
BUILD_TIME := $(shell date +%FT%T%z)

LDFLAGS=-ldflags "-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"

.DEFAULT_GOAL: all

.PHONY: all
all: amiigo

amiigo:
	cd ${SOURCEDIR}; go build -trimpath ${LDFLAGS} -o ../${BINARY}

.PHONY: test
test:
	go test ./...

.PHONY: testv
testv:
	go test -v ./...

.PHONY: install
install:
	cd ${SOURCEDIR}; GOBIN=/usr/local/bin/ go install ${LDFLAGS}

.PHONY: clean
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
