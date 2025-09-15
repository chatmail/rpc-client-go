package main // replace with your package name

import (
	"testing"

	"github.com/chatmail/rpc-client-go/deltachat"
	"github.com/stretchr/testify/assert"
)

func TestEchoBot(t *testing.T) {
	acfactory.WithOnlineBot(func(bot *deltachat.Bot, botAcc deltachat.AccountId) {
		go runEchoBot(bot, botAcc) // this is the function we are testing
		acfactory.WithOnlineAccount(func(uRpc *deltachat.Rpc, uAccId deltachat.AccountId) {
			chatId := acfactory.CreateChat(uRpc, uAccId, bot.Rpc, botAcc)
			uRpc.MiscSendTextMessage(uAccId, chatId, "hi")
			msg := acfactory.NextMsg(uRpc, uAccId)
			assert.Equal(t, "hi", msg.Text) // check that bot echoes back the "hi" message from user
		})
	})
}
