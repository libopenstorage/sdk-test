[![Build Status](https://travis-ci.org/libopenstorage/sdk-test.svg?branch=master)](https://travis-ci.org/libopenstorage/sdk-test)

# sdk-test

## Updating

* Make sure to have updated (github.com/libopenstorage/openstorage-sdk-clients)[https://github.com/libopenstorage/openstorage-sdk-clients] with the latest.
* Type `dep ensure -update github.com/libopenstorage/openstorage-sdk-clients`
* Add the tests accordingly. The tests require that the latest container `openstorage/mock-sdk-server` has been created and pushed to Docker hub by the Travis builds of the master branch in `github.com/libopenstorage/openstorage`.
