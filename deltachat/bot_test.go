package deltachat

import (
	"testing"

	"github.com/chatmail/rpc-client-go/deltachat/option"
	"github.com/chatmail/rpc-client-go/deltachat/transport"
	"github.com/stretchr/testify/assert"
)

func TestBot_NewBot(t *testing.T) {
	t.Parallel()
	acfactory.WithRpc(func(rpc *Rpc) {
		bot := NewBot(rpc)
		assert.NotNil(t, bot)
	})
}

func TestBot_On(t *testing.T) {
	t.Parallel()
	acfactory.WithRunningBot(func(bot *Bot, botAcc AccountId) {
		acfactory.WithOnlineAccount(func(accRpc *Rpc, accId AccountId) {
			incomingMsg := make(chan *MsgSnapshot)
			bot.On(EventIncomingMsg{}, func(bot *Bot, botAcc AccountId, event Event) {
				ev := event.(EventIncomingMsg)
				snapshot, _ := bot.Rpc.GetMessage(botAcc, ev.MsgId)
				incomingMsg <- snapshot
			})

			chatWithBot := acfactory.CreateChat(accRpc, accId, bot.Rpc, botAcc)
			_, err := accRpc.MiscSendTextMessage(accId, chatWithBot, "test1")
			assert.Nil(t, err)
			msg := <-incomingMsg
			assert.Equal(t, "test1", msg.Text)
			bot.RemoveEventHandler(EventIncomingMsg{})
			close(incomingMsg)
		})
	})
}

func TestBot_OnNewMsg(t *testing.T) {
	t.Parallel()
	acfactory.WithRunningBot(func(bot *Bot, botAcc AccountId) {
		acfactory.WithOnlineAccount(func(accRpc *Rpc, accId AccountId) {
			bot.OnNewMsg(func(bot *Bot, botAcc AccountId, msgId MsgId) {
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

func TestBot_processMessages(t *testing.T) {
	t.Parallel()
	acfactory.WithRunningBot(func(bot *Bot, botAcc AccountId) {
		bot.processMessages(botAcc)
	})
}

func TestBot_Stop(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineBot(func(bot *Bot, botAcc AccountId) {
		bot.On(EventInfo{}, func(bot *Bot, botAcc AccountId, event Event) { bot.Stop() })
		done := make(chan error)

		go func() {
			done <- bot.Run()
		}()
		assert.Nil(t, <-done)

		go func() {
			done <- bot.Run()
		}()
		assert.Nil(t, <-done)

		bot.On(EventInfo{}, func(bot *Bot, botAcc AccountId, event Event) { bot.Rpc.Transport.(*transport.IOTransport).Close() })
		go func() {
			done <- bot.Run()
		}()
		assert.Nil(t, <-done)
	})
}

func TestBot_SetUiConfig(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredBot(func(bot *Bot, botAcc AccountId) {
		assert.Nil(t, bot.SetUiConfig(botAcc, "testkey", option.Some("testing")))
		val, err := bot.GetUiConfig(botAcc, "testkey")
		assert.Nil(t, err)
		assert.Equal(t, val.Unwrap(), "testing")

		val, err = bot.GetUiConfig(botAcc, "unknown-key")
		assert.Nil(t, err)
		assert.True(t, val.IsNone())
	})
}
