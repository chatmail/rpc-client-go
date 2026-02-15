#!/usr/bin/env bash

echo "Checking code with gofmt..."
OUTPUT=`gofmt -d .`
if [ -n "$OUTPUT" ]
then
    echo "$OUTPUT"
    exit 1
fi

echo "Checking code with golangci-lint..."
if ! command -v golangci-lint &> /dev/null
then
    echo "golangci-lint not found, installing..."
    # binary will be $(go env GOPATH)/bin/golangci-lint
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.4.0
fi

if ! golangci-lint run
then
    exit 1
fi

if ! command -v deltachat-rpc-server &> /dev/null
then
    echo "deltachat-rpc-server not found, installing..."
    curl -L https://github.com/chatmail/core/releases/download/v2.42.0/deltachat-rpc-server-x86_64-linux --output deltachat-rpc-server
    chmod +x deltachat-rpc-server
    export PATH=`pwd`:"$PATH"
fi

if ! command -v courtney &> /dev/null
then
    echo "courtney not found, installing..."
    go install github.com/dave/courtney@master
fi

for i in ./examples/basic/*.go
do
    echo "Testing examples: $i"
    if ! go build -v "$i"
    then
        exit 1
    fi
done
echobot_full="examples/echobot_full/"
echo "Testing examples: $echobot_full"
cd $echobot_full
if ! go test -v
then
    exit 1
fi
cd ../..
echo "Done testing examples"

courtney -v -t="./..." ${TEST_EXTRA_TAGS:--t="-parallel=1"}
go tool cover -func=coverage.out -o=coverage-percent.out
