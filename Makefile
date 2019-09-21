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

.PHONY: preview
preview: generate
	xdg-open gh-pages/index.html

.PHONY: generate
generate: generate-gh-pages generate-fixtures

.PHONY: generate-gh-pages
generate-gh-pages: build
	./etc/generate-gh-pages

.PHONY: generate-fixtures
generate-fixtures: build
	./etc/generate-fixtures

.PHONY: fuzz
fuzz: build
	@echo also see "http://lcamtuf.coredump.cx/afl/README.txt"
	go get github.com/dvyukov/go-fuzz/go-fuzz
	go get github.com/dvyukov/go-fuzz/go-fuzz-build
	mkdir -p fuzz fuzz/corpus
	cp org/testdata/*.org fuzz/corpus
	go-fuzz-build github.com/niklasfasching/go-org/org
	go-fuzz -bin=./org-fuzz.zip -workdir=fuzz
