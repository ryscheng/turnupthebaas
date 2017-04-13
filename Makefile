.PHONY: get-tools test lint unit

get-tools:
	go get -u github.com/kardianos/govendor
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install

test: lint unit

lint:
	gometalinter --vendor ./...
	#golint ./... | grep -v ^vendor || echo "golint: done"

unit:
	govendor test +local
