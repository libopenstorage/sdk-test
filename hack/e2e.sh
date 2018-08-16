#!/bin/bash

fail() {
	echo "$1"
	exit 1
}

MOCKSDK="mocksdk"
PORT="8180"
GWPORT="8181"

docker pull openstorage/mock-sdk-server
docker stop ${MOCKSDK} > /dev/null 2>&1
docker run --rm --name ${MOCKSDK} -d -p ${PORT}:9100 -p ${GWPORT}:9110 openstorage/mock-sdk-server || fail "Unable to start server"
timeout 30 sh -c 'until curl --silent -X GET -d {} http://localhost:8181/v1/clusters/current | grep STATUS_OK; do sleep 1; done'
./cmd/sdk-test/sdk-test --sdk.endpoint=127.0.0.1:${PORT} --sdk.cpg=./cmd/sdk-test/cb.yaml ; ret=$?
docker stop ${MOCKSDK} > /dev/null 2>&1
exit $ret
