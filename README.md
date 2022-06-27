[![Build Status](https://travis-ci.org/servehub/serve.svg?branch=master)](https://travis-ci.org/servehub/serve)

# Installing

## Go

Make sure you have Go 1.18.+ installed:

```sh
brew install go
```

## Dependencies

```sh
# install dependencies via makefile script
make deps

# set $GOPATH env variable (should be your home directory)
export GOPATH=${HOME}/go

# download dependency 
go mod vendor

# build
make

# test to ensure everything is working
make test
```

