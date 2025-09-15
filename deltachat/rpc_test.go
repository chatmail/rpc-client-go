package deltachat

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chatmail/rpc-client-go/deltachat/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRpc_CheckEmailValidity(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
		valid, err := rpc.CheckEmailValidity("test@example.com")
		require.Nil(t, err)
		require.True(t, valid)
	})
}

func TestRpc_MiscSetDraft_and_MiscSendDraft(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
		chatId, err := rpc.CreateGroupChat(accId, "test group", true)
		require.Nil(t, err)
		err = rpc.MiscSetDraft(accId, chatId, option.Some("test"), option.None[string](), option.None[string](), option.None[MsgId](), option.None[MsgType]())
		require.Nil(t, err)
		_, err = rpc.MiscSendDraft(accId, chatId)
		require.Nil(t, err)
	})
}

func TestRpc_SetChatVisibility(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
		chatId, err := rpc.CreateGroupChat(accId, "test group", true)
		assert.Nil(t, err)
		assert.Nil(t, rpc.SetChatVisibility(accId, chatId, ChatVisibilityPinned))
		assert.Nil(t, rpc.SetChatVisibility(accId, chatId, ChatVisibilityArchived))
		assert.Nil(t, rpc.SetChatVisibility(accId, chatId, ChatVisibilityNormal))
	})
}

func TestRpc_GetChatIdByContactId(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
		contactId, err := rpc.CreateContact(accId, "test@example.com", "")
		require.Nil(t, err)
		chatId, err := rpc.GetChatIdByContactId(accId, contactId)
		require.Nil(t, err)
		require.NotEqual(t, chatId, 0)
	})
}

func TestAccount_Select(t *testing.T) {
	t.Parallel()
	acfactory.WithRpc(func(rpc *Rpc) {
		accId, err := rpc.AddAccount()
		assert.Nil(t, err)
		assert.Nil(t, rpc.SelectAccount(accId))
	})
}

func TestAccount_StartIo(t *testing.T) {
	t.Parallel()
	acfactory.WithRpc(func(rpc *Rpc) {
		accId, err := rpc.AddAccount()
		assert.Nil(t, err)
		assert.Nil(t, rpc.StartIo(accId))
	})
}

func TestAccount_StopIo(t *testing.T) {
	t.Parallel()
	acfactory.WithRpc(func(rpc *Rpc) {
		accId, err := rpc.AddAccount()
		assert.Nil(t, err)
		assert.Nil(t, rpc.StopIo(accId))
	})
}

func TestAccount_Connectivity(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId AccountId) {
		conn, err := rpc.GetConnectivity(accId)
		require.Nil(t, err)
		require.Greater(t, conn, uint(0))

		html, err := rpc.GetConnectivityHtml(accId)
		require.Nil(t, err)
		require.NotEmpty(t, html)
	})
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
		html, err := rpc.GetConnectivityHtml(accId)
		require.Nil(t, err)
		require.NotEmpty(t, html)
	})
}

func TestAccount_Info(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId AccountId) {
		info, err := rpc.GetInfo(accId)
		assert.Nil(t, err)
		assert.NotEmpty(t, info["sqlite_version"])
	})
}

func TestAccount_Size(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId AccountId) {
		size, err := rpc.GetAccountFileSize(accId)
		assert.Nil(t, err)
		assert.NotEqual(t, size, 0)
	})
}

func TestAccount_IsConfigured(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId AccountId) {
		configured, err := rpc.IsConfigured(accId)
		assert.Nil(t, err)
		assert.False(t, configured)

		assert.Nil(t, rpc.Configure(accId))

		configured, err = rpc.IsConfigured(accId)
		assert.Nil(t, err)
		assert.True(t, configured)
	})
}

func TestAccount_SetConfig(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId AccountId) {
		assert.Nil(t, rpc.SetConfig(accId, "displayname", option.Some("test name")))
		name, err := rpc.GetConfig(accId, "displayname")
		assert.Nil(t, err)
		assert.Equal(t, name.Unwrap(), "test name")

		err = rpc.BatchSetConfig(accId, map[string]option.Option[string]{
			"displayname": option.Some("new name"),
			"selfstatus":  option.Some("test status"),
		})
		assert.Nil(t, err)
		name, err = rpc.GetConfig(accId, "displayname")
		assert.Nil(t, err)
		assert.Equal(t, name.Unwrap(), "new name")

		assert.Nil(t, rpc.SetConfig(accId, "selfavatar", option.Some(acfactory.TestImage())))
	})
}

func TestAccount_Remove(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId AccountId) {
		assert.Nil(t, rpc.RemoveAccount(accId))
	})
}

func TestAccount_Configure(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId AccountId) {
		assert.Nil(t, rpc.Configure(accId))
	})
}

func TestAccount_Contacts(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
		ids, err := rpc.GetContactIds(accId, 0, option.None[string]())
		assert.Nil(t, err)
		assert.Empty(t, ids)

		ids, err = rpc.GetContactIds(accId, 0, option.Some("unknown"))
		assert.Nil(t, err)
		assert.Empty(t, ids)
	})
}

func TestAccount_GetContactByAddr(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
		contactId, err := rpc.CreateContact(accId, "null@localhost", "test")
		assert.Nil(t, err)
		assert.NotNil(t, contactId)

		contactId2, err := rpc.LookupContactIdByAddr(accId, "unknown@example.com")
		assert.Nil(t, err)
		assert.True(t, contactId2.IsNone())

		contactId2, err = rpc.LookupContactIdByAddr(accId, "null@localhost")
		assert.Nil(t, err)
		assert.True(t, contactId2.IsSome())
		assert.Equal(t, contactId, contactId2.Unwrap())
	})
}

func TestAccount_BlockedContacts(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
		contactId, err := rpc.CreateContact(accId, "null@localhost", "test")
		assert.Nil(t, err)

		blocked, err := rpc.GetBlockedContacts(accId)
		assert.Nil(t, err)
		assert.Empty(t, blocked)

		assert.Nil(t, rpc.BlockContact(accId, contactId))

		blocked, err = rpc.GetBlockedContacts(accId)
		assert.Nil(t, err)
		assert.NotEmpty(t, blocked)
	})
}

func TestAccount_CreateBroadcastList(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
		_, err := rpc.CreateBroadcastList(accId)
		assert.Nil(t, err)
	})
}

func TestAccount_CreateGroup(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
		_, err := rpc.CreateGroupChat(accId, "test group", true)
		assert.Nil(t, err)
	})
}

func TestAccount_QrCode(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc1 *Rpc, accId1 AccountId) {
		qrdata, svg, err := rpc1.GetChatSecurejoinQrCodeSvg(accId1, option.None[ChatId]())
		assert.Nil(t, err)
		assert.NotEmpty(t, qrdata)
		assert.NotEmpty(t, svg)

		acfactory.WithOnlineAccount(func(rpc2 *Rpc, accId2 AccountId) {
			acfactory.IntroduceEachOther(rpc1, accId1, rpc2, accId2)
			_, err := rpc2.SecureJoin(accId2, qrdata)
			assert.Nil(t, err)
			acfactory.WaitForEvent(rpc1, accId1, EventSecurejoinInviterProgress{})
			acfactory.WaitForEvent(rpc2, accId2, EventSecurejoinJoinerProgress{})
		})
	})
}

func TestAccount_ImportBackup(t *testing.T) {
	t.Parallel()
	var backup string
	passphrase := option.Some("password")
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
		dir := acfactory.MkdirTemp()
		assert.Nil(t, rpc.ExportBackup(accId, dir, passphrase))
		files, err := os.ReadDir(dir)
		assert.Nil(t, err)
		assert.Equal(t, len(files), 1)
		backup = filepath.Join(dir, files[0].Name())
		assert.FileExists(t, backup)
	})

	acfactory.WithRpc(func(rpc *Rpc) {
		accId, err := rpc.AddAccount()
		assert.Nil(t, err)
		assert.Nil(t, rpc.ImportBackup(accId, backup, passphrase))
		_, err = rpc.GetSystemInfo()
		assert.Nil(t, err)
	})
}

func TestAccount_ExportBackup(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
		dir := acfactory.MkdirTemp()
		assert.Nil(t, rpc.ExportBackup(accId, dir, option.Some("test-phrase")))
		files, err := os.ReadDir(dir)
		assert.Nil(t, err)
		assert.Equal(t, len(files), 1)
	})
}

func TestAccount_GetBackup(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc1 *Rpc, accId1 AccountId) {
		go func() { assert.Nil(t, rpc1.ProvideBackup(accId1)) }()
		var err error
		var qrData string
		qrData, err = rpc1.GetBackupQr(accId1)
		for err != nil {
			time.Sleep(time.Millisecond * 200)
			qrData, err = rpc1.GetBackupQr(accId1)
		}
		assert.NotNil(t, qrData)

		qrSvg, err := rpc1.GetBackupQrSvg(accId1)
		assert.Nil(t, err)
		assert.NotNil(t, qrSvg)

		acfactory.WithRpc(func(rpc2 *Rpc) {
			accId2, err := rpc2.AddAccount()
			assert.Nil(t, err)
			assert.Nil(t, rpc2.GetBackup(accId2, qrData))
		})
	})
}

func TestAccount_InitiateAutocryptKeyTransfer(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
		code, err := rpc.InitiateAutocryptKeyTransfer(accId)
		assert.Nil(t, err)
		assert.NotEmpty(t, code)
	})
}

func TestAccount_FreshMsgs(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc1 *Rpc, accId1 AccountId) {
		acfactory.WithOnlineAccount(func(rpc2 *Rpc, accId2 AccountId) {
			chatId2 := acfactory.CreateChat(rpc2, accId2, rpc1, accId1)
			_, err := rpc2.MiscSendTextMessage(accId2, chatId2, "hi")
			assert.Nil(t, err)
			msg := acfactory.NextMsg(rpc1, accId1)
			assert.Equal(t, msg.Text, "hi")

			msgs, err := rpc1.GetFreshMsgs(accId1)
			assert.Nil(t, err)
			assert.NotEmpty(t, msgs)

			assert.Nil(t, rpc1.MarkseenMsgs(accId1, msgs))

			msgs, err = rpc1.GetFreshMsgs(accId1)
			assert.Nil(t, err)
			assert.Empty(t, msgs)
		})
	})
}

func TestAccount_GetNextMsgs(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineBot(func(bot *Bot, botAcc AccountId) {
		msgs, err := bot.Rpc.GetNextMsgs(botAcc)
		assert.Nil(t, err)
		assert.Empty(t, msgs)
		acfactory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
			msgs, err := rpc.GetNextMsgs(accId)
			assert.Nil(t, err)
			assert.NotEmpty(t, msgs) // messages from device chat

			assert.Nil(t, rpc.MarkseenMsgs(accId, msgs))

			msgs, err = rpc.GetNextMsgs(accId)
			assert.Nil(t, err)
			assert.Empty(t, msgs)
		})
	})
}

func TestAccount_DeleteMsgs(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
		chatId, err := rpc.CreateGroupChat(accId, "test group", true)
		assert.Nil(t, err)
		_, err = rpc.MiscSendTextMessage(accId, chatId, "hi")
		assert.Nil(t, err)

		msgs, err := rpc.GetMessageIds(accId, chatId, false, false)
		assert.Nil(t, err)
		assert.NotEmpty(t, msgs)

		assert.Nil(t, rpc.DeleteMessages(accId, msgs))

		msgs, err = rpc.GetMessageIds(accId, chatId, false, false)
		assert.Nil(t, err)
		assert.Empty(t, msgs)
	})
}

func TestAccount_SearchMessages(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
		chatId, err := rpc.CreateGroupChat(accId, "test group", true)
		assert.Nil(t, err)
		msgId, err := rpc.MiscSendTextMessage(accId, chatId, "hi")
		assert.Nil(t, err)

		msgs, err := rpc.SearchMessages(accId, "hi", option.None[ChatId]())
		assert.Nil(t, err)
		assert.NotEmpty(t, msgs)
		assert.Equal(t, msgId, msgs[0])
	})
}

func TestAccount_GetChatlistEntries(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
		_, err := rpc.CreateGroupChat(accId, "test group", true)
		assert.Nil(t, err)

		noFlag := option.None[uint]()
		noContact := option.None[ContactId]()
		entries, err := rpc.GetChatlistEntries(accId, noFlag, option.Some("unknown"), noContact)
		assert.Nil(t, err)
		assert.Empty(t, entries)

		entries, err = rpc.GetChatlistEntries(accId, noFlag, option.None[string](), noContact)
		assert.Nil(t, err)
		assert.NotEmpty(t, entries)

		items, err := rpc.GetChatlistItemsByEntries(accId, entries)
		assert.Nil(t, err)
		assert.NotEmpty(t, items)
	})
}

func TestAccount_AddDeviceMsg(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
		msgId, err := rpc.AddDeviceMessage(accId, "test", option.Some(MsgData{Text: "new message"}))
		require.Nil(t, err)
		msg, err := rpc.GetMessage(accId, msgId)
		require.Nil(t, err)
		require.Equal(t, msg.Text, "new message")
	})
}

func TestChat_Basics(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
		chatId, err := rpc.CreateGroupChat(accId, "test group", true)
		assert.Nil(t, err)
		assert.Nil(t, rpc.AcceptChat(accId, chatId))
		assert.Nil(t, rpc.MarknoticedChat(accId, chatId))
		_, err = rpc.GetFirstUnreadMessageOfChat(accId, chatId)
		assert.Nil(t, err)

		_, err = rpc.GetBasicChatInfo(accId, chatId)
		assert.Nil(t, err)

		_, err = rpc.GetFullChatById(accId, chatId)
		assert.Nil(t, err)

		assert.Nil(t, rpc.BlockChat(accId, chatId))

		chatId, err = rpc.CreateGroupChat(accId, "test group 2", true)
		assert.Nil(t, err)
		assert.Nil(t, rpc.DeleteChat(accId, chatId))
	})
}

func TestChat_Groups(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
		chatId, err := rpc.CreateGroupChat(accId, "test group", false)
		assert.Nil(t, err)
		assert.Nil(t, rpc.SetChatProfileImage(accId, chatId, option.Some(acfactory.TestImage())))
		assert.Nil(t, rpc.SetChatProfileImage(accId, chatId, option.None[string]()))
		assert.Nil(t, rpc.SetChatName(accId, chatId, "new name"))

		_, err = rpc.GetChatContacts(accId, chatId)
		assert.Nil(t, err)

		assert.Nil(t, rpc.SetChatEphemeralTimer(accId, chatId, 9000))

		_, err = rpc.GetChatEncryptionInfo(accId, chatId)
		assert.Nil(t, err)

		_, err = rpc.SendMsg(accId, chatId, MsgData{Text: "test message"})
		assert.Nil(t, err)
	})
}

func TestMsg_Reactions(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId AccountId) {
		chatId, err := rpc.CreateGroupChat(accId, "test group", false)
		require.Nil(t, err)

		var msgId MsgId
		msgId, err = rpc.SendMsg(accId, chatId, MsgData{Text: "test message"})
		require.Nil(t, err)

		_, err = rpc.SendReaction(accId, msgId, ":)")
		require.Nil(t, err)

		data, err2 := rpc.GetMessageReactions(accId, msgId)
		require.Nil(t, err2)
		reactions := data.Unwrap().Reactions
		require.Len(t, reactions, 1)
		require.Equal(t, reactions[0].Emoji, ":)")

		msg, err := rpc.GetMessage(accId, msgId)
		require.Nil(t, err)
		reactions = msg.Reactions.Reactions
		require.Len(t, reactions, 1)
		require.Equal(t, reactions[0].Emoji, ":)")
	})
}
