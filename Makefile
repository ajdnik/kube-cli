BINARY=kube-cli
VERSION="0.1.6"
BUILD=`date +%FT%T%z`
LDFLAGS=-ldflags "-X github.com/ajdnik/kube-cli/version.version=${VERSION} -X github.com/ajdnik/kube-cli/version.build=${BUILD}"

build:
	go build ${LDFLAGS} -o ${BINARY}

install:
	go install ${LDFLAGS}

clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

changelog:
	git-chglog -c .chglog/changelog/config.yml -o CHANGELOG.md --next-tag ${VERSION} ..${VERSION}

deps:
	go get github.com/inconshreveable/mousetrap
	go get -u github.com/spf13/cobra
	go get -u github.com/git-chglog/git-chglog/cmd/git-chglog
	go get github.com/mitchellh/gox
	go get github.com/c4milo/github-release
	go get github.com/dustin/go-humanize
	go get github.com/denormal/go-gitignore
	go get github.com/fatih/color
	go get gopkg.in/yaml.v2
	go get -u cloud.google.com/go/storage
	go get github.com/briandowns/spinner

compile:
	@rm -rf build/
	@gox ${LDFLAGS} \
	-osarch="darwin/amd64" \
	-osarch="linux/amd64" \
	-osarch="windows/amd64" \
	-output "build/{{.Dir}}_{{.OS}}_{{.Arch}}/$(BINARY)" \
	./...

dist: compile
	$(eval FILES := $(shell ls build))
	@rm -rf dist && mkdir dist
	@for f in $(FILES); do \
		(cd $(shell pwd)/build/$$f && tar -cvzf ../../dist/$$f.tar.gz *); \
		(cd $(shell pwd)/dist && shasum -a 512 $$f.tar.gz > $$f.sha512); \
		echo $$f; \
	done

release: dist changelog
	git add CHANGELOG.md
	git commit -m "chore: updated changelog"
	git add Makefile
	git commit -m "chore: version bumped"
	git push
	git-chglog -c .chglog/release/config.yml -o RELEASE.md --next-tag ${VERSION} ${VERSION}
	github-release ajdnik/$(BINARY) $(VERSION) "$$(git rev-parse --abbrev-ref HEAD)" "## Changelog<br/>$$(cat RELEASE.md)" 'dist/*'
	@rm RELEASE.md
	git pull

default: build

.PHONY: dist release changelog compile deps build clean install
