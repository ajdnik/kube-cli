BINARY=kube-cli

VERSION="0.1.2"
BUILD=`date +%FT%T%z`

LDFLAGS=-ldflags "-X github.com/ajdnik/kube-cli/version.version=${VERSION} -X github.com/ajdnik/kube-cli/version.build=${BUILD}"

build:
	go build ${LDFLAGS} -o ${BINARY}

install:
	go install ${LDFLAGS}

clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

changelog:
	git-chglog -c .chglog/changelog/config.yml -o CHANGELOG.md

deps:
	go get github.com/inconshreveable/mousetrap
	go get -u github.com/spf13/cobra
	go get -u github.com/git-chglog/git-chglog/cmd/git-chglog
	go get github.com/mitchellh/gox
	go get github.com/c4milo/github-release

compile:
	@rm -rf build/
	@gox ${LDFLAGS} \
	-osarch="darwin/amd64" \
	-os="linux" \
	-os="windows" \
	-os="solaris" \
	-output "build/{{.Dir}}_$(VERSION)_{{.OS}}_{{.Arch}}/$(BINARY)" \
	./...

dist: compile
	$(eval FILES := $(shell ls build))
	@rm -rf dist && mkdir dist
	@for f in $(FILES); do \
		(cd $(shell pwd)/build/$$f && tar -cvzf ../../dist/$$f.tar.gz *); \
		(cd $(shell pwd)/dist && shasum -a 512 $$f.tar.gz > $$f.sha512); \
		echo $$f; \
	done

release: dist
	@latest_tag=$$(git describe --tags `git rev-list --tags --max-count=1`); \
	comparison="$$latest_tag..HEAD"; \
	if [ -z "$$latest_tag" ]; then comparison=""; fi; \
	git-chglog -c .chglog/release/config.yml -o RELEASE.md $$comparison \
	changelog=$$(cat RELEASE.md); \
	@rm RELEASE.md \
	github-release ajdnik/$(BINARY) $(VERSION) "$$(git rev-parse --abbrev-ref HEAD)" "**Changelog**<br/>$$changelog" 'dist/*'; \
	git pull

default: install

.PHONY: dist release changelog compile deps build clean install
