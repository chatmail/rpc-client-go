package deltachat

import (
	"context"
	"fmt"
	"sync"

	"github.com/chatmail/rpc-client-go/deltachat/option"
)

type EventHandler func(bot *Bot, accId AccountId, event Event)
type NewMsgHandler func(bot *Bot, accId AccountId, msgId MsgId)

// BotRunningErr is returned by Bot.Run() if the Bot is already running
type BotRunningErr struct{}

func (self *BotRunningErr) Error() string {
	return "bot is already running"
}

// Delta Chat bot that listen to account events, multiple accounts supported.
type Bot struct {
	Rpc              *Rpc
	newMsgHandler    NewMsgHandler
	onUnhandledEvent EventHandler
	handlerMap       map[eventType]EventHandler
	handlerMapMutex  sync.RWMutex
	ctxMutex         sync.Mutex
	ctx              context.Context
	stop             context.CancelFunc
}

// Create a new Bot that will process events for all created accounts.
func NewBot(rpc *Rpc) *Bot {
	return &Bot{Rpc: rpc, handlerMap: make(map[eventType]EventHandler)}
}

// Set an EventHandler for the given event type. Calling On() several times
// with the same event type will override the previously set EventHandler.
func (self *Bot) On(event Event, handler EventHandler) {
	self.handlerMapMutex.Lock()
	self.handlerMap[event.eventType()] = handler
	self.handlerMapMutex.Unlock()
}

// Set an EventHandler to handle events whithout an EventHandler set via On().
// Calling OnUnhandledEvent() several times will override the previously set EventHandler.
func (self *Bot) OnUnhandledEvent(handler EventHandler) {
	self.onUnhandledEvent = handler
}

// Remove EventHandler for the given event type.
func (self *Bot) RemoveEventHandler(event Event) {
	self.handlerMapMutex.Lock()
	delete(self.handlerMap, event.eventType())
	self.handlerMapMutex.Unlock()
}

// Set the NewMsgHandler for this bot.
func (self *Bot) OnNewMsg(handler NewMsgHandler) {
	self.newMsgHandler = handler
}

// Configure one of the bot's accounts.
func (self *Bot) Configure(accId AccountId, addr string, password string) error {
	err := self.Rpc.BatchSetConfig(
		accId,
		map[string]option.Option[string]{
			"bot":     option.Some("1"),
			"addr":    option.Some(addr),
			"mail_pw": option.Some(password),
		},
	)
	if err != nil {
		return err
	}
	return self.Rpc.Configure(accId)
}

// Set UI-specific configuration value in the given account.
// This is useful for custom 3rd party settings set by bot programs.
func (self *Bot) SetUiConfig(accId AccountId, key string, value option.Option[string]) error {
	return self.Rpc.SetConfig(accId, "ui."+key, value)
}

// Get custom UI-specific configuration value set with SetUiConfig().
func (self *Bot) GetUiConfig(accId AccountId, key string) (option.Option[string], error) {
	return self.Rpc.GetConfig(accId, "ui."+key)
}

// Process events until Stop() is called. If the bot is already running, BotRunningErr is returned.
func (self *Bot) Run() error {
	self.ctxMutex.Lock()
	if self.ctx != nil && self.ctx.Err() == nil {
		self.ctxMutex.Unlock()
		return &BotRunningErr{}
	}
	self.ctx, self.stop = context.WithCancel(context.Background())
	self.ctxMutex.Unlock()

	self.Rpc.StartIoForAllAccounts() //nolint:errcheck
	ids, _ := self.Rpc.GetAllAccountIds()
	for _, accId := range ids {
		if isConf, _ := self.Rpc.IsConfigured(accId); isConf {
			self.processMessages(accId) // Process old messages.
		}
	}

	eventChan := make(chan struct {
		AccountId
		Event
	})
	go func() {
		for {
			rpc := &Rpc{Context: self.ctx, Transport: self.Rpc.Transport}
			accId, event, err := rpc.GetNextEvent()
			if err != nil {
				close(eventChan)
				break
			}
			eventChan <- struct {
				AccountId
				Event
			}{accId, event}
		}
	}()

	for {
		evData, ok := <-eventChan
		if !ok {
			self.Stop()
			return nil
		}
		self.onEvent(evData.AccountId, evData.Event)
		if evData.Event.eventType() == eventTypeIncomingMsg {
			self.processMessages(evData.AccountId)
		}
	}
}

// Return true if bot is running (Bot.Run() is running) or false otherwise.
func (self *Bot) IsRunning() bool {
	self.ctxMutex.Lock()
	defer self.ctxMutex.Unlock()
	return self.ctx != nil && self.ctx.Err() == nil
}

// Stop processing events.
func (self *Bot) Stop() {
	self.ctxMutex.Lock()
	defer self.ctxMutex.Unlock()
	if self.ctx != nil && self.ctx.Err() == nil {
		self.stop()
	}
}

func (self *Bot) onEvent(accId AccountId, event Event) {
	self.handlerMapMutex.RLock()
	handler, ok := self.handlerMap[event.eventType()]
	self.handlerMapMutex.RUnlock()
	if ok {
		handler(self, accId, event)
	} else if self.onUnhandledEvent != nil {
		self.onUnhandledEvent(self, accId, event)
	}
}

func (self *Bot) processMessages(accId AccountId) {
	msgIds, err := self.Rpc.GetNextMsgs(accId)
	if err != nil {
		return
	}
	for _, msgId := range msgIds {
		self.Rpc.SetConfig(accId, "last_msg_id", option.Some(fmt.Sprintf("%v", msgId))) //nolint:errcheck
		if self.newMsgHandler != nil {
			self.newMsgHandler(self, accId, msgId)
		}
	}
}
