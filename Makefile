.PHONY: default
default: test

.PHONY: test
test:
	go test ./... -v

.PHONY: build
build:
	go build cmd/org/*

.PHONY: setup
setup:
	git config core.hooksPath etc/githooks
