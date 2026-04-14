package deltachat

import (
	"context"
	"sync"
)

type EventHandler func(bot *Bot, accId uint32, event EventType)
type NewMsgHandler func(bot *Bot, accId uint32, msgId uint32)

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
	handlerMap       map[string]EventHandler
	handlerMapMutex  sync.RWMutex
	ctxMutex         sync.Mutex
	ctx              context.Context
	stop             context.CancelFunc
}

// Create a new Bot that will process events for all created accounts.
func NewBot(rpc *Rpc) *Bot {
	return &Bot{Rpc: rpc, handlerMap: make(map[string]EventHandler)}
}

// Set an EventHandler for the given event type. Calling On() several times
// with the same event type will override the previously set EventHandler.
func (bot *Bot) On(event EventType, handler EventHandler) {
	bot.handlerMapMutex.Lock()
	bot.handlerMap[event.GetKind()] = handler
	bot.handlerMapMutex.Unlock()
}

// Set an EventHandler to handle events whithout an EventHandler set via On().
// Calling OnUnhandledEvent() several times will override the previously set EventHandler.
func (bot *Bot) OnUnhandledEvent(handler EventHandler) {
	bot.onUnhandledEvent = handler
}

// Remove EventHandler for the given event type.
func (bot *Bot) RemoveEventHandler(event EventType) {
	bot.handlerMapMutex.Lock()
	delete(bot.handlerMap, event.GetKind())
	bot.handlerMapMutex.Unlock()
}

// Set the NewMsgHandler for this bot.
func (bot *Bot) OnNewMsg(handler NewMsgHandler) {
	bot.newMsgHandler = handler
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

	eventChan := make(chan Event)
	go func() {
		for {
			rpc := &Rpc{Context: bot.ctx, Transport: bot.Rpc.Transport}
			event, err := rpc.GetNextEvent()
			if err != nil {
				close(eventChan)
				break
			}
			eventChan <- event
		}
	}()

	for {
		evData, ok := <-eventChan
		if !ok {
			bot.Stop()
			return nil
		}
		bot.onEvent(evData.ContextId, evData.Event)
		if event, ok := evData.Event.(*EventTypeIncomingMsg); ok {
			if bot.newMsgHandler != nil {
				bot.newMsgHandler(bot, evData.ContextId, event.MsgId)
			}
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

func (bot *Bot) onEvent(accId uint32, event EventType) {
	bot.handlerMapMutex.RLock()
	handler, ok := bot.handlerMap[event.GetKind()]
	bot.handlerMapMutex.RUnlock()
	if ok {
		handler(bot, accId, event)
	} else if bot.onUnhandledEvent != nil {
		bot.onUnhandledEvent(bot, accId, event)
	}
}
