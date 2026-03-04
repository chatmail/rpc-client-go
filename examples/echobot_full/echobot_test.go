package main // replace with your package name

import (
	"testing"

	"github.com/chatmail/rpc-client-go/v2/deltachat"
	"github.com/stretchr/testify/assert"
)

func TestEchoBot(t *testing.T) {
	acfactory.WithOnlineBot(func(bot *deltachat.Bot, botAcc uint32) {
		go runEchoBot(bot, botAcc) // this is the function we are testing
		acfactory.WithOnlineAccount(func(uRpc *deltachat.Rpc, uAccId uint32) {
			chatId := acfactory.CreateChat(uRpc, uAccId, bot.Rpc, botAcc)
			_, _ = uRpc.MiscSendTextMessage(uAccId, chatId, "hi")
			msg := acfactory.NextMsg(uRpc, uAccId)
			assert.Equal(t, "hi", msg.Text) // check that bot echoes back the "hi" message from user
		})
	})
}
