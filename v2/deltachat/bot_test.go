package deltachat

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBot_NewBot(t *testing.T) {
	t.Parallel()
	acfactory.WithRpc(func(rpc *Rpc) {
		bot := NewBot(rpc)
		require.NotNil(t, bot)
	})
}

func TestBot_BotRunningErr(t *testing.T) {
	t.Parallel()
	err := &BotRunningErr{}
	require.NotEmpty(t, err.Error())
}

func TestBot_IsRunning(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineBot(func(bot *Bot, botAcc uint32) {
		require.False(t, bot.IsRunning())
		done := make(chan error)
		go func() { done <- bot.Run() }()
		for !bot.IsRunning() {
			select {
			case err := <-done:
				require.Failf(t, "bot.Run() exited before bot started running", "%v", err)
				return
			case <-time.After(10 * time.Second):
				require.Fail(t, "timeout waiting for bot to start running")
				return
			default:
				time.Sleep(10 * time.Millisecond)
			}
		}
		require.True(t, bot.IsRunning())
		bot.Stop()
		require.Nil(t, <-done)
		require.False(t, bot.IsRunning())
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
			require.Nil(t, err)
			msg := <-incomingMsg
			require.Equal(t, "test1", msg.Text)
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
				require.Nil(t, err)
			})

			chatWithBot := acfactory.CreateChat(accRpc, accId, bot.Rpc, botAcc)
			_, err := accRpc.MiscSendTextMessage(accId, chatWithBot, "test2")
			require.Nil(t, err)
			msg := acfactory.NextMsg(accRpc, accId)
			require.Equal(t, "test2", msg.Text)
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
			require.Nil(t, err)

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
		require.Nil(t, <-done)

		go func() {
			done <- bot.Run()
		}()
		require.Nil(t, <-done)

		bot.On(&EventTypeInfo{}, func(bot *Bot, botAcc uint32, event EventType) { bot.Rpc.Transport.(*IOTransport).Close() })
		go func() {
			done <- bot.Run()
		}()
		require.Nil(t, <-done)
	})
}

func TestBot_RunAlreadyRunning(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredBot(func(bot *Bot, botAcc uint32) {
		done := make(chan error, 1)
		go func() { done <- bot.Run() }()
		// Wait until bot is running.
		for !bot.IsRunning() {
			select {
			case err := <-done:
				require.Failf(t, "bot.Run() exited before bot started running", "%v", err)
				return
			default:
				time.Sleep(5 * time.Millisecond)
			}
		}
		// Now calling Run() again should return BotRunningErr
		err := bot.Run()
		require.NotNil(t, err)
		require.True(t, errors.As(err, new(*BotRunningErr)))
		bot.Stop()
		require.Nil(t, <-done)
	})
}
