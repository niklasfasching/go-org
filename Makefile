.PHONY: default
default: test

.PHONY: install
install:
	go get -t ./...

.PHONY: build
build: install
	go build main.go

.PHONY: test
test: install
	go test ./... -v

.PHONY: setup
setup:
	git config core.hooksPath etc/githooks

case=example
.PHONY: render
render:
	bash -c "go run main.go org/testdata/$(case).org html |& html2text"
