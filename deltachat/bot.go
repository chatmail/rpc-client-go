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

func (error *BotRunningErr) Error() string {
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
func (bot *Bot) On(event Event, handler EventHandler) {
	bot.handlerMapMutex.Lock()
	bot.handlerMap[event.eventType()] = handler
	bot.handlerMapMutex.Unlock()
}

// Set an EventHandler to handle events whithout an EventHandler set via On().
// Calling OnUnhandledEvent() several times will override the previously set EventHandler.
func (bot *Bot) OnUnhandledEvent(handler EventHandler) {
	bot.onUnhandledEvent = handler
}

// Remove EventHandler for the given event type.
func (bot *Bot) RemoveEventHandler(event Event) {
	bot.handlerMapMutex.Lock()
	delete(bot.handlerMap, event.eventType())
	bot.handlerMapMutex.Unlock()
}

// Set the NewMsgHandler for this bot.
func (bot *Bot) OnNewMsg(handler NewMsgHandler) {
	bot.newMsgHandler = handler
}

// Configure one of the bot's accounts.
func (bot *Bot) Configure(accId AccountId, addr string, password string) error {
	err := bot.Rpc.BatchSetConfig(
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
	return bot.Rpc.Configure(accId)
}

// Set UI-specific configuration value in the given account.
// This is useful for custom 3rd party settings set by bot programs.
func (bot *Bot) SetUiConfig(accId AccountId, key string, value option.Option[string]) error {
	return bot.Rpc.SetConfig(accId, "ui."+key, value)
}

// Get custom UI-specific configuration value set with SetUiConfig().
func (bot *Bot) GetUiConfig(accId AccountId, key string) (option.Option[string], error) {
	return bot.Rpc.GetConfig(accId, "ui."+key)
}

// Process events until Stop() is called. If the bot is already running, BotRunningErr is returned.
func (bot *Bot) Run() error {
	bot.ctxMutex.Lock()
	if bot.ctx != nil && bot.ctx.Err() == nil {
		bot.ctxMutex.Unlock()
		return &BotRunningErr{}
	}
	bot.ctx, bot.stop = context.WithCancel(context.Background())
	bot.ctxMutex.Unlock()

	bot.Rpc.StartIoForAllAccounts() //nolint:errcheck
	ids, _ := bot.Rpc.GetAllAccountIds()
	for _, accId := range ids {
		if isConf, _ := bot.Rpc.IsConfigured(accId); isConf {
			bot.processMessages(accId) // Process old messages.
		}
	}

	eventChan := make(chan struct {
		AccountId AccountId
		Event     Event
	})
	go func() {
		for {
			rpc := &Rpc{Context: bot.ctx, Transport: bot.Rpc.Transport}
			accId, event, err := rpc.GetNextEvent()
			if err != nil {
				close(eventChan)
				break
			}
			eventChan <- struct {
				AccountId AccountId
				Event     Event
			}{accId, event}
		}
	}()

	for {
		evData, ok := <-eventChan
		if !ok {
			bot.Stop()
			return nil
		}
		bot.onEvent(evData.AccountId, evData.Event)
		if evData.Event.eventType() == eventTypeIncomingMsg {
			bot.processMessages(evData.AccountId)
		}
	}
}

// Return true if bot is running (Bot.Run() is running) or false otherwise.
func (bot *Bot) IsRunning() bool {
	bot.ctxMutex.Lock()
	defer bot.ctxMutex.Unlock()
	return bot.ctx != nil && bot.ctx.Err() == nil
}

// Stop processing events.
func (bot *Bot) Stop() {
	bot.ctxMutex.Lock()
	defer bot.ctxMutex.Unlock()
	if bot.ctx != nil && bot.ctx.Err() == nil {
		bot.stop()
	}
}

func (bot *Bot) onEvent(accId AccountId, event Event) {
	bot.handlerMapMutex.RLock()
	handler, ok := bot.handlerMap[event.eventType()]
	bot.handlerMapMutex.RUnlock()
	if ok {
		handler(bot, accId, event)
	} else if bot.onUnhandledEvent != nil {
		bot.onUnhandledEvent(bot, accId, event)
	}
}

func (bot *Bot) processMessages(accId AccountId) {
	msgIds, err := bot.Rpc.GetNextMsgs(accId)
	if err != nil {
		return
	}
	for _, msgId := range msgIds {
		bot.Rpc.SetConfig(accId, "last_msg_id", option.Some(fmt.Sprintf("%v", msgId))) //nolint:errcheck
		if bot.newMsgHandler != nil {
			bot.newMsgHandler(bot, accId, msgId)
		}
	}
}
