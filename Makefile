VERSION?="1.4.6"
DEST?=./bin
SUFFIX?=""
TARGET_OS=linux darwin
TARGET_ARCH=amd64

default: install

vet:
	go vet `go list ./... | grep -v /vendor/`

test:
	go test -cover -v `go list ./... | grep -v /vendor/`

deps:
	@echo "==> Install dependencies..."
	go get -u github.com/jteeuwen/go-bindata/...

build-configs:
	${GOPATH}/bin/go-bindata -pkg config -o manifest/config/config.go config/*.yml

build-serve:
	go build -ldflags "-s -w -X main.version=${VERSION}" -o ${DEST}/serve${SUFFIX} serve.go

build-serve-tools:
	go build -ldflags "-s -w -X main.version=${VERSION}" -o ${DEST}/serve-tools${SUFFIX} tools/cmd.go

install: build-configs build-serve
	cp ${DEST}/{serve,serve-tools} ${GOPATH}/bin/

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
