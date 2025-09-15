# Chatmail API for Go

![Latest release](https://img.shields.io/github/v/tag/deltachat/deltachat-rpc-client-go?label=release)
[![Go Reference](https://pkg.go.dev/badge/github.com/deltachat/deltachat-rpc-client-go.svg)](https://pkg.go.dev/github.com/deltachat/deltachat-rpc-client-go)
[![CI](https://github.com/deltachat/deltachat-rpc-client-go/actions/workflows/ci.yml/badge.svg)](https://github.com/deltachat/deltachat-rpc-client-go/actions/workflows/ci.yml)
![Coverage](https://img.shields.io/badge/Coverage-72.0%25-brightgreen)
[![Go Report Card](https://goreportcard.com/badge/github.com/deltachat/deltachat-rpc-client-go)](https://goreportcard.com/report/github.com/deltachat/deltachat-rpc-client-go)

Chatmail client & bot API for Golang.

## Install

```sh
go get -u github.com/deltachat/deltachat-rpc-client-go
```

### Installing deltachat-rpc-server

This package depends on a standalone Chatmail RPC server
`deltachat-rpc-server` program that must be available in your
`PATH`. For installation instructions check:
https://github.com/chatmail/core/tree/main/deltachat-rpc-server

## Usage

Example echo-bot that will echo back any text message you send to
it:

<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/echobot_full/echobot.go) -->
<!-- The below code snippet is automatically added from ./examples/echobot_full/echobot.go -->
```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/deltachat/deltachat-rpc-client-go/deltachat"
	"github.com/deltachat/deltachat-rpc-client-go/deltachat/transport"
)

func logEvent(bot *deltachat.Bot, accId deltachat.AccountId, event deltachat.Event) {
	switch ev := event.(type) {
	case deltachat.EventInfo:
		log.Printf("INFO: %v", ev.Msg)
	case deltachat.EventWarning:
		log.Printf("WARNING: %v", ev.Msg)
	case deltachat.EventError:
		log.Printf("ERROR: %v", ev.Msg)
	}
}

func runEchoBot(bot *deltachat.Bot, accId deltachat.AccountId) {
	sysinfo, _ := bot.Rpc.GetSystemInfo()
	log.Println("Running deltachat core", sysinfo["deltachat_core_version"])

	bot.On(deltachat.EventInfo{}, logEvent)
	bot.On(deltachat.EventWarning{}, logEvent)
	bot.On(deltachat.EventError{}, logEvent)
	bot.OnNewMsg(func(bot *deltachat.Bot, accId deltachat.AccountId, msgId deltachat.MsgId) {
		msg, _ := bot.Rpc.GetMessage(accId, msgId)
		if msg.FromId > deltachat.ContactLastSpecial {
			bot.Rpc.MiscSendTextMessage(accId, msg.ChatId, msg.Text)
		}
	})

	if isConf, _ := bot.Rpc.IsConfigured(accId); !isConf {
		log.Println("Bot not configured, configuring...")
		err := bot.Configure(accId, os.Args[1], os.Args[2])
		if err != nil {
			log.Fatalln(err)
		}
	}

	addr, _ := bot.Rpc.GetConfig(accId, "configured_addr")
	log.Println("Listening at:", addr.Unwrap())
	bot.Run()
}

func main() {
	trans := transport.NewIOTransport()
	trans.Open()
	defer trans.Close()
	rpc := &deltachat.Rpc{Context: context.Background(), Transport: trans}
	runEchoBot(deltachat.NewBot(rpc), deltachat.GetAccount(rpc))
}
```
<!-- MARKDOWN-AUTO-DOCS:END -->

Save the previous code snippet as `echobot.go` then run:

```sh
go mod init echobot; go mod tidy
go run ./echobot.go bot@example.com PASSWORD
```

Check the [examples folder](./examples)
for more examples.

## Developing bots faster ⚡

If what you want is to develop bots, you probably should use this
library together with [deltabot-cli-go](deltabotcli), it takes
away the repetitive process of creating the bot CLI and let you
focus on writing your message processing logic.

## Testing your code

`deltachat.AcFactory` is provided to help users of this library
to unit-test their code.

### Local mail server

You need to have a local fake email server running. The easiest
way to do that is with Docker:

```
$ docker pull ghcr.io/deltachat/mail-server-tester:release
$ docker run -it --rm -p 3025:25 -p 3110:110 -p 3143:143 -p 3465:465 -p 3993:993 ghcr.io/deltachat/mail-server-tester
```

### Using AcFactory

After setting up the fake email server, create a file called `main_test.go` inside your tests folder,
and save it with the following content:

<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/echobot_full/main_test.go) -->
<!-- The below code snippet is automatically added from ./examples/echobot_full/main_test.go -->
```go
package main // replace with your package name

import (
	"testing"

	"github.com/deltachat/deltachat-rpc-client-go/deltachat"
)

var acfactory *deltachat.AcFactory

func TestMain(m *testing.M) {
	acfactory = &deltachat.AcFactory{}
	acfactory.TearUp()
	defer acfactory.TearDown()
	m.Run()
}
```
<!-- MARKDOWN-AUTO-DOCS:END -->

Now in your other test files you can do:

<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/echobot_full/echobot_test.go) -->
<!-- The below code snippet is automatically added from ./examples/echobot_full/echobot_test.go -->
```go
package main // replace with your package name

import (
	"testing"

	"github.com/deltachat/deltachat-rpc-client-go/deltachat"
	"github.com/stretchr/testify/assert"
)

func TestEchoBot(t *testing.T) {
	acfactory.WithOnlineBot(func(bot *deltachat.Bot, botAcc deltachat.AccountId) {
		go runEchoBot(bot, botAcc) // this is the function we are testing
		acfactory.WithOnlineAccount(func(uRpc *deltachat.Rpc, uAccId deltachat.AccountId) {
			chatId := acfactory.CreateChat(uRpc, uAccId, bot.Rpc, botAcc)
			uRpc.MiscSendTextMessage(uAccId, chatId, "hi")
			msg := acfactory.NextMsg(uRpc, uAccId)
			assert.Equal(t, "hi", msg.Text) // check that bot echoes back the "hi" message from user
		})
	})
}
```
<!-- MARKDOWN-AUTO-DOCS:END -->

### GitHub action

To run the tests in a GitHub action with the fake mail server service:

<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/echobot_full/.github/workflows/ci.yml) -->
<!-- The below code snippet is automatically added from ./examples/echobot_full/.github/workflows/ci.yml -->
```yml
name: Test

on:
  push:
    branches: [ "master" ]
  pull_request:
    branches: [ "master" ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v3
        with:
          go-version: 1.19

      - name: Run tests
        run: |
          go test -v

    services:
      mail_server:
        image: ghcr.io/deltachat/mail-server-tester:release
        ports:
          - 3025:25
          - 3143:143
          - 3465:465
          - 3993:993
```
<!-- MARKDOWN-AUTO-DOCS:END -->

Check the complete example at [examples/echobot_full](./examples/echobot_full)

## Contributing

Pull requests are welcome! check [CONTRIBUTING.md](./CONTRIBUTING.md)


[deltabotcli]: https://github.com/deltachat-bot/deltabot-cli-go/