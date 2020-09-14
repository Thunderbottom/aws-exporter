.PHONY : build run fresh test clean

BIN := aws-exporter

build:
	go build -o ${BIN}

clean:
	go clean
	- rm -f ${BIN}

fresh: clean build run

lint:
	find . -name "*.go" -exec ${GOPATH}/bin/golint {} \;

run:
	./${BIN}

test: lint
	go test
