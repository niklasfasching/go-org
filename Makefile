.PHONY: default
default: test

.PHONY: install
install:
	go get -t ./...

.PHONY: build
build: install
	go build .

.PHONY: test
test: install
	go test ./... -v

.PHONY: setup
setup:
	git config core.hooksPath etc/githooks

case=example
.PHONY: render
render:
	go run main.go org/testdata/$(case).org html | html2text

.PHONY: generate-gh-pages
generate-gh-pages: build
	./etc/generate-gh-pages

.PHONY: generate-html-fixtures
generate-html-fixtures: build
	./etc/generate-html-fixtures
