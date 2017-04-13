.PHONY: get-tools test lint

get-tools:
	go get -u github.com/kardianos/govendor
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install

test: lint unittest

unittest:
	govendor test +local

lint:
	gometalinter --vendor ./...
	#golint ./... | grep -v ^vendor || echo "golint: done"

