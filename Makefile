.PHONY: get-tools test lint unit

get-tools:
	go get -u github.com/kardianos/govendor
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install

test: lint unit

lint:
	gometalinter --vendor --tests --deadline=60s \
		--disable-all \
		--enable=vet \
		--enable=vetshadow \
		--enable=golint \
		--enable=ineffassign \
		--enable=goconst \
		./...
	#golint ./... | grep -v ^vendor || echo "golint: done"

unit:
	govendor test +local
