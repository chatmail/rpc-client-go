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
		err := bot.Rpc.SetConfigFromQr(accId, os.Args[1])
		if err != nil {
			log.Fatalln(err)
		}
	}

	inviteLink, _ := bot.Rpc.GetChatSecurejoinQrCode(accId, option.None[deltachat.ChatId]())
	log.Println("Listening at:", inviteLink)
	bot.Run()
}

func main() {
	trans := transport.NewIOTransport()
	trans.Open()
	defer trans.Close()
	rpc := &deltachat.Rpc{Context: context.Background(), Transport: trans}
	runEchoBot(deltachat.NewBot(rpc), deltachat.GetAccount(rpc))
}
