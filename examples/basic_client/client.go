package main

import (
	"context"
	"log"
	"os"

	"github.com/chatmail/rpc-client-go/v2/deltachat"
)

// Dummy function that just prints some events, here your client's UI would process the event
func handleEvent(rpc *deltachat.Rpc, accId uint32, event deltachat.EventType) {
	switch ev := event.(type) {
	case *deltachat.EventTypeInfo:
		log.Println("INFO:", ev.Msg)
	case *deltachat.EventTypeWarning:
		log.Println("WARNING:", ev.Msg)
	case *deltachat.EventTypeError:
		log.Println("ERROR:", ev.Msg)
	case *deltachat.EventTypeIncomingMsg:
		snapshot, _ := rpc.GetMessage(accId, ev.MsgId)
		log.Printf("Got new message from %v: %v", snapshot.Sender.DisplayName, snapshot.Text)
	}
}

func main() {
	trans := deltachat.NewIOTransport()
	trans.Stderr = nil // disable printing logs from core RPC, do this if your client is a TUI
	// start communication with Delta Chat core
	if err := trans.Open(); err != nil {
		log.Fatalln(err)
	}
	defer trans.Close()

	rpc := &deltachat.Rpc{Context: context.Background(), Transport: trans}

	accounts, _ := rpc.GetAllAccountIds()
	var accId uint32
	if len(accounts) == 0 {
		accId, _ = rpc.AddAccount()
	} else {
		accId = accounts[0]
	}

	if configured, _ := rpc.IsConfigured(accId); configured {
		if err := rpc.StartIo(accId); err != nil {
			log.Fatalln(err)
		}
	} else {
		log.Println("Account not configured, configuring...")
		if err := rpc.AddTransportFromQr(accId, os.Args[1]); err != nil {
			log.Fatalln(err)
		}
	}

	inviteLink, _ := rpc.GetChatSecurejoinQrCode(accId, nil)
	log.Println("Listening on:", inviteLink)

	for {
		event, err := rpc.GetNextEvent()
		if err != nil {
			break
		}
		handleEvent(rpc, event.ContextId, event.Event)
	}
}
