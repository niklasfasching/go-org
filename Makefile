.PHONY: default
default: test

go_files=$(shell find . -name '*.go' ! -path './docs/*')

go-org: $(go_files) go.mod go.sum
	go get -d ./...
	go build .

build: go-org

.PHONY: test
test:
	go get -d -t ./...
	go test ./... -v

.PHONY: setup
setup:
	git config core.hooksPath etc/githooks
	command -v go > /dev/null || (echo "go not installed" && false)

.PHONY: preview
preview: generate
	xdg-open docs/index.html

.PHONY: generate
generate: generate-gh-pages generate-fixtures

.PHONY: generate-gh-pages
generate-gh-pages: build
	./etc/generate-gh-pages

.PHONY: generate-fixtures
generate-fixtures: build
	./etc/generate-fixtures

.PHONY: serve-gh-pages
serve-gh-pages: generate-gh-pages
	cd docs && mkdir go-org && mv * go-org 2> /dev/null || true
	cd docs && python3 -m http.server

.PHONY: fuzz
fuzz: build
	@echo also see "http://lcamtuf.coredump.cx/afl/README.txt"
	go get github.com/dvyukov/go-fuzz/go-fuzz
	go get github.com/dvyukov/go-fuzz/go-fuzz-build
	mkdir -p fuzz fuzz/corpus
	cp org/testdata/*.org fuzz/corpus
	go-fuzz-build github.com/niklasfasching/go-org/org
	go-fuzz -bin=./org-fuzz.zip -workdir=fuzz
