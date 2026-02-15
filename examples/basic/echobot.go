package main

import (
	"context"
	"log"
	"os"

	"github.com/chatmail/rpc-client-go/deltachat"
	"github.com/chatmail/rpc-client-go/deltachat/option"
	"github.com/chatmail/rpc-client-go/deltachat/transport"
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

func main() {
	trans := transport.NewIOTransport()
	trans.Open()
	defer trans.Close()
	rpc := &deltachat.Rpc{Context: context.Background(), Transport: trans}

	sysinfo, _ := rpc.GetSystemInfo()
	log.Println("Running deltachat core", sysinfo["deltachat_core_version"])

	bot := deltachat.NewBot(rpc)
	accId := deltachat.GetAccount(rpc)

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
		rpc.SetConfigFromQr(accId, os.Args[1])
		err := bot.Rpc.Configure(accId)
		if err != nil {
			log.Fatalln(err)
		}
	}

	inviteLink, _ := bot.Rpc.GetChatSecurejoinQrCode(accId, option.None[deltachat.ChatId]())
	log.Println("Listening at:", inviteLink)
	bot.Run()
}
