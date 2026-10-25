## Contributing

After doing your modifications make sure all tests pass, and add tests to ensure code coverage of your
new modifications.

### Running the test suite

To run the integration tests run:

```
./scripts/run_tests.sh
```

The `run_tests.sh` script installs `deltachat-rpc-server` 2.62.0 if it is missing
and rejects an installed server with a different version before running lint
or integration tests. The download is for x86_64 Linux; on other platforms,
install the matching server yourself.

The integration suite creates disposable accounts on `ci-chatmail.testrun.org`
and exchanges test messages. It requires network access. Local test cleanup
does not immediately delete the remote accounts.

To run a single test, for example `TestRpc_SetChatVisibility`:

```
cd v2/
go test -v ./... -run TestRpc_SetChatVisibility
```

### Updating dependencies

```
cd v2
go get -u ./...
```

### Updating the auto-generated bindings

To update the auto-generated RPC bindings code:

```
./scripts/update_rpc.sh
```

Run this from the repository root with the intended server in `PATH`.
The script uses a pinned `dcrpcgen` revision, reads the installed server's
`--openrpc` schema, and updates the expected server version in
`scripts/run_tests.sh` after generation succeeds. Review the generated diff for
removed methods, types, and fields; update affected tests and the changelog.

The 2.62.0 bindings were generated with `dcrpcgen` commit
`c9aefe03318edeae178ead5894526aec3e4d9175`.

For focused checks that do not create remote accounts or start mail I/O:

```sh
cd v2
go test ./deltachat -run '^(TestRpc_ServerSchema|TestRpc_Server262Requests|TestServer262.*|TestRunTestsServerVersion|TestRpc_GetAppVersion|TestRpc_IsSendingFinished)$' -count=1
```

`TestRpc_ServerSchema` compares the exported RPC method set with the installed
server's schema. The other focused checks cover new request parameters, errors,
response decoding, local account behavior, and the test runner's version check.
