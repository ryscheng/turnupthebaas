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
	$(docker) run --rm -v ./$(net_name):/talek_shared talek-cli bash -c "cd /talek_shared && talekutil --common --outfile common.conf && \
	talekutil --replica --incommon common.conf --private --index 0 --name replica0 --address http://127.0.0.1:8081 --outfile replica0.json && \
	talekutil --replica --incommon common.conf --private --index 1 --name replica1 --address http://127.0.0.1:8082 --outfile replica1.json && \
	talekutil --replica --incommon common.conf --private --index 2 --name replica2 --address http://127.0.0.1:8083 --outfile replica2.json && \
	talekutil --trustdomain --index 0 --name replica0 --address http://127.0.0.1:8081 --infile replica0.json --outfile replica0.pub.json && \
	talekutil --trustdomain --index 1 --name replica1 --address http://127.0.0.1:8082 --infile replica1.json --outfile replica1.pub.json && \
	talekutil --trustdomain --index 2 --name replica2 --address http://127.0.0.1:8083 --infile replica2.json --outfile replica2.pub.json && \
	talekutil --client --infile common.conf --trustdomains replica0.pub.json,replica1.pub.json,replica2.pub.json --outfile talek.conf && \
	sed -i -e 's/\"FrontendAddr\": \"\"/\"FrontendAddr\": \"http:\/\/127.0.0.1:8080\/rpc\"/' talek.conf"

docker-build-cli.stamp: docker-build.stamp
	$(docker) build -t talek-cli:latest ./cli/
	touch $@

testnet-start: testnet-build-config
	cd $(net_name); DOCKER_USER=${docker_user} $(docker_compose) up --remove-orphans -d; $(docker_compose) top
	touch $(net_name)/running.stamp

testnet-stop:
	-cd $(net_name) && DOCKER_USER=${docker_user} $(docker_compose) down --remove-orphans; rm -fv running.stamp

testnet-clean: testnet-stop
	-$(docker) rmi talek-cli
	-$(docker) rmi talek-base
	rm -f docker-build-cli.stamp
	rm -f docker-build.stamp

testnet-cli:
	$(docker) run --rm --network host -it -v ./$(net_name):/talek_shared -w /talek_shared talek-cli:latest bash

testnet-test-write-and-read:
	$(docker) run --rm --network host -it -v ./$(net_name):/talek_shared -w /talek_shared talek-cli:latest bash -c " \
	rm -f test_output.log && \
	talekclient --verbose --create --topic writehandle && \
	cp writehandle readhandle && \
	talekclient --verbose --topic writehandle --write Hello1 && \
	talekclient --verbose --topic writehandle --write Hello2 && \
	talekclient --verbose --topic writehandle --write Hello3 && \
	talekclient --verbose --topic readhandle --read | tee >(grep Hello >> test_output.log) && \
	talekclient --verbose --topic readhandle --read | tee >(grep Hello >> test_output.log) && \
	talekclient --verbose --topic readhandle --read | tee >(grep Hello >> test_output.log) && \
	cat test_output.log && \
	sha256sum -c <<<'81ff3368be913d26d96879a0215e4fbace25449f029c7aabf07350207f0aa980  test_output.log' \
		&& echo 'Successfully wrote and read three items!' \
		|| echo 'Failed to read something.'"
