package main

import (
	"context"
	"log"
	"os"

	"github.com/chatmail/rpc-client-go/deltachat"
	"github.com/chatmail/rpc-client-go/deltachat/option"
	"github.com/chatmail/rpc-client-go/deltachat/transport"
)

// Get the first available account or create a new one if none exists.
// A real client would need to provide account selection.
func getAccount(rpc *deltachat.Rpc) deltachat.AccountId {
	accounts, _ := rpc.GetAllAccountIds()
	var accId deltachat.AccountId
	if len(accounts) == 0 {
		accId, _ = rpc.AddAccount()
	} else {
		accId = accounts[0]
	}
	return accId
}

// Dummy function that just prints some events, here your client's UI would process the event
func handleEvent(rpc *deltachat.Rpc, accId deltachat.AccountId, event deltachat.Event) {
	switch ev := event.(type) {
	case deltachat.EventInfo:
		log.Println("INFO:", ev.Msg)
	case deltachat.EventWarning:
		log.Println("WARNING:", ev.Msg)
	case deltachat.EventError:
		log.Println("ERROR:", ev.Msg)
	case deltachat.EventIncomingMsg:
		snapshot, _ := rpc.GetMessage(accId, ev.MsgId)
		log.Printf("Got new message from %v: %v", snapshot.Sender.DisplayName, snapshot.Text)
	}
}

func main() {
	trans := transport.NewIOTransport()
	trans.Stderr = nil // disable printing logs from core RPC, do this if your client is a TUI
	trans.Open()       // start communication with Delta Chat core
	defer trans.Close()

	rpc := &deltachat.Rpc{Context: context.Background(), Transport: trans}
	accId := getAccount(rpc)

	if configured, _ := rpc.IsConfigured(accId); configured {
		rpc.StartIo(accId)
	} else {
		log.Println("Account not configured, configuring...")
		rpc.SetConfigFromQr(accId, os.Args[1])
		if err := rpc.Configure(accId); err != nil {
			log.Fatalln(err)
		}
	}

	addr, _ := rpc.GetConfig(accId, "addr")
	log.Println("Using account:", addr.Unwrap())
	inviteLink, _ := rpc.GetChatSecurejoinQrCode(accId, option.None[deltachat.ChatId]())
	log.Println("Listening on:", inviteLink)

	for {
		accId2, event, err := rpc.GetNextEvent()
		if err != nil {
			break
		}
		handleEvent(rpc, accId2, event)
	}
}
