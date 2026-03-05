package deltachat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBot_NewBot(t *testing.T) {
	t.Parallel()
	acfactory.WithRpc(func(rpc *Rpc) {
		bot := NewBot(rpc)
		assert.NotNil(t, bot)
	})
}

func TestBot_BotRunningErr(t *testing.T) {
	t.Parallel()
	err := &BotRunningErr{}
	assert.NotEmpty(t, err.Error())
}

func TestBot_IsRunning(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineBot(func(bot *Bot, botAcc uint32) {
		assert.False(t, bot.IsRunning())
		done := make(chan error)
		go func() {
			done <- bot.Run()
		}()
		for !bot.IsRunning() {
			select {
			case err := <-done:
				assert.Failf(t, "bot.Run() exited before bot started running", "%v", err)
				return
			case <-time.After(10 * time.Second):
				assert.Fail(t, "timeout waiting for bot to start running")
				return
			default:
				time.Sleep(10 * time.Millisecond)
			}
		}
		assert.True(t, bot.IsRunning())
		bot.Stop()
		assert.Nil(t, <-done)
		assert.False(t, bot.IsRunning())
	})
}

func TestBot_On(t *testing.T) {
	t.Parallel()
	acfactory.WithRunningBot(func(bot *Bot, botAcc uint32) {
		acfactory.WithOnlineAccount(func(accRpc *Rpc, accId uint32) {
			incomingMsg := make(chan Message)
			bot.On(&EventTypeIncomingMsg{}, func(bot *Bot, botAcc uint32, event EventType) {
				ev := event.(*EventTypeIncomingMsg)
				snapshot, _ := bot.Rpc.GetMessage(botAcc, ev.MsgId)
				incomingMsg <- snapshot
			})

			chatWithBot := acfactory.CreateChat(accRpc, accId, bot.Rpc, botAcc)
			_, err := accRpc.MiscSendTextMessage(accId, chatWithBot, "test1")
			assert.Nil(t, err)
			msg := <-incomingMsg
			assert.Equal(t, "test1", msg.Text)
			bot.RemoveEventHandler(&EventTypeIncomingMsg{})
			close(incomingMsg)
		})
	})
}

func TestBot_OnNewMsg(t *testing.T) {
	t.Parallel()
	acfactory.WithRunningBot(func(bot *Bot, botAcc uint32) {
		acfactory.WithOnlineAccount(func(accRpc *Rpc, accId uint32) {
			bot.OnNewMsg(func(bot *Bot, botAcc uint32, msgId uint32) {
				snapshot, _ := bot.Rpc.GetMessage(botAcc, msgId)
				_, err := bot.Rpc.MiscSendTextMessage(botAcc, snapshot.ChatId, snapshot.Text)
				assert.Nil(t, err)
			})

			chatWithBot := acfactory.CreateChat(accRpc, accId, bot.Rpc, botAcc)
			_, err := accRpc.MiscSendTextMessage(accId, chatWithBot, "test2")
			assert.Nil(t, err)
			msg := acfactory.NextMsg(accRpc, accId)
			assert.Equal(t, "test2", msg.Text)
		})
	})
}

func TestBot_OnUnhandledEvent(t *testing.T) {
	t.Parallel()
	acfactory.WithRunningBot(func(bot *Bot, botAcc uint32) {
		acfactory.WithOnlineAccount(func(accRpc *Rpc, accId uint32) {
			unhandled := make(chan EventType, 10)
			bot.OnUnhandledEvent(func(bot *Bot, botAcc uint32, event EventType) {
				unhandled <- event
			})

			chatWithBot := acfactory.CreateChat(accRpc, accId, bot.Rpc, botAcc)
			_, err := accRpc.MiscSendTextMessage(accId, chatWithBot, "unhandled test")
			assert.Nil(t, err)

			// Wait for at least one unhandled event to arrive, but avoid hanging indefinitely.
			select {
			case <-unhandled:
				// received expected unhandled event
			case <-time.After(10 * time.Second):
				t.Fatalf("timeout waiting for unhandled event")
			}
		})
	})
}

func TestBot_processMessages(t *testing.T) {
	t.Parallel()
	acfactory.WithRunningBot(func(bot *Bot, botAcc uint32) {
		bot.processMessages(botAcc)
	})
}

func TestBot_Stop(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineBot(func(bot *Bot, botAcc uint32) {
		bot.On(&EventTypeInfo{}, func(bot *Bot, botAcc uint32, event EventType) { bot.Stop() })
		done := make(chan error)

		go func() {
			done <- bot.Run()
		}()
		assert.Nil(t, <-done)

		go func() {
			done <- bot.Run()
		}()
		assert.Nil(t, <-done)

		bot.On(&EventTypeInfo{}, func(bot *Bot, botAcc uint32, event EventType) { bot.Rpc.Transport.(*IOTransport).Close() })
		go func() {
			done <- bot.Run()
		}()
		assert.Nil(t, <-done)
	})
}
