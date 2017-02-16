VERSION?="1.3.0"
DEST?=./bin

default: install

vet:
	echo "==> Running vet..."
	go vet `go list ./... | grep -v /vendor/`

test:
	echo "==> Running tests..."
	go test -cover -v `go list ./... | grep -v /vendor/`

deps:
	echo "==> Install dependencies..."
	go get -u github.com/jteeuwen/go-bindata/...

build-configs:
	echo "==> Build configs..."
	${GOPATH}/bin/go-bindata -pkg config -o config/config.go config/*.yml

build-serve:
	echo "==> Build serve binaries..."
	go build -v -ldflags "-s -w -X main.version=${VERSION}" -o ${DEST}/serve serve.go

build-serve-tools:
	echo "==> Build serve-tools binaries..."
	go build -v -ldflags "-s -w -X main.version=${VERSION}" -o ${DEST}/serve-tools tools/cmd.go

install: build-serve
	echo "==> Copy binaries to \$GOPATH/bin/..."
	cp ${DEST}/* ${GOPATH}/bin/

all: deps build-configs vet test build-serve build-serve-tools install
