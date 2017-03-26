VERSION?="1.4.0"
DEST?=./bin
SUFFIX?=""

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
	${GOPATH}/bin/go-bindata -pkg config -o manifest/config/config.go config/*.yml

build-serve:
	echo "==> Build serve binaries..."
	go build -v -ldflags "-s -w -X main.version=${VERSION}" -o ${DEST}/serve${SUFFIX} serve.go

build-serve-tools:
	echo "==> Build serve-tools binaries..."
	go build -v -ldflags "-s -w -X main.version=${VERSION}" -o ${DEST}/serve-tools${SUFFIX} tools/cmd.go

install: build-configs build-serve
	echo "==> Copy binaries to \$GOPATH/bin/..."
	cp ${DEST}/* ${GOPATH}/bin/

dist: build-configs
	GOOS=linux SUFFIX=-v${VERSION}-linux-amd64 make build-serve
	GOOS=darwin SUFFIX=-v${VERSION}-darwin-amd64 make build-serve
	GOOS=linux SUFFIX=-v${VERSION}-linux-amd64 make build-serve-tools
	GOOS=darwin SUFFIX=-v${VERSION}-darwin-amd64 make build-serve-tools

all: deps build-configs vet test build-serve build-serve-tools install
