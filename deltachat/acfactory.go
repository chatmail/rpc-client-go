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
func (self *AcFactory) TearUp() {
	if self.DefaultCfg == nil {
		self.DefaultCfg = map[string]option.Option[string]{
			"mail_server":   option.Some("localhost"),
			"send_server":   option.Some("localhost"),
			"mail_port":     option.Some("3143"),
			"send_port":     option.Some("3025"),
			"mail_security": option.Some("3"),
			"send_security": option.Some("3"),
			"mvbox_move":    option.Some("0"),
		}

	}
	self.startTime = time.Now().Unix()

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}
	self.tempDir = dir

	self.tearUp = true
}

// Do cleanup, removing temporary directories and files created by the configured test accounts.
// Usually TearDown() is called with defer immediately after the creation of the AcFactory instance.
func (self *AcFactory) TearDown() {
	self.ensureTearUp()
	os.RemoveAll(self.tempDir)
}

// MkdirTemp creates a new temporary directory. The directory is automatically removed on TearDown().
func (self *AcFactory) MkdirTemp() string {
	dir, err := os.MkdirTemp(self.tempDir, "")
	if err != nil {
		panic(err)
	}
	return dir
}

// Call the given function passing a new Rpc as parameter.
func (self *AcFactory) WithRpc(callback func(*Rpc)) {
	self.ensureTearUp()
	trans := transport.NewIOTransport()
	if !self.Debug {
		trans.Stderr = nil
	}
	dir := self.MkdirTemp()
	trans.AccountsDir = filepath.Join(dir, "accounts")
	err := trans.Open()
	if err != nil {
		panic(err)
	}
	defer trans.Close()

	callback(&Rpc{Context: context.Background(), Transport: trans})
}

// Get a new Account that is not yet configured, but it is ready to be configured.
func (self *AcFactory) WithUnconfiguredAccount(callback func(*Rpc, AccountId)) {
	self.WithRpc(func(rpc *Rpc) {
		accId, err := rpc.AddAccount()
		if err != nil {
			panic(err)
		}
		self.serialMutex.Lock()
		self.serial++
		serial := self.serial
		self.serialMutex.Unlock()

		if len(self.DefaultCfg) != 0 {
			err = rpc.BatchSetConfig(accId, self.DefaultCfg)
			if err != nil {
				panic(err)
			}
		}
		err = rpc.BatchSetConfig(accId, map[string]option.Option[string]{
			"addr":    option.Some(fmt.Sprintf("acc%v.%v@localhost", serial, self.startTime)),
			"mail_pw": option.Some(fmt.Sprintf("password%v", serial)),
		})
		if err != nil {
			panic(err)
		}

		callback(rpc, accId)
	})
}

// Get a new account configured and with I/O already started.
func (self *AcFactory) WithOnlineAccount(callback func(*Rpc, AccountId)) {
	self.WithUnconfiguredAccount(func(rpc *Rpc, accId AccountId) {
		err := rpc.Configure(accId)
		if err != nil {
			panic(err)
		}

		callback(rpc, accId)
	})
}

// Get a new bot not yet configured, but its account is ready to be configured.
func (self *AcFactory) WithUnconfiguredBot(callback func(*Bot, AccountId)) {
	self.WithUnconfiguredAccount(func(rpc *Rpc, accId AccountId) {
		bot := NewBot(rpc)
		callback(bot, accId)
	})
}

// Get a new bot configured and with its account I/O already started. The bot is not running yet.
func (self *AcFactory) WithOnlineBot(callback func(*Bot, AccountId)) {
	self.WithUnconfiguredAccount(func(rpc *Rpc, accId AccountId) {
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
func (self *AcFactory) WithRunningBot(callback func(*Bot, AccountId)) {
	self.WithOnlineBot(func(bot *Bot, accId AccountId) {
		var err error
		go func() { err = bot.Run() }()
		for {
			if bot.IsRunning() {
				break
			}
			if err != nil {
				panic(err)
			}
		}

		callback(bot, accId)
	})
}

// Wait for the next incoming message in the given account.
func (self *AcFactory) NextMsg(rpc *Rpc, accId AccountId) *MsgSnapshot {
	event := self.WaitForEvent(rpc, accId, EventIncomingMsg{}).(EventIncomingMsg)
	msg, err := rpc.GetMessage(accId, event.MsgId)
	if err != nil {
		panic(err)
	}
	return msg
}

// Introduce two accounts to each other creating a 1:1 chat between them and exchanging messages.
func (self *AcFactory) IntroduceEachOther(rpc1 *Rpc, accId1 AccountId, rpc2 *Rpc, accId2 AccountId) {
	chatId := self.CreateChat(rpc1, accId1, rpc2, accId2)
	_, err := rpc1.MiscSendTextMessage(accId1, chatId, "hi")
	if err != nil {
		panic(err)
	}
	self.WaitForEventInChat(rpc1, accId1, chatId, EventMsgsChanged{})
	snapshot := self.NextMsg(rpc2, accId2)
	if snapshot.Text != "hi" {
		panic("unexpected message: " + snapshot.Text)
	}

	err = rpc2.AcceptChat(accId2, snapshot.ChatId)
	if err != nil {
		panic(err)
	}
	_, err = rpc2.MiscSendTextMessage(accId2, snapshot.ChatId, "hello")
	if err != nil {
		panic(err)
	}
	self.WaitForEventInChat(rpc2, accId2, snapshot.ChatId, EventMsgsChanged{})
	snapshot = self.NextMsg(rpc1, accId1)
	if snapshot.Text != "hello" {
		panic("unexpected message: " + snapshot.Text)
	}
}

// Create a 1:1 chat with accId2 in the chatlist of accId1.
func (self *AcFactory) CreateChat(rpc1 *Rpc, accId1 AccountId, rpc2 *Rpc, accId2 AccountId) ChatId {
	addr2, err := rpc2.GetConfig(accId2, "configured_addr")
	if err != nil {
		panic(err)
	}

	contactId, err := rpc1.CreateContact(accId1, addr2.Unwrap(), "")
	if err != nil {
		panic(err)
	}

	chatId, err := rpc1.CreateChatByContactId(accId1, contactId)
	if err != nil {
		panic(err)
	}

	return chatId
}

// Get a path to an image file that can be used for testing.
func (self *AcFactory) TestImage() string {
	var img string
	self.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
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
func (self *AcFactory) TestWebxdc() string {
	self.ensureTearUp()
	dir := self.MkdirTemp()
	path := filepath.Join(dir, "test.xdc")
	zipFile, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer zipFile.Close()
	writer := zip.NewWriter(zipFile)
	defer writer.Close()

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
func (self *AcFactory) WaitForEventInChat(rpc *Rpc, accId AccountId, chatId ChatId, event Event) Event {
	for {
		event = self.WaitForEvent(rpc, accId, event)
		if getChatId(event) == chatId {
			return event
		}
	}
}

// Wait for an event of the same type as the given event.
func (self *AcFactory) WaitForEvent(rpc *Rpc, accId AccountId, event Event) Event {
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
			if self.Debug {
				fmt.Printf("Got awaited event %v\n", ev.eventType())
			}
			return ev
		}
		if self.Debug {
			fmt.Printf("Waiting for event %v, got: %v\n", event.eventType(), ev.eventType())
		}
	}
}

func (self *AcFactory) ensureTearUp() {
	if !self.tearUp {
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
