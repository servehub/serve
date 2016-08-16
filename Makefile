VERSION?="1.3.0"
DEST?=./bin

default: test

test:
	echo "==> Running tests..."
	go test -v `go list ./... | grep -v /vendor/`

build:
	echo "==> Build binaries..."
	go build -v -ldflags "-s -w -X main.version=${VERSION}" -o ${DEST}/serve serve.go
	go build -v -ldflags "-s -w -X main.version=${VERSION}" -o ${DEST}/serve-tools tools/cmd.go

install: test build
	echo "==> Copy binaries to \$GOPATH/bin/..."
	cp ${DEST}/* ${GOPATH}/bin/
