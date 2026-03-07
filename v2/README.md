# Chatmail API for Go

[![Latest release](https://img.shields.io/github/v/tag/chatmail/rpc-client-go?label=release)](https://pkg.go.dev/github.com/chatmail/rpc-client-go/v2)
[![Go Reference](https://pkg.go.dev/badge/github.com/chatmail/rpc-client-go/v2.svg)](https://pkg.go.dev/github.com/chatmail/rpc-client-go/v2)
[![CI](https://github.com/chatmail/rpc-client-go/actions/workflows/ci.yml/badge.svg)](https://github.com/chatmail/rpc-client-go/actions/workflows/ci.yml)
![Coverage](https://img.shields.io/badge/Coverage-89.8%25-brightgreen)
[![Go Report Card](https://goreportcard.com/badge/github.com/chatmail/rpc-client-go/v2)](https://goreportcard.com/report/github.com/chatmail/rpc-client-go/v2)

Chatmail client & bot API for Golang.

## Install

```sh
go get -u github.com/chatmail/rpc-client-go/v2
```

### Installing deltachat-rpc-server

This package depends on a standalone Chatmail RPC server
`deltachat-rpc-server` program that must be available in your
`PATH`. For installation instructions check:
https://github.com/chatmail/core/tree/main/deltachat-rpc-server

## Developing bots faster ⚡

If you want to develop bots, you should probably use this
library together with [deltabot-cli-go][deltabotcli], it takes
away the repetitive process of creating the bot CLI and let you
focus on writing your message processing logic.

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

	"github.com/chatmail/rpc-client-go/v2/deltachat"
)

func logEvent(bot *deltachat.Bot, accId uint32, event deltachat.EventType) {
	switch ev := event.(type) {
	case *deltachat.EventTypeInfo:
		log.Printf("INFO: %v", ev.Msg)
	case *deltachat.EventTypeWarning:
		log.Printf("WARNING: %v", ev.Msg)
	case *deltachat.EventTypeError:
		log.Printf("ERROR: %v", ev.Msg)
	}
}

func runEchoBot(bot *deltachat.Bot, accId uint32) {
	sysinfo, _ := bot.Rpc.GetSystemInfo()
	log.Println("Running deltachat core", sysinfo["deltachat_core_version"])

	bot.On(&deltachat.EventTypeInfo{}, logEvent)
	bot.On(&deltachat.EventTypeWarning{}, logEvent)
	bot.On(&deltachat.EventTypeError{}, logEvent)
	bot.OnNewMsg(func(bot *deltachat.Bot, accId uint32, msgId uint32) {
		msg, _ := bot.Rpc.GetMessage(accId, msgId)
		if msg.FromId > deltachat.ContactLastSpecial {
			reply := deltachat.MessageData{Text: &msg.Text}
			if _, err := bot.Rpc.SendMsg(accId, msg.ChatId, reply); err != nil {
				log.Printf("ERROR: %v", err)
			}
		}
	})

	if isConf, _ := bot.Rpc.IsConfigured(accId); !isConf {
		log.Println("Bot not configured, configuring...")
		botFlag := "1"
		if err := bot.Rpc.SetConfig(accId, "bot", &botFlag); err != nil {
			log.Fatalln(err)
		}
		if err := bot.Rpc.AddTransportFromQr(accId, os.Args[1]); err != nil {
			log.Fatalln(err)
		}
	}

	inviteLink, _ := bot.Rpc.GetChatSecurejoinQrCode(accId, nil)
	log.Println("Listening at:", inviteLink)
	if err := bot.Run(); err != nil {
		log.Fatalln(err)
	}
}

// Get the first available account or create a new one if none exists.
func getAccount(rpc *deltachat.Rpc) uint32 {
	accounts, _ := rpc.GetAllAccountIds()
	var accId uint32
	if len(accounts) == 0 {
		accId, _ = rpc.AddAccount()
	} else {
		accId = accounts[0]
	}
	return accId
}

func main() {
	trans := deltachat.NewIOTransport()
	if err := trans.Open(); err != nil {
		log.Fatalln(err)
	}
	defer trans.Close()
	rpc := &deltachat.Rpc{Context: context.Background(), Transport: trans}
	runEchoBot(deltachat.NewBot(rpc), getAccount(rpc))
}
```
<!-- MARKDOWN-AUTO-DOCS:END -->

Save the previous code snippet as `echobot.go` then run:

```sh
go mod init echobot; go mod tidy
go run ./echobot.go dcaccount:nine.testrun.org
```

Check the [examples folder](./examples)
for more examples.

## Testing your code

`deltachat.AcFactory` is provided to help users of this library
to unit-test their code.

### Using AcFactory

Create a file called `main_test.go` inside your tests folder,
and save it with the following content:

<!-- MARKDOWN-AUTO-DOCS:START (CODE:src=./examples/echobot_full/main_test.go) -->
<!-- The below code snippet is automatically added from ./examples/echobot_full/main_test.go -->
```go
package main // replace with your package name

import (
	"testing"

	"github.com/chatmail/rpc-client-go/v2/deltachat"
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

	"github.com/chatmail/rpc-client-go/v2/deltachat"
	"github.com/stretchr/testify/assert"
)

func TestEchoBot(t *testing.T) {
	acfactory.WithOnlineBot(func(bot *deltachat.Bot, botAcc uint32) {
		go runEchoBot(bot, botAcc) // this is the function we are testing
		acfactory.WithOnlineAccount(func(uRpc *deltachat.Rpc, uAccId uint32) {
			chatId := acfactory.CreateChat(uRpc, uAccId, bot.Rpc, botAcc)
			_, _ = uRpc.MiscSendTextMessage(uAccId, chatId, "hi")
			msg := acfactory.NextMsg(uRpc, uAccId)
			assert.Equal(t, "hi", msg.Text) // check that bot echoes back the "hi" message from user
		})
	})
}
```
<!-- MARKDOWN-AUTO-DOCS:END -->

Check the complete example at [examples/echobot_full](./examples/echobot_full)

## Contributing

Pull requests are welcome! check [CONTRIBUTING.md](./CONTRIBUTING.md)

[deltabotcli]: https://github.com/deltachat-bot/deltabot-cli-go/
