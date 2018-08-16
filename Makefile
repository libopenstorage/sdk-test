all: sdk-test

sdk-test:
	$(MAKE) -C cmd/sdk-test

install:
	$(MAKE) -C cmd/sdk-test install

test: install
	./hack/e2e.sh
