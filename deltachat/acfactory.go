package deltachat

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/chatmail/rpc-client-go/deltachat/option"
	"github.com/chatmail/rpc-client-go/deltachat/transport"
)

// AcFactory facilitates unit testing Delta Chat clients/bots.
// It must be used in conjunction with a test mail server service, for example:
// https://github.com/deltachat/mail-server-tester
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
	// DefaultCfg is the default settings to apply to new created accounts
	DefaultCfg  map[string]option.Option[string]
	Debug       bool
	tempDir     string
	serial      int64
	startTime   int64
	serialMutex sync.Mutex
	tearUp      bool
}

// Prepare the AcFactory.
//
// If the test mail server has not standard configuration, you should set the custom configuration
// here.
func (factory *AcFactory) TearUp() {
	if factory.DefaultCfg == nil {
		factory.DefaultCfg = map[string]option.Option[string]{
			"mail_server":   option.Some("localhost"),
			"send_server":   option.Some("localhost"),
			"mail_port":     option.Some("3143"),
			"send_port":     option.Some("3025"),
			"mail_security": option.Some("3"),
			"send_security": option.Some("3"),
			"mvbox_move":    option.Some("0"),
		}

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
	trans := transport.NewIOTransport()
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
func (factory *AcFactory) WithUnconfiguredAccount(callback func(*Rpc, AccountId)) {
	factory.WithRpc(func(rpc *Rpc) {
		accId, err := rpc.AddAccount()
		if err != nil {
			panic(err)
		}
		factory.serialMutex.Lock()
		factory.serial++
		serial := factory.serial
		factory.serialMutex.Unlock()

		if len(factory.DefaultCfg) != 0 {
			err = rpc.BatchSetConfig(accId, factory.DefaultCfg)
			if err != nil {
				panic(err)
			}
		}
		err = rpc.BatchSetConfig(accId, map[string]option.Option[string]{
			"addr":    option.Some(fmt.Sprintf("acc%v.%v@localhost", serial, factory.startTime)),
			"mail_pw": option.Some(fmt.Sprintf("password%v", serial)),
		})
		if err != nil {
			panic(err)
		}

		callback(rpc, accId)
	})
}

// Get a new account configured and with I/O already started.
func (factory *AcFactory) WithOnlineAccount(callback func(*Rpc, AccountId)) {
	factory.WithUnconfiguredAccount(func(rpc *Rpc, accId AccountId) {
		err := rpc.Configure(accId)
		if err != nil {
			panic(err)
		}

		callback(rpc, accId)
	})
}

// Get a new bot not yet configured, but its account is ready to be configured.
func (factory *AcFactory) WithUnconfiguredBot(callback func(*Bot, AccountId)) {
	factory.WithUnconfiguredAccount(func(rpc *Rpc, accId AccountId) {
		bot := NewBot(rpc)
		callback(bot, accId)
	})
}

// Get a new bot configured and with its account I/O already started. The bot is not running yet.
func (factory *AcFactory) WithOnlineBot(callback func(*Bot, AccountId)) {
	factory.WithUnconfiguredAccount(func(rpc *Rpc, accId AccountId) {
		addr, _ := rpc.GetConfig(accId, "addr")
		pass, _ := rpc.GetConfig(accId, "mail_pw")
		bot := NewBot(rpc)
		err := bot.Configure(accId, addr.Unwrap(), pass.Unwrap())
		if err != nil {
			panic(err)
		}

		callback(bot, accId)
	})
}

// Get a new bot configured and already listening to new events/messages.
// It is ensured that Bot.IsRunning() is true for the returned bot.
func (factory *AcFactory) WithRunningBot(callback func(*Bot, AccountId)) {
	factory.WithOnlineBot(func(bot *Bot, accId AccountId) {
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
func (factory *AcFactory) NextMsg(rpc *Rpc, accId AccountId) *MsgSnapshot {
	event := factory.WaitForEvent(rpc, accId, EventIncomingMsg{}).(EventIncomingMsg)
	msg, err := rpc.GetMessage(accId, event.MsgId)
	if err != nil {
		panic(err)
	}
	return msg
}

// Introduce two accounts to each other creating a 1:1 chat between them.
func (factory *AcFactory) IntroduceEachOther(rpc1 *Rpc, accId1 AccountId, rpc2 *Rpc, accId2 AccountId) {
	qrdata, err := rpc1.GetChatSecurejoinQrCode(accId1, option.None[ChatId]())
	if err != nil {
		panic(err)
	}
	_, err = rpc2.SecureJoin(accId2, qrdata)
	if err != nil {
		panic(err)
	}

	for {
		event := factory.WaitForEvent(rpc1, accId1, EventSecurejoinInviterProgress{}).(EventSecurejoinInviterProgress)
		if event.Progress == 1000 {
			break
		}
	}

	for {
		event := factory.WaitForEvent(rpc2, accId2, EventSecurejoinJoinerProgress{}).(EventSecurejoinJoinerProgress)
		if event.Progress == 1000 {
			break
		}
	}
}

// Create a 1:1 chat with accId2 in the chatlist of accId1.
func (factory *AcFactory) CreateChat(rpc1 *Rpc, accId1 AccountId, rpc2 *Rpc, accId2 AccountId) ChatId {
	vcard, err := rpc2.makeVcard(accId2, []ContactId{ContactSelf})
	if err != nil {
		panic(err)
	}
	ids, err := rpc1.importVcardContents(accId1, vcard)
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
func (factory *AcFactory) TestImage() string {
	var img string
	factory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
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
// with the given ChatId.
func (factory *AcFactory) WaitForEventInChat(rpc *Rpc, accId AccountId, chatId ChatId, event Event) Event {
	for {
		event = factory.WaitForEvent(rpc, accId, event)
		if getChatId(event) == chatId {
			return event
		}
	}
}

// Wait for an event of the same type as the given event.
func (factory *AcFactory) WaitForEvent(rpc *Rpc, accId AccountId, event Event) Event {
	for {
		accId2, ev, err := rpc.GetNextEvent()
		if err != nil {
			panic(err)
		}
		if accId != accId2 {
			fmt.Printf("WARNING: Waiting for event in account %v, but got event for account %v, discarding event %#v.\n", accId, accId2, event)
			continue
		}
		if ev.eventType() == event.eventType() {
			if factory.Debug {
				fmt.Printf("Got awaited event %v\n", ev.eventType())
			}
			return ev
		}
		if factory.Debug {
			fmt.Printf("Waiting for event %v, got: %v\n", event.eventType(), ev.eventType())
		}
	}
}

func (factory *AcFactory) ensureTearUp() {
	if !factory.tearUp {
		panic("TearUp() required")
	}
}

func getChatId(event Event) ChatId {
	var chatId ChatId
	switch ev := event.(type) {
	case EventMsgsChanged:
		chatId = ev.ChatId
	case EventReactionsChanged:
		chatId = ev.ChatId
	case EventIncomingMsg:
		chatId = ev.ChatId
	case EventMsgsNoticed:
		chatId = ev.ChatId
	case EventMsgDelivered:
		chatId = ev.ChatId
	case EventMsgFailed:
		chatId = ev.ChatId
	case EventMsgRead:
		chatId = ev.ChatId
	case EventMsgDeleted:
		chatId = ev.ChatId
	case EventChatModified:
		chatId = ev.ChatId
	case EventChatEphemeralTimerModified:
		chatId = ev.ChatId
	}
	return chatId
}
