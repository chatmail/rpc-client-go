#!/usr/bin/env bash
set -euo pipefail

RPC_SERVER_VERSION=2.62.0

echo "Checking code with gofmt..."
OUTPUT=`gofmt -d .`
if [ -n "$OUTPUT" ]
then
    echo "$OUTPUT"
    exit 1
fi

if ! command -v deltachat-rpc-server &> /dev/null
then
    echo "deltachat-rpc-server not found, installing..."
    curl -fL "https://github.com/chatmail/core/releases/download/v${RPC_SERVER_VERSION}/deltachat-rpc-server-x86_64-linux" --output deltachat-rpc-server
    chmod +x deltachat-rpc-server
    export PATH=`pwd`:"$PATH"
fi

SERVER_VERSION=$(deltachat-rpc-server --version 2>&1)
if [ "$SERVER_VERSION" != "$RPC_SERVER_VERSION" ]; then
    echo "Expected deltachat-rpc-server $RPC_SERVER_VERSION, found $SERVER_VERSION" >&2
    exit 1
fi

echo "Checking code with golangci-lint..."
if ! command -v golangci-lint &> /dev/null
then
    echo "golangci-lint not found, installing..."
    # binary will be $(go env GOPATH)/bin/golangci-lint
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/main/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.12.2
fi

(cd v2 && golangci-lint run)

for i in examples/*
do
    echo "Testing: $i"
    cd "$i"
    go mod edit -replace=github.com/chatmail/rpc-client-go/v2=../../v2
    go mod tidy
    golangci-lint run
    go build -v
    go test -v
    go mod edit -dropreplace github.com/chatmail/rpc-client-go/v2
    cd ../..
done
echo "Done testing examples"

cd v2
# add -parallel=1 to avoid running tests in parallel
go test -v ./... -coverprofile coverage.out
go tool cover -func=coverage.out -o=../coverage-percent.out
