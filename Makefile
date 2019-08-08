export BASE_DIR=$(shell git rev-parse --show-toplevel)
export GOPATH=$(shell echo ${BASE_DIR}| sed 's@\(.*\)/src/github.com.*@\1@g')
all: sdk-test

sdk-test:
	$(MAKE) -C cmd/sdk-test

install:
	$(MAKE) -C cmd/sdk-test install

test: install
	./hack/e2e.sh
