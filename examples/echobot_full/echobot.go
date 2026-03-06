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
