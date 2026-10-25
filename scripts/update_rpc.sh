#!/usr/bin/env bash
# Update auto-generated RPC bindings code.
# Also update the deltachat-rpc-server version used in run_tests.sh
set -euo pipefail

pip install 'git+https://github.com/chatmail/dcrpcgen@c9aefe03318edeae178ead5894526aec3e4d9175'

SCRIPTS=$(dirname "${BASH_SOURCE[0]}")
VERSION=$(deltachat-rpc-server --version 2>&1)
echo "$VERSION"

deltachat-rpc-server --openrpc > schema.json

dcrpcgen go --schema schema.json -o ./v2/deltachat
gofmt -w ./v2/deltachat
sed -i -E "s|^RPC_SERVER_VERSION=.*$|RPC_SERVER_VERSION=${VERSION}|" "${SCRIPTS}/run_tests.sh"
