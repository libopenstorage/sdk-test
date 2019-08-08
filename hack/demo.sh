#!/bin/bash

fail() {
	echo "$1"
	exit 1
}

PORT="9020"
GWPORT="8181"
make install
./cmd/sdk-test/sdk-test --ginkgo.focus="Security" --ginkgo.v --sdk.endpoint=70.0.74.173:${PORT} --sdk.issuer="openstorage.io" --sdk.sharedsecret="Password1"
