#!/usr/bin/env bash
# Update auto-generated RPC bindings code.
# Also update the deltachat-rpc-server version used in run_tests.sh

pip install -U git+https://github.com/chatmail/dcrpcgen

SCRIPTS=$(dirname "${BASH_SOURCE[0]}")
VERSION=$(deltachat-rpc-server --version 2>&1 | tr -d '[:space:]')
echo $VERSION
sed -i -E "s|(download/)[^/]+(/deltachat-rpc-server)|\1v${VERSION}\2|g" "${SCRIPTS}/run_tests.sh"

deltachat-rpc-server --openrpc > schema.json

dcrpcgen go --schema schema.json -o ./v2/deltachat
gofmt -w .
