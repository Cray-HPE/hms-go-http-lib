NAME ?= hms-go-http-lib
VERSION ?= $(shell cat .version)

all: image unittest coverage

image:
	docker build --pull ${DOCKER_ARGS} --tag '${NAME}:${VERSION}' .

unittest:
	./runUnitTest.sh

coverage:
	./runCoverage.sh

buildbase:
	docker build -t cray/hms-go-http-lib-build-base -f Dockerfile.build-base .

