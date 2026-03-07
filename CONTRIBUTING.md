## Contributing

After doing your modifications make sure all tests pass, and add tests to ensure code coverage of your
new modifications.

### Running the test suite

To run the integration tests run:

```
./scripts/run_tests.sh
```

The `run_tests.sh` script will install `deltachat-rpc-server` (if needed) and run all tests.

To run a single test, for example `TestRpc_SetChatVisibility`:

```
cd v2/
go test -v ./... -run TestRpc_SetChatVisibility
```

### Updating the auto-generated bindings

To update the auto-generated RPC bindings code:

```
./scripts/update_rpc.sh
```
