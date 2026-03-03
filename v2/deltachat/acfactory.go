package deltachat

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AcFactory facilitates unit testing Delta Chat clients/bots.
//
// Typical usage is as follows:
//
//	import (
//		"testing"
//		"github.com/chatmail/rpc-client-go/deltachat"
//	)

//  var acfactory *deltachat.AcFactory

//	func TestMain(m *testing.M) {
//		acfactory = &deltachat.AcFactory{}
//		acfactory.TearUp()
//		defer acfactory.TearDown()
//		m.Run()
//	}
type AcFactory struct {
	// ConfigQr is the DCACCOUNT: URI used to create new accounts
	ConfigQr  string
	Debug     bool
	tempDir   string
	startTime int64
	tearUp    bool
}

// Prepare the AcFactory.
func (factory *AcFactory) TearUp() {
	if factory.ConfigQr == "" {
		factory.ConfigQr = "dcaccount:nine.testrun.org"
	}
	factory.startTime = time.Now().Unix()

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}
	factory.tempDir = dir

	factory.tearUp = true
}

// Do cleanup, removing temporary directories and files created by the configured test accounts.
// Usually TearDown() is called with defer immediately after the creation of the AcFactory instance.
func (factory *AcFactory) TearDown() {
	factory.ensureTearUp()
	if err := os.RemoveAll(factory.tempDir); err != nil {
		panic(err)
	}
}

// MkdirTemp creates a new temporary directory. The directory is automatically removed on TearDown().
func (factory *AcFactory) MkdirTemp() string {
	dir, err := os.MkdirTemp(factory.tempDir, "")
	if err != nil {
		panic(err)
	}
	return dir
}

// Call the given function passing a new Rpc as parameter.
func (factory *AcFactory) WithRpc(callback func(*Rpc)) {
	factory.ensureTearUp()
	trans := NewIOTransport()
	if !factory.Debug {
		trans.Stderr = nil
	}
	dir := factory.MkdirTemp()
	trans.AccountsDir = filepath.Join(dir, "accounts")
	err := trans.Open()
	if err != nil {
		panic(err)
	}
	defer trans.Close()

	callback(&Rpc{Context: context.Background(), Transport: trans})
}

// Get a new Account that is not yet configured, but it is ready to be configured.
func (factory *AcFactory) WithUnconfiguredAccount(callback func(*Rpc, uint32)) {
	factory.WithRpc(func(rpc *Rpc) {
		accId, err := rpc.AddAccount()
		if err != nil {
			panic(err)
		}
		callback(rpc, accId)
	})
}

// Get a new account configured and with I/O already started.
func (factory *AcFactory) WithOnlineAccount(callback func(*Rpc, uint32)) {
	factory.WithUnconfiguredAccount(func(rpc *Rpc, accId uint32) {
		if err := rpc.AddTransportFromQr(accId, factory.ConfigQr); err != nil {
			panic(err)
		}
		callback(rpc, accId)
	})
}

// Get a new bot not yet configured, but its account is ready to be configured.
func (factory *AcFactory) WithUnconfiguredBot(callback func(*Bot, uint32)) {
	factory.WithUnconfiguredAccount(func(rpc *Rpc, accId uint32) {
		botFlag := "1"
		if err := rpc.SetConfig(accId, "bot", &botFlag); err != nil {
			panic(err)
		}
		bot := NewBot(rpc)
		callback(bot, accId)
	})
}

// Get a new bot configured and with its account I/O already started. The bot is not running yet.
func (factory *AcFactory) WithOnlineBot(callback func(*Bot, uint32)) {
	factory.WithUnconfiguredBot(func(bot *Bot, accId uint32) {
		if err := bot.Rpc.AddTransportFromQr(accId, factory.ConfigQr); err != nil {
			panic(err)
		}
		callback(bot, accId)
	})
}

// Get a new bot configured and already listening to new events/messages.
// It is ensured that Bot.IsRunning() is true for the returned bot.
func (factory *AcFactory) WithRunningBot(callback func(*Bot, uint32)) {
	factory.WithOnlineBot(func(bot *Bot, accId uint32) {
		var err error
		go func() { err = bot.Run() }()
		for !bot.IsRunning() {
			if err != nil {
				panic(err)
			}
		}

		callback(bot, accId)
	})
}

// Wait for the next incoming message in the given account.
func (factory *AcFactory) NextMsg(rpc *Rpc, accId uint32) Message {
	event := factory.WaitForEvent(rpc, accId, &EventTypeIncomingMsg{}).(*EventTypeIncomingMsg)
	msg, err := rpc.GetMessage(accId, event.MsgId)
	if err != nil {
		panic(err)
	}
	return msg
}

// Introduce two accounts to each other creating a 1:1 chat between them.
func (factory *AcFactory) IntroduceEachOther(rpc1 *Rpc, accId1 uint32, rpc2 *Rpc, accId2 uint32) {
	qrdata, err := rpc1.GetChatSecurejoinQrCode(accId1, nil)
	if err != nil {
		panic(err)
	}
	_, err = rpc2.SecureJoin(accId2, qrdata)
	if err != nil {
		panic(err)
	}

	for {
		event := factory.WaitForEvent(rpc1, accId1, &EventTypeSecurejoinInviterProgress{}).(*EventTypeSecurejoinInviterProgress)
		if event.Progress == 1000 {
			break
		}
	}

	for {
		event := factory.WaitForEvent(rpc2, accId2, &EventTypeSecurejoinJoinerProgress{}).(*EventTypeSecurejoinJoinerProgress)
		if event.Progress == 1000 {
			break
		}
	}
}

// Create a 1:1 chat with accId2 in the chatlist of accId1.
func (factory *AcFactory) CreateChat(rpc1 *Rpc, accId1 uint32, rpc2 *Rpc, accId2 uint32) uint32 {
	vcard, err := rpc2.MakeVcard(accId2, []uint32{ContactSelf})
	if err != nil {
		panic(err)
	}
	ids, err := rpc1.ImportVcardContents(accId1, vcard)
	if err != nil {
		panic(err)
	}
	chatId, err := rpc1.CreateChatByContactId(accId1, ids[0])
	if err != nil {
		panic(err)
	}

	return chatId
}

// Get a path to an image file that can be used for testing.
func (factory *AcFactory) TestImage() *string {
	var img *string
	factory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		chatId, err := rpc.CreateChatByContactId(accId, ContactSelf)
		if err != nil {
			panic(err)
		}
		chatData, err := rpc.GetBasicChatInfo(accId, chatId)
		if err != nil {
			panic(err)
		}
		img = chatData.ProfileImage
	})
	return img
}

// Get a path to a Webxdc file that can be used for testing.
func (factory *AcFactory) TestWebxdc() string {
	factory.ensureTearUp()
	dir := factory.MkdirTemp()
	path := filepath.Join(dir, "test.xdc")
	zipFile, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := zipFile.Close(); err != nil {
			panic(err)
		}
	}()

	writer := zip.NewWriter(zipFile)
	defer func() {
		if err := writer.Close(); err != nil {
			panic(err)
		}
	}()

	var files = []struct {
		Name, Body string
	}{
		{"index.html", `<html><head><script src="webxdc.js"></script></head><body>test</body></html>`},
		{"manifest.toml", `name = "TestApp"`},
	}
	for _, file := range files {
		f, err := writer.Create(file.Name)
		if err != nil {
			panic(err)
		}
		_, err = f.Write([]byte(file.Body))
		if err != nil {
			panic(err)
		}
	}

	err = writer.Close()
	if err != nil {
		panic(err)
	}

	return path
}

// Wait for an event of the same type as the given event, the event must belong to the chat
// with the given chat id.
func (factory *AcFactory) WaitForEventInChat(rpc *Rpc, accId uint32, chatId uint32, event EventType) EventType {
	for {
		event = factory.WaitForEvent(rpc, accId, event)
		if getChatId(event) == chatId {
			return event
		}
	}
}

// Wait for an event of the same type as the given event.
func (factory *AcFactory) WaitForEvent(rpc *Rpc, accId uint32, event EventType) EventType {
	for {
		ev, err := rpc.GetNextEvent()
		if err != nil {
			panic(err)
		}
		if accId != ev.ContextId {
			fmt.Printf("WARNING: Waiting for event %v in account %v, but got event for account %v, discarding event %#v.\n", event.GetKind(), accId, ev.ContextId, ev)
			continue
		}
		if ev.Event.GetKind() == event.GetKind() {
			if factory.Debug {
				fmt.Printf("Got awaited event %v\n", ev.Event.GetKind())
			}
			return ev.Event
		}
		if factory.Debug {
			fmt.Printf("Waiting for event %v, got: %v\n", event.GetKind(), ev.Event.GetKind())
		}
	}
}

func (factory *AcFactory) ensureTearUp() {
	if !factory.tearUp {
		panic("TearUp() required")
	}
}

func getChatId(event EventType) uint32 {
	var chatId uint32
	switch ev := event.(type) {
	case *EventTypeMsgsChanged:
		chatId = ev.ChatId
	case *EventTypeReactionsChanged:
		chatId = ev.ChatId
	case *EventTypeIncomingMsg:
		chatId = ev.ChatId
	case *EventTypeMsgsNoticed:
		chatId = ev.ChatId
	case *EventTypeMsgDelivered:
		chatId = ev.ChatId
	case *EventTypeMsgFailed:
		chatId = ev.ChatId
	case *EventTypeMsgRead:
		chatId = ev.ChatId
	case *EventTypeMsgDeleted:
		chatId = ev.ChatId
	case *EventTypeChatModified:
		chatId = ev.ChatId
	case *EventTypeChatEphemeralTimerModified:
		chatId = ev.ChatId
	}
	return chatId
}
