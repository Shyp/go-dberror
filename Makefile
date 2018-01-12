.PHONY: install test

GOOSE := $(GOPATH)/bin/goose

install:
	go install ./...

test:
	go test -race ./... -timeout 2s

$(GOOSE):
	go get -u github.com/kevinburke/goose/cmd/goose

test-install: | $(GOOSE)
	-createdb dberror
	go get -u github.com/letsencrypt/boulder/test
	go get -u ./...
	$(GOOSE) up
