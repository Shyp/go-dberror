.PHONY: install test

install:
	go install ./...

test:
	go test ./...

test-install: 
	go get bitbucket.org/liamstask/goose/cmd/goose
	goose up
