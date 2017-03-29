VERSION := `cat VERSION`
LDFLAGS := -X "main.Version=$(VERSION)"
GO_BUILD := go build -v -ldflags "$(LDFLAGS)"

all: lint test build

build: deps
	$(GO_BUILD) -o build/appcop .

clean:
	go clean -v .
	rm -rf build

debug: deps
	$(GO_BUILD) -race -tags 'debug' -o build/appcop .

deps:
	go get -u github.com/Masterminds/glide
	glide install

test-deps:
	go get -u github.com/jstemmer/go-junit-report
	mkdir -p build/test-results

lint: deps lint-deps onlylint

lint-deps:
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install

release: lint test
	GOARCH=amd64 GOOS=linux $(GO_BUILD) -o build/appcop .

test: deps test-deps onlytest

pack: test lint build
	docker build -t appcop . && mkdir -p dist && docker run -v ${PWD}/dist:/work/dist appcop

onlytest: build
	go test -cover -v $$(go list ./... | grep -v /vendor/) | tee build/test-results/report.log
	cat build/test-results/report.log | go-junit-report -set-exit-code > ../tests.xml

onlylint: build
	gometalinter \
	--deadline=90s \
	--disable=dupl \
	--disable=gotype \
	--disable=interfacer \
	--disable=vetshadow \
	--vendor ./... \
	--exclude="lookup_unix.go"
