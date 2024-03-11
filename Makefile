docker_compose?= $(shell if which podman|grep -q .; then echo DOCKER_HOST="unix://$$XDG_RUNTIME_DIR/podman/podman.sock" docker-compose; else echo docker-compose; fi)
docker_user?=$(shell if echo ${docker}|grep -q podman; then echo 0:0; else echo ${uid}:${gid}; fi)
docker=$(shell if which podman|grep -q .; then echo podman; else echo docker; fi)
net_name=talek_net

.PHONY: get-tools test lint unit

get-tools:
	go get github.com/go-playground/overalls
	go get github.com/mattn/goveralls

test: lint unit

lint:
	go fmt ./...
	go vet ./...

unit:
	go test ./...

coverage:
	overalls -project=github.com/privacylab/talek -covermode=count -debug
	goveralls -coverprofile=overalls.coverprofile -service=travis-ci

ci: SHELL:=/bin/bash   # HERE: this is setting the shell for 'ci' only
ci: lint
	overalls -project=github.com/privacylab/talek -covermode=count -debug -- -tags travis
	@if [[ "${TRAVIS_JOB_NUMBER}" =~ ".1" ]]; then\
		echo "Uploading coverage to Coveralls.io"; \
		goveralls -coverprofile=overalls.coverprofile -service=travis-ci; \
	fi

docker-build.stamp:
	$(docker) build -t talek-base:latest ./
	#$(docker) build -t talek-replica:latest ./cli/talekreplica/
	touch $@

docker-bash:
	$(docker) run -it talek-base:latest bash

testnet-build-config: docker-build-cli.stamp
	$(docker) run --rm -v ./$(net_name):/talek_shared talek-cli bash -c "cd /talek_shared && talekutil --common --outfile common.json && \
	talekutil --replica --incommon common.json --private --index 1 --name replica1 --address :8081 --outfile replica1.json && \
	talekutil --replica --incommon common.json --private --index 2 --name replica2 --address :8082 --outfile replica2.json && \
	talekutil --replica --incommon common.json --private --index 3 --name replica3 --address :8083 --outfile replica3.json && \
	talekutil --trustdomain --infile replica1.json --outfile replica1.pub.json && \
	talekutil --trustdomain --infile replica2.json --outfile replica2.pub.json && \
	talekutil --trustdomain --infile replica3.json --outfile replica3.pub.json && \
	talekutil --client --infile common.json --trustdomains replica1.pub.json,replica2.pub.json,replica3.pub.json --outfile talek.json"

docker-build-cli.stamp: docker-build.stamp
	$(docker) build -t talek-cli:latest ./cli/
	touch $@

testnet-start: testnet-build-config
	cd $(net_name); DOCKER_USER=${docker_user} $(docker_compose) up --remove-orphans -d; $(docker_compose) top
	touch $(net_name)/running.stamp

testnet-stop:
	cd $(net_name) && DOCKER_USER=${docker_user} $(docker_compose) down --remove-orphans; rm -fv running.stamp

testnet-clean:
	$(docker) rmi talek-cli
	$(docker) rmi talek-base
	rm docker-build-cli.stamp
	rm docker-build.stamp

testnet-cli:
	$(docker) run --rm --network host -it -v ./$(net_name):/talek_shared talek-cli:latest bash
