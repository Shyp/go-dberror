.PHONY: install test

install:
	go install ./...

test:
	go test ./...

test-install: 
	go get bitbucket.org/liamstask/goose/cmd/goose
	go get github.com/letsencrypt/boulder/test
	goose up
