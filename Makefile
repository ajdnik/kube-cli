BINARY=kube-cli

VERSION=`git describe --tags`
BUILD=`date +%FT%T%z`

LDFLAGS=-ldflags "-X github.com/ajdnik/kube-cli/version.version=${VERSION} -X github.com/ajdnik/kube-cli/version.build=${BUILD}"

build:
	go build ${LDFLAGS} -o ${BINARY}

install:
	go install ${LDFLAGS}

clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

default: install

.PHONY: build clean install
