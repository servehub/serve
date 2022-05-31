[![Build Status](https://travis-ci.org/servehub/serve.svg?branch=master)](https://travis-ci.org/servehub/serve)

# Installing

## Go

Make sure you have Go installed:

```sh
brew install go
```

## Dependencies

```sh
# install dependencies via makefile script
make deps

# set $GOPATH env variable (should be your home directory)
export GOPATH=${HOME}/go

# build
make

# test to ensure everything is working
make test
```

# Testing

## Coverage report plugin

```sh
# run coverage uploader plugin (using "project" as an example)
DATABASE_URL="postgres://postgres:postgres@localhost:5432" \
go run serve.go test.coverage \
    --manifest=project/manifest.yml \
    --coverage-file=test.exec \
    --repo=my-repo \
    --branch=main \
    --ref=abc123 \
    --test-type=unit
```

```sql
-- view binary data (postgres)
SELECT encode(coverage_file::bytea, 'escape')
FROM public.coverage_reports
ORDER BY id ASC
```
