## Contributing

After doing your modifications make sure all tests pass, and add tests to ensure code coverage of your
new modifications.

### Running the test suite

To run the integration tests run:

```
$ ./scripts/run_tests.sh
```

The `run_tests.sh` script will install `deltachat-rpc-server` (if needed) and run all tests.

To run all `Account` tests:

```
go test -v ./... -run TestAccount
```

To run a single test, for example `TestChat_SetName`:

```
go test -v ./... -run TestChat_SetName
```
