#! /bin/bash

GOOS=linux GOARCH=amd64 go build -o container/main ./bin/tester/tester.go && \
docker build -t femaref/log ./container && \
docker run -it femaref/log:latest