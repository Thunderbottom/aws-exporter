.PHONY : build run fresh test clean

BIN := aws-exporter

build:
	go build -o ${BIN} -ldflags="-s -w"

clean:
	go clean
	- rm -f ${BIN}

dist:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(MAKE) build

fresh: clean build run

lint:
	find . -name "*.go" -exec ${GOPATH}/bin/golint {} \;

run:
	./${BIN}

test: lint
	go test
