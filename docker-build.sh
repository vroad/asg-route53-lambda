#!/bin/bash -eu
SOURCE=/go/src/github.com/vroad/asg-route53-lambda
docker run --rm -v "$PWD:$SOURCE" -w "$SOURCE" \
  golang:1.15 bash -c 'GOOS=linux GOARCH=amd64 go build -o main'
