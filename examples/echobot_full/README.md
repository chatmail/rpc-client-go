# Echo-bot example

This is an example echo-bot project with unit-tests using the `acfactory` package.

## Running the tests

You need to have a local fake email server running. The easiest way to do that is with Docker:

```
$ docker pull ghcr.io/deltachat/mail-server-tester:release
$ docker run -it --rm -p 3025:25 -p 3110:110 -p 3143:143 -p 3465:465 -p 3993:993 ghcr.io/deltachat/mail-server-tester:release
```

Then you can run the tests:

```
go test -v
```
