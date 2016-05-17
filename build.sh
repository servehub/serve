#!/bin/bash -ex

ROOT=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)

# install golang
if [ ! -x /var/go/.golang/go/bin/go ]; then
  wget https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz -O /tmp/go1.6.2.linux-amd64.tar.gz
  mkdir -p /var/go/.golang
  tar -C /var/go/.golang -xzvf /tmp/go1.6.2.linux-amd64.tar.gz
  rm -f /tmp/go1.6.2.linux-amd64.tar.gz
fi

# install glide (go deps manager)
if [ ! -x /var/go/.golang/go/bin/glide ]; then
  wget https://github.com/Masterminds/glide/releases/download/0.10.2/glide-0.10.2-linux-amd64.tar.gz -O /tmp/glide-0.10.2-linux-amd64.tar.gz
  tar -C /tmp/ -xzvf /tmp/glide-0.10.2-linux-amd64.tar.gz
  cp /tmp/linux-amd64/glide /var/go/.golang/go/bin/glide
  rm -f /tmp/glide-0.10.2-linux-amd64.tar.gz
  rm -rf /tmp/linux-amd64/
fi

export PATH=$PATH:/var/go/.golang/go/bin
export GOPATH=$ROOT/.gopath

WORK_DIR=$GOPATH/src/github.com/InnovaCo/serve
mkdir -p $WORK_DIR
cp -a ./* $WORK_DIR/

pushd $WORK_DIR

glide install --cache
go build -v -ldflags '-s -w' -o $ROOT/bin/serve serve.go
