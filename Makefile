VERSION := `cat VERSION`
LDFLAGS := -X "main.Version=$(VERSION)"
GO_BUILD := go build -v -ldflags "$(LDFLAGS)"

CURRENTDIR = $(shell pwd)

COVERAGEDIR = $(CURRENTDIR)/coverage
PACKAGES = $(shell go list ./... | grep -v /vendor/)
TEST_TARGETS = $(PACKAGES)

all: lint test build

build: deps
	$(GO_BUILD) -o build/appcop .

clean:
	go clean -v .
	rm -rf build

debug: deps
	$(GO_BUILD) -race -tags 'debug' -o build/appcop .

deps:
	@mkdir -p $(COVERAGEDIR)
	@go get github.com/modocache/gover
	@go get -u github.com/Masterminds/glide
	@glide install

lint: deps lint-deps onlylint

lint-deps:
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install

release: lint test
	GOARCH=amd64 GOOS=linux $(GO_BUILD) -o build/appcop .

test: deps $(SOURCES) $(TEST_TARGETS)
	gover $(COVERAGEDIR) $(COVERAGEDIR)/gover.coverprofile

$(TEST_TARGETS):
	go test -coverprofile=coverage/$(shell basename $@).coverprofile $@

pack: test lint build
	docker build -t appcop . && mkdir -p dist && docker run -v ${PWD}/dist:/work/dist appcop

onlylint: build
	gometalinter \
	--deadline=720s \
	--disable=dupl \
	--disable=gotype \
	--disable=gas \
	--disable=golint \
	--enable=gofmt \
	--vendor
