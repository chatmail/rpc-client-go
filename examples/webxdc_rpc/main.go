package main

import (
	"context"
	"log"
	"os"

	"github.com/chatmail/rpc-client-go/deltachat"
	"github.com/chatmail/rpc-client-go/deltachat/transport"
	"github.com/chatmail/rpc-client-go/deltachat/xdcrpc"
)

func onEvent(bot *deltachat.Bot, accId deltachat.AccountId, event deltachat.Event) {
	switch ev := event.(type) {
	case deltachat.EventWebxdcStatusUpdate:
		xdcrpc.HandleMessage(bot.Rpc, accId, ev.MsgId, ev.StatusUpdateSerial, &API{})
	}
}

func main() {
	trans := transport.NewIOTransport()
	trans.Open()
	defer trans.Close()
	rpc := &deltachat.Rpc{Context: context.Background(), Transport: trans}

	bot := deltachat.NewBot(rpc)
	accId := deltachat.GetAccount(rpc)

	bot.OnUnhandledEvent(onEvent)
	bot.OnNewMsg(func(bot *deltachat.Bot, accId deltachat.AccountId, msgId deltachat.MsgId) {
		msg, _ := bot.Rpc.GetMessage(accId, msgId)
		if msg.FromId > deltachat.ContactLastSpecial {
			// TODO: send the webxdc app here
			return
		}
	})

	if isConf, _ := bot.Rpc.IsConfigured(accId); !isConf {
		log.Println("Bot not configured, configuring...")
		err := bot.Configure(accId, os.Args[1], os.Args[2])
		if err != nil {
			log.Fatalln(err)
		}
	}

	bot.Run()
}
