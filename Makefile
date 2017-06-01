SHELL:=/bin/bash
VERSION?=$(shell git describe --tags --abbrev=0 | sed 's/v//')
DEST?=./bin
SUFFIX?=""
TARGET_OS=linux darwin
TARGET_ARCH=amd64
PACKAGE=github.com/servehub/serve

export CGO_ENABLED=0

default: install

deps:
	@echo "==> Install dependencies..."
	go get github.com/Masterminds/glide
	glide i -v
	go get github.com/jteeuwen/go-bindata/...
	go get github.com/alecthomas/gometalinter
	gometalinter --install

build-configs:
	${GOPATH}/bin/go-bindata -pkg config -o manifest/config/config.go config/*.yml

lint: build-configs
	gometalinter --config=gometalinter.json --fast ./...

test: build-configs
	go test -cover -v `go list ./... | grep -v /vendor/`

test-manifests: build-configs build-serve build-serve-tools
	for file in `ls ${PWD}/tests/manifests/*.yml`; do \
		${DEST}/serve-tools test-runner --file $$file --serve ${DEST}/serve --config-path=${PWD}/tests/; \
	done

build-serve:
	go build -ldflags "-s -w -X main.version=${VERSION}" -o ${DEST}/serve${SUFFIX} serve.go

build-serve-tools:
	go build -ldflags "-s -w -X main.version=${VERSION}" -o ${DEST}/serve-tools${SUFFIX} tools/cmd.go

install: build-configs build-serve
	for f in serve serve-tools; do \
		if [ -f ${DEST}/$$f ]; then cp ${DEST}/$$f ${GOPATH}/bin/; fi \
	done

clean:
	@echo "==> Cleanup old binaries..."
	rm -f ${DEST}/*

dist: clean build-configs
	@echo "==> Build dist..."

	for GOOS in ${TARGET_OS}; do \
		for GOARCH in ${TARGET_ARCH}; do \
			GOOS=$$GOOS GOARCH=$$GOARCH SUFFIX=-v${VERSION}-$$GOOS-$$GOARCH make build-serve; \
		done \
	done

docker-dist:
	docker run --rm -v "${PWD}":/go/src/${PACKAGE} -w /go/src/${PACKAGE} golang:1.8 /bin/sh -c 'make deps && make dist'

bump-tag:
	TAG=$$(echo "v${VERSION}" | awk -F. '{$$NF = $$NF + 1;} 1' | sed 's/ /./g'); \
	git tag $$TAG; \
	git push && git push --tags

release: dist
	@echo "==> Create github release and upload files..."

	-github-release release \
		--user servehub \
		--repo serve \
		--tag v${VERSION}

	for GOOS in ${TARGET_OS}; do \
		for GOARCH in ${TARGET_ARCH}; do \
			github-release upload \
				--user servehub \
				--repo serve \
				--tag v${VERSION} \
				--name serve-v${VERSION}-$$GOOS-$$GOARCH \
				--file ${DEST}/serve-v${VERSION}-$$GOOS-$$GOARCH \
				--replace; \
		done \
	done
