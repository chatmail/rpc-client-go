#!/usr/bin/env bash

pip install -U git+https://github.com/chatmail/dcrpcgen
deltachat-rpc-server --version
deltachat-rpc-server --openrpc > schema.json
dcrpcgen go --schema schema.json -o ./v2/deltachat
gofmt -w .
