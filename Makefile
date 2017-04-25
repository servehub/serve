SHELL:=/bin/bash
VERSION?=$$(git describe --tags --abbrev=0)
DEST?=./bin
SUFFIX?=""
TARGET_OS=linux darwin
TARGET_ARCH=amd64
PACKAGE=github.com/servehub/serve

export CGO_ENABLED=0

default: install

vet:
	go vet `go list ./... | grep -v /vendor/`

test:
	go test -cover -v `go list ./... | grep -v /vendor/`

deps:
	@echo "==> Install dependencies..."
	go get github.com/Masterminds/glide
	go get github.com/jteeuwen/go-bindata/...
	glide i -v

build-configs:
	${GOPATH}/bin/go-bindata -pkg config -o manifest/config/config.go config/*.yml

build-serve:
	go build -ldflags "-s -w -X main.version=${VERSION}" -o ${DEST}/serve${SUFFIX} serve.go

build-serve-tools:
	go build -ldflags "-s -w -X main.version=${VERSION}" -o ${DEST}/serve-tools${SUFFIX} tools/cmd.go

install: build-configs build-serve
	cp ${DEST}/serve ${GOPATH}/bin/

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
