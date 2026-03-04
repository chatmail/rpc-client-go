package deltachat

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRpc_CheckEmailValidity(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		valid, err := rpc.CheckEmailValidity("test@example.com")
		require.Nil(t, err)
		require.True(t, valid)
	})
}

func TestRpc_MiscSetDraft_and_MiscSendDraft(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		chatId, err := rpc.CreateGroupChat(accId, "test group", true)
		require.Nil(t, err)
		err = rpc.MiscSetDraft(accId, chatId, strptr("test"), nil, nil, nil, nil)
		require.Nil(t, err)
		_, err = rpc.MiscSendDraft(accId, chatId)
		require.Nil(t, err)
	})
}

func TestRpc_SetChatVisibility(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		chatId, err := rpc.CreateGroupChat(accId, "test group", true)
		require.Nil(t, err)
		require.Nil(t, rpc.SetChatVisibility(accId, chatId, ChatVisibilityPinned))
		require.Nil(t, rpc.SetChatVisibility(accId, chatId, ChatVisibilityArchived))
		require.Nil(t, rpc.SetChatVisibility(accId, chatId, ChatVisibilityNormal))
	})
}

func TestRpc_GetChatIdByContactId(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		contactId, err := rpc.CreateContact(accId, "test@example.com", nil)
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
		require.Nil(t, err)
		require.Nil(t, rpc.SelectAccount(accId))
	})
}

func TestAccount_StartIo(t *testing.T) {
	t.Parallel()
	acfactory.WithRpc(func(rpc *Rpc) {
		accId, err := rpc.AddAccount()
		require.Nil(t, err)
		require.Nil(t, rpc.StartIo(accId))
	})
}

func TestAccount_StopIo(t *testing.T) {
	t.Parallel()
	acfactory.WithRpc(func(rpc *Rpc) {
		accId, err := rpc.AddAccount()
		require.Nil(t, err)
		require.Nil(t, rpc.StopIo(accId))
	})
}

func TestAccount_Connectivity(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId uint32) {
		conn, err := rpc.GetConnectivity(accId)
		require.Nil(t, err)
		require.Greater(t, conn, uint32(0))

		html, err := rpc.GetConnectivityHtml(accId)
		require.Nil(t, err)
		require.NotEmpty(t, html)
	})
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		html, err := rpc.GetConnectivityHtml(accId)
		require.Nil(t, err)
		require.NotEmpty(t, html)
	})
}

func TestAccount_Info(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId uint32) {
		info, err := rpc.GetInfo(accId)
		require.Nil(t, err)
		require.NotEmpty(t, info["sqlite_version"])
	})
}

func TestAccount_Size(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId uint32) {
		size, err := rpc.GetAccountFileSize(accId)
		require.Nil(t, err)
		require.NotEqual(t, size, 0)
	})
}

func TestAccount_IsConfigured(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId uint32) {
		configured, err := rpc.IsConfigured(accId)
		require.Nil(t, err)
		require.False(t, configured)

		require.Nil(t, rpc.SetConfigFromQr(accId, acfactory.ConfigQr))
		require.Nil(t, rpc.Configure(accId))

		configured, err = rpc.IsConfigured(accId)
		require.Nil(t, err)
		require.True(t, configured)
	})
}

func TestAccount_SetConfig(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId uint32) {
		require.Nil(t, rpc.SetConfig(accId, "displayname", strptr("test name")))
		name, err := rpc.GetConfig(accId, "displayname")
		require.Nil(t, err)
		require.Equal(t, "test name", *name)

		err = rpc.BatchSetConfig(accId, map[string]*string{
			"displayname": strptr("new name"),
			"selfstatus":  strptr("test status"),
		})
		require.Nil(t, err)
		name, err = rpc.GetConfig(accId, "displayname")
		require.Nil(t, err)
		require.Equal(t, "new name", *name)

		require.Nil(t, rpc.SetConfig(accId, "selfavatar", acfactory.TestImage()))
	})
}

func TestAccount_Remove(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId uint32) {
		require.Nil(t, rpc.RemoveAccount(accId))
	})
}

func TestAccount_Configure(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId uint32) {
		require.Nil(t, rpc.SetConfigFromQr(accId, acfactory.ConfigQr))
		require.Nil(t, rpc.Configure(accId))
	})
}

func TestAccount_Contacts(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		ids, err := rpc.GetContactIds(accId, 0, nil)
		require.Nil(t, err)
		require.Empty(t, ids)

		ids, err = rpc.GetContactIds(accId, 0, strptr("unknown"))
		require.Nil(t, err)
		require.Empty(t, ids)
	})
}

func TestAccount_GetContactByAddr(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		contactId, err := rpc.CreateContact(accId, "null@localhost", strptr("test"))
		require.Nil(t, err)
		require.NotNil(t, contactId)

		contactId2, err := rpc.LookupContactIdByAddr(accId, "unknown@example.com")
		require.Nil(t, err)
		require.Nil(t, contactId2)

		contactId2, err = rpc.LookupContactIdByAddr(accId, "null@localhost")
		require.Nil(t, err)
		require.NotNil(t, contactId2)
		require.Equal(t, contactId, *contactId2)
	})
}

func TestAccount_BlockedContacts(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		contactId, err := rpc.CreateContact(accId, "null@localhost", strptr("test"))
		require.Nil(t, err)

		blocked, err := rpc.GetBlockedContacts(accId)
		require.Nil(t, err)
		require.Empty(t, blocked)

		require.Nil(t, rpc.BlockContact(accId, contactId))

		blocked, err = rpc.GetBlockedContacts(accId)
		require.Nil(t, err)
		require.NotEmpty(t, blocked)
	})
}

func TestAccount_CreateBroadcastList(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		_, err := rpc.CreateBroadcastList(accId)
		require.Nil(t, err)
	})
}

func TestAccount_CreateGroup(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		_, err := rpc.CreateGroupChat(accId, "test group", true)
		require.Nil(t, err)
	})
}

func TestAccount_GetChatSecurejoinQrCodeSvg(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc1 *Rpc, accId1 uint32) {
		pair, err := rpc1.GetChatSecurejoinQrCodeSvg(accId1, nil)
		require.Nil(t, err)
		require.NotEmpty(t, pair.First)
		require.NotEmpty(t, pair.Second)

		acfactory.WithOnlineAccount(func(rpc2 *Rpc, accId2 uint32) {
			_, err := rpc2.SecureJoin(accId2, pair.First)
			require.Nil(t, err)
			acfactory.WaitForEvent(rpc1, accId1, &EventTypeSecurejoinInviterProgress{})
			acfactory.WaitForEvent(rpc2, accId2, &EventTypeSecurejoinJoinerProgress{})
		})
	})
}

func TestAccount_GetChatSecurejoinQrCode(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc1 *Rpc, accId1 uint32) {
		qrdata, err := rpc1.GetChatSecurejoinQrCode(accId1, nil)
		require.Nil(t, err)
		require.NotEmpty(t, qrdata)

		acfactory.WithOnlineAccount(func(rpc2 *Rpc, accId2 uint32) {
			_, err := rpc2.SecureJoin(accId2, qrdata)
			require.Nil(t, err)
			acfactory.WaitForEvent(rpc1, accId1, &EventTypeSecurejoinInviterProgress{})
			acfactory.WaitForEvent(rpc2, accId2, &EventTypeSecurejoinJoinerProgress{})
		})
	})
}

func TestAccount_ImportBackup(t *testing.T) {
	t.Parallel()
	var backup string
	passphrase := strptr("password")
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		dir := acfactory.MkdirTemp()
		require.Nil(t, rpc.ExportBackup(accId, dir, passphrase))
		files, err := os.ReadDir(dir)
		require.Nil(t, err)
		require.Equal(t, len(files), 1)
		backup = filepath.Join(dir, files[0].Name())
		require.FileExists(t, backup)
	})

	acfactory.WithRpc(func(rpc *Rpc) {
		accId, err := rpc.AddAccount()
		require.Nil(t, err)
		require.Nil(t, rpc.ImportBackup(accId, backup, passphrase))
		_, err = rpc.GetSystemInfo()
		require.Nil(t, err)
	})
}

func TestAccount_ExportBackup(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		dir := acfactory.MkdirTemp()
		require.Nil(t, rpc.ExportBackup(accId, dir, strptr("test-phrase")))
		files, err := os.ReadDir(dir)
		require.Nil(t, err)
		require.Equal(t, len(files), 1)
	})
}

func TestAccount_GetBackup(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc1 *Rpc, accId1 uint32) {
		go func() { require.Nil(t, rpc1.ProvideBackup(accId1)) }()
		var err error
		var qrData string
		qrData, err = rpc1.GetBackupQr(accId1)
		for err != nil {
			time.Sleep(time.Millisecond * 200)
			qrData, err = rpc1.GetBackupQr(accId1)
		}
		require.NotNil(t, qrData)

		qrSvg, err := rpc1.GetBackupQrSvg(accId1)
		require.Nil(t, err)
		require.NotNil(t, qrSvg)

		acfactory.WithRpc(func(rpc2 *Rpc) {
			accId2, err := rpc2.AddAccount()
			require.Nil(t, err)
			require.Nil(t, rpc2.GetBackup(accId2, qrData))
		})
	})
}

func TestAccount_InitiateAutocryptKeyTransfer(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		code, err := rpc.InitiateAutocryptKeyTransfer(accId)
		require.Nil(t, err)
		require.NotEmpty(t, code)
	})
}

func TestAccount_FreshMsgs(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc1 *Rpc, accId1 uint32) {
		acfactory.WithOnlineAccount(func(rpc2 *Rpc, accId2 uint32) {
			chatId2 := acfactory.CreateChat(rpc2, accId2, rpc1, accId1)
			_, err := rpc2.MiscSendTextMessage(accId2, chatId2, "hi")
			require.Nil(t, err)
			msg := acfactory.NextMsg(rpc1, accId1)
			require.Equal(t, msg.Text, "hi")

			msgs, err := rpc1.GetFreshMsgs(accId1)
			require.Nil(t, err)
			require.NotEmpty(t, msgs)

			require.Nil(t, rpc1.MarkseenMsgs(accId1, msgs))

			msgs, err = rpc1.GetFreshMsgs(accId1)
			require.Nil(t, err)
			require.Empty(t, msgs)
		})
	})
}

func TestAccount_GetNextMsgs(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineBot(func(bot *Bot, botAcc uint32) {
		msgs, err := bot.Rpc.GetNextMsgs(botAcc)
		require.Nil(t, err)
		require.Empty(t, msgs)
		acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
			msgs, err := rpc.GetNextMsgs(accId)
			require.Nil(t, err)
			require.NotEmpty(t, msgs) // messages from device chat

			require.Nil(t, rpc.MarkseenMsgs(accId, msgs))

			msgs, err = rpc.GetNextMsgs(accId)
			require.Nil(t, err)
			require.Empty(t, msgs)
		})
	})
}

func TestAccount_DeleteMsgs(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		chatId, err := rpc.CreateGroupChat(accId, "test group", true)
		require.Nil(t, err)
		_, err = rpc.MiscSendTextMessage(accId, chatId, "hi")
		require.Nil(t, err)

		msgs, err := rpc.GetMessageIds(accId, chatId, false, false)
		require.Nil(t, err)
		require.NotEmpty(t, msgs)

		require.Nil(t, rpc.DeleteMessages(accId, msgs))

		msgs, err = rpc.GetMessageIds(accId, chatId, false, false)
		require.Nil(t, err)
		require.Empty(t, msgs)
	})
}

func TestAccount_SearchMessages(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		chatId, err := rpc.CreateGroupChat(accId, "test group", true)
		require.Nil(t, err)
		msgId, err := rpc.MiscSendTextMessage(accId, chatId, "hi")
		require.Nil(t, err)

		msgs, err := rpc.SearchMessages(accId, "hi", nil)
		require.Nil(t, err)
		require.NotEmpty(t, msgs)
		require.Equal(t, msgId, msgs[0])
	})
}

func TestAccount_GetChatlistEntries(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		_, err := rpc.CreateGroupChat(accId, "test group", true)
		require.Nil(t, err)

		entries, err := rpc.GetChatlistEntries(accId, nil, strptr("unknown"), nil)
		require.Nil(t, err)
		require.Empty(t, entries)

		entries, err = rpc.GetChatlistEntries(accId, nil, nil, nil)
		require.Nil(t, err)
		require.NotEmpty(t, entries)

		items, err := rpc.GetChatlistItemsByEntries(accId, entries)
		require.Nil(t, err)
		require.NotEmpty(t, items)
	})
}

func TestAccount_AddDeviceMsg(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		msgId, err := rpc.AddDeviceMessage(accId, "test", &MessageData{Text: strptr("new message")})
		require.Nil(t, err)
		msg, err := rpc.GetMessage(accId, *msgId)
		require.Nil(t, err)
		require.Equal(t, msg.Text, "new message")
	})
}

func TestChat_Basics(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		chatId, err := rpc.CreateGroupChat(accId, "test group", true)
		require.Nil(t, err)
		require.Nil(t, rpc.AcceptChat(accId, chatId))
		require.Nil(t, rpc.MarknoticedChat(accId, chatId))
		_, err = rpc.GetFirstUnreadMessageOfChat(accId, chatId)
		require.Nil(t, err)

		_, err = rpc.GetBasicChatInfo(accId, chatId)
		require.Nil(t, err)

		_, err = rpc.GetFullChatById(accId, chatId)
		require.Nil(t, err)

		require.Nil(t, rpc.BlockChat(accId, chatId))

		chatId, err = rpc.CreateGroupChat(accId, "test group 2", true)
		require.Nil(t, err)
		require.Nil(t, rpc.DeleteChat(accId, chatId))
	})
}

func TestChat_Groups(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		chatId, err := rpc.CreateGroupChat(accId, "test group", false)
		require.Nil(t, err)
		require.Nil(t, rpc.SetChatProfileImage(accId, chatId, acfactory.TestImage()))
		require.Nil(t, rpc.SetChatProfileImage(accId, chatId, nil))
		require.Nil(t, rpc.SetChatName(accId, chatId, "new name"))

		_, err = rpc.GetChatContacts(accId, chatId)
		require.Nil(t, err)

		require.Nil(t, rpc.SetChatEphemeralTimer(accId, chatId, 9000))

		_, err = rpc.GetChatEncryptionInfo(accId, chatId)
		require.Nil(t, err)

		_, err = rpc.SendMsg(accId, chatId, MessageData{Text: strptr("test message")})
		require.Nil(t, err)
	})
}

func TestMsg_Reactions(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		chatId, err := rpc.CreateGroupChat(accId, "test group", false)
		require.Nil(t, err)

		var msgId uint32
		msgId, err = rpc.SendMsg(accId, chatId, MessageData{Text: strptr("test message")})
		require.Nil(t, err)

		_, err = rpc.SendReaction(accId, msgId, []string{":)"})
		require.Nil(t, err)

		data, err2 := rpc.GetMessageReactions(accId, msgId)
		require.Nil(t, err2)
		reactions := data.Reactions
		require.Len(t, reactions, 1)
		require.Equal(t, reactions[0].Emoji, ":)")

		msg, err := rpc.GetMessage(accId, msgId)
		require.Nil(t, err)
		reactions = msg.Reactions.Reactions
		require.Len(t, reactions, 1)
		require.Equal(t, reactions[0].Emoji, ":)")
	})
}
