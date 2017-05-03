.PHONY: get-tools test lint unit

get-tools:
	go get -u github.com/kardianos/govendor
	go get github.com/go-playground/overalls
	go get github.com/mattn/goveralls
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install

test: lint unit

ci: lint coverage

lint:
	gometalinter --vendor --tests --deadline=60s \
		--disable-all \
		--enable=gofmt \
		--enable=vet \
		--enable=vetshadow \
		--enable=golint \
		--enable=ineffassign \
		--enable=goconst \
		./...
	#golint ./... | grep -v ^vendor || echo "golint: done"

unit:
	govendor test +local

coverage:
	overalls -project=github.com/privacylab/talek -covermode=count -debug -- -tags travis
	goveralls -coverprofile=overalls.coverprofile -service=travis-ci

docker-build:
	docker build -t talek-base:latest ./
	docker build -t talek-replica:latest ./cli/talekreplica/

docker-bash:
	docker run -it talek-base:latest bash
