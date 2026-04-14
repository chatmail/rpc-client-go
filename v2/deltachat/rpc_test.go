package deltachat

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/creachadair/jrpc2"
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
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		err := rpc.MiscSetDraft(accId, chatId, strptr("test"), nil, nil, nil, nil)
		require.Nil(t, err)
		_, err = rpc.MiscSendDraft(accId, chatId)
		require.Nil(t, err)
	})
}

func TestRpc_SetChatVisibility(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
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

		require.Nil(t, rpc.SetConfig(accId, "selfavatar", strptr(acfactory.TestImage())))
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
		addr := "user@example.com"
		contactId, err := rpc.CreateContact(accId, addr, strptr("test"))
		require.Nil(t, err)
		require.NotNil(t, contactId)

		contactId2, err := rpc.LookupContactIdByAddr(accId, "unknown@example.com")
		require.Nil(t, err)
		require.Nil(t, contactId2)

		contactId2, err = rpc.LookupContactIdByAddr(accId, addr)
		require.Nil(t, err)
		require.NotNil(t, contactId2)
		require.Equal(t, contactId, *contactId2)
	})
}

func TestAccount_BlockedContacts(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		contactId, err := rpc.CreateContact(accId, "user@example.com", strptr("test"))
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
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		_, err := rpc.MiscSendTextMessage(accId, chatId, "hi")
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
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
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
		entries, err := rpc.GetChatlistEntries(accId, nil, strptr("unknown"), nil)
		require.Nil(t, err)
		require.Empty(t, entries)

		entries, err = rpc.GetChatlistEntries(accId, nil, nil, nil)
		require.Nil(t, err)
		require.NotEmpty(t, entries)

		items, err := rpc.GetChatlistItemsByEntries(accId, entries)
		require.Nil(t, err)
		require.NotEmpty(t, items)

		rpc.Transport.(*IOTransport).Close()
		_, err = rpc.GetChatlistItemsByEntries(accId, entries)
		require.NotNil(t, err)
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
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		require.Nil(t, rpc.AcceptChat(accId, chatId))
		require.Nil(t, rpc.MarknoticedChat(accId, chatId))
		_, err := rpc.GetFirstUnreadMessageOfChat(accId, chatId)
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
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		require.Nil(t, rpc.SetChatProfileImage(accId, chatId, strptr(acfactory.TestImage())))
		require.Nil(t, rpc.SetChatProfileImage(accId, chatId, nil))
		require.Nil(t, rpc.SetChatName(accId, chatId, "new name"))

		_, err := rpc.GetChatContacts(accId, chatId)
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
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		msgId, err := rpc.SendMsg(accId, chatId, MessageData{Text: strptr("test message")})
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

func TestRpc_GetSystemInfo(t *testing.T) {
	t.Parallel()
	acfactory.WithRpc(func(rpc *Rpc) {
		info, err := rpc.GetSystemInfo()
		require.Nil(t, err)
		require.NotEmpty(t, info)
	})
}

func TestAccount_GetAllAccountIds(t *testing.T) {
	t.Parallel()
	acfactory.WithRpc(func(rpc *Rpc) {
		ids, err := rpc.GetAllAccountIds()
		require.Nil(t, err)
		require.Empty(t, ids)

		accId, err := rpc.AddAccount()
		require.Nil(t, err)

		ids, err = rpc.GetAllAccountIds()
		require.Nil(t, err)
		require.Contains(t, ids, accId)
	})
}

func TestAccount_GetSelectedAccountId(t *testing.T) {
	t.Parallel()
	acfactory.WithRpc(func(rpc *Rpc) {
		accId, err := rpc.AddAccount()
		require.Nil(t, err)
		require.Nil(t, rpc.SelectAccount(accId))

		selected, err := rpc.GetSelectedAccountId()
		require.Nil(t, err)
		require.NotNil(t, selected)
		require.Equal(t, accId, *selected)
	})
}

func TestAccount_GetAllAccounts(t *testing.T) {
	t.Parallel()
	acfactory.WithRpc(func(rpc *Rpc) {
		_, err := rpc.AddAccount()
		require.Nil(t, err)

		accounts, err := rpc.GetAllAccounts()
		require.Nil(t, err)
		require.NotEmpty(t, accounts)

		rpc.Transport.(*IOTransport).Close()
		_, err = rpc.GetAllAccounts()
		require.NotNil(t, err)
	})
}

func TestAccount_GetAccountInfo(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId uint32) {
		info, err := rpc.GetAccountInfo(accId)
		require.Nil(t, err)
		require.NotNil(t, info)
		require.Equal(t, (&AccountUnconfigured{}).GetKind(), info.GetKind())

		rpc.Transport.(*IOTransport).Close()
		_, err = rpc.GetAccountInfo(accId)
		require.NotNil(t, err)
	})
}

func TestAccount_StartStopIoForAllAccounts(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId uint32) {
		require.Nil(t, rpc.StartIoForAllAccounts())
		require.Nil(t, rpc.StopIoForAllAccounts())
	})
}

func TestAccount_GetBlobDir(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId uint32) {
		blobDir, err := rpc.GetBlobDir(accId)
		require.Nil(t, err)
		require.NotNil(t, blobDir)
	})
}

func TestAccount_GetStorageUsageReportString(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId uint32) {
		report, err := rpc.GetStorageUsageReportString(accId)
		require.Nil(t, err)
		require.NotEmpty(t, report)
	})
}

func TestAccount_GetMigrationError(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId uint32) {
		migErr, err := rpc.GetMigrationError(accId)
		require.Nil(t, err)
		require.Nil(t, migErr)
	})
}

func TestAccount_BatchGetConfig(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId uint32) {
		require.Nil(t, rpc.SetConfig(accId, "displayname", strptr("Alice")))

		cfg, err := rpc.BatchGetConfig(accId, []string{"displayname", "selfstatus"})
		require.Nil(t, err)
		require.NotNil(t, cfg["displayname"])
		require.Equal(t, "Alice", *cfg["displayname"])
	})
}

func TestAccount_GetAllUiConfigKeys(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId uint32) {
		keys, err := rpc.GetAllUiConfigKeys(accId)
		require.Nil(t, err)
		require.NotNil(t, keys)
	})
}

func TestAccount_SetAccountsOrder(t *testing.T) {
	t.Parallel()
	acfactory.WithRpc(func(rpc *Rpc) {
		acc1, err := rpc.AddAccount()
		require.Nil(t, err)
		acc2, err := rpc.AddAccount()
		require.Nil(t, err)

		require.Nil(t, rpc.SetAccountsOrder([]uint32{acc2, acc1}))

		ids, err := rpc.GetAllAccountIds()
		require.Nil(t, err)
		require.Equal(t, []uint32{acc2, acc1}, ids)
	})
}

func TestAccount_EstimateAutoDeletionCount(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		count, err := rpc.EstimateAutoDeletionCount(accId, false, 3600)
		require.Nil(t, err)
		require.Equal(t, uint(0), count)
	})
}

func TestAccount_MaybeNetwork(t *testing.T) {
	t.Parallel()
	acfactory.WithRpc(func(rpc *Rpc) {
		require.Nil(t, rpc.MaybeNetwork())
	})
}

func TestAccount_GetProviderInfo(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accId uint32) {
		info, err := rpc.GetProviderInfo(accId, "user@gmail.com")
		require.Nil(t, err)
		require.NotNil(t, info)
	})
}

func TestContact_GetContact(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		addr := "test@example.com"
		contactId, err := rpc.CreateContact(accId, addr, strptr("Test User"))
		require.Nil(t, err)

		contact, err := rpc.GetContact(accId, contactId)
		require.Nil(t, err)
		require.Equal(t, addr, contact.Address)
	})
}

func TestContact_GetContacts(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		_, err := rpc.CreateContact(accId, "alice@example.com", strptr("Alice"))
		require.Nil(t, err)

		contacts, err := rpc.GetContacts(accId, ContactFlagAddress, nil)
		require.Nil(t, err)
		require.NotEmpty(t, contacts)

		contacts, err = rpc.GetContacts(accId, 0, nil)
		require.Nil(t, err)
		require.Empty(t, contacts)
	})
}

func TestContact_GetContactsByIds(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		contactId, err := rpc.CreateContact(accId, "bob@example.com", strptr("Bob"))
		require.Nil(t, err)

		contacts, err := rpc.GetContactsByIds(accId, []uint32{contactId})
		require.Nil(t, err)
		require.NotEmpty(t, contacts)
	})
}

func TestContact_DeleteContact(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		contactId, err := rpc.CreateContact(accId, "delete@example.com", nil)
		require.Nil(t, err)
		require.Nil(t, rpc.DeleteContact(accId, contactId))
	})
}

func TestContact_UnblockContact(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		contactId, err := rpc.CreateContact(accId, "block@example.com", nil)
		require.Nil(t, err)
		require.Nil(t, rpc.BlockContact(accId, contactId))
		require.Nil(t, rpc.UnblockContact(accId, contactId))

		blocked, err := rpc.GetBlockedContacts(accId)
		require.Nil(t, err)
		require.Empty(t, blocked)
	})
}

func TestContact_ChangeContactName(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		contactId, err := rpc.CreateContact(accId, "rename@example.com", strptr("OldName"))
		require.Nil(t, err)
		require.Nil(t, rpc.ChangeContactName(accId, contactId, "NewName"))

		contact, err := rpc.GetContact(accId, contactId)
		require.Nil(t, err)
		require.Equal(t, "NewName", contact.Name)
	})
}

func TestContact_GetContactEncryptionInfo(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		acfactory.WithOnlineAccount(func(rpc2 *Rpc, accId2 uint32) {
			contactId := acfactory.ImportContact(rpc, accId, rpc2, accId2)
			info, err := rpc.GetContactEncryptionInfo(accId, contactId)
			require.Nil(t, err)
			require.NotEmpty(t, info)
		})
	})
}

func TestChat_Description(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		require.Nil(t, rpc.SetChatDescription(accId, chatId, "A test description"))

		desc, err := rpc.GetChatDescription(accId, chatId)
		require.Nil(t, err)
		require.Equal(t, "A test description", desc)
	})
}

func TestChat_MuteDuration(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		muted, err := rpc.IsChatMuted(accId, chatId)
		require.Nil(t, err)
		require.False(t, muted)

		require.Nil(t, rpc.SetChatMuteDuration(accId, chatId, &MuteDurationForever{}))

		muted, err = rpc.IsChatMuted(accId, chatId)
		require.Nil(t, err)
		require.True(t, muted)

		require.Nil(t, rpc.SetChatMuteDuration(accId, chatId, &MuteDurationNotMuted{}))

		muted, err = rpc.IsChatMuted(accId, chatId)
		require.Nil(t, err)
		require.False(t, muted)
	})
}

func TestChat_GetChatEphemeralTimer(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		timer, err := rpc.GetChatEphemeralTimer(accId, chatId)
		require.Nil(t, err)
		require.Equal(t, uint32(0), timer)

		require.Nil(t, rpc.SetChatEphemeralTimer(accId, chatId, 300))

		timer, err = rpc.GetChatEphemeralTimer(accId, chatId)
		require.Nil(t, err)
		require.Equal(t, uint32(300), timer)
	})
}

func TestChat_MarknoticedAllChats(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		require.Nil(t, rpc.MarknoticedAllChats(accId))
	})
}

func TestChat_GetSimilarChatIds(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		ids, err := rpc.GetSimilarChatIds(accId, chatId)
		require.Nil(t, err)
		require.NotNil(t, ids)
	})
}

func TestChat_CreateGroupChatUnencrypted(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		chatId, err := rpc.CreateGroupChatUnencrypted(accId, "unencrypted group")
		require.Nil(t, err)
		require.NotEqual(t, uint32(0), chatId)
	})
}

func TestChat_CanSend(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		canSend, err := rpc.CanSend(accId, chatId)
		require.Nil(t, err)
		require.True(t, canSend)
	})
}

func TestChat_AddRemoveContactFromChat(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		acfactory.WithOnlineAccount(func(rpc2 *Rpc, accId2 uint32) {
			contactId := acfactory.ImportContact(rpc, accId, rpc2, accId2)

			require.Nil(t, rpc.AddContactToChat(accId, chatId, contactId))

			contacts, err := rpc.GetChatContacts(accId, chatId)
			require.Nil(t, err)
			require.Contains(t, contacts, contactId)

			require.Nil(t, rpc.RemoveContactFromChat(accId, chatId, contactId))

			contacts, err = rpc.GetChatContacts(accId, chatId)
			require.Nil(t, err)
			require.NotContains(t, contacts, contactId)
		})
	})
}

func TestChat_GetPastChatContacts(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		acfactory.WithOnlineAccount(func(rpc2 *Rpc, accId2 uint32) {
			contactId := acfactory.ImportContact(rpc, accId, rpc2, accId2)
			require.Nil(t, rpc.AddContactToChat(accId, chatId, contactId))

			// send message so group gets promoted and past chat contacts is set
			_, err := rpc.SendMsg(accId, chatId, MessageData{Text: strptr("test")})
			require.Nil(t, err)

			require.Nil(t, rpc.RemoveContactFromChat(accId, chatId, contactId))
			past, err := rpc.GetPastChatContacts(accId, chatId)
			require.Nil(t, err)
			require.Contains(t, past, contactId)
		})
	})
}

func TestMsg_GetDraftAndRemoveDraft(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		draft, err := rpc.GetDraft(accId, chatId)
		require.Nil(t, err)
		require.Nil(t, draft)

		require.Nil(t, rpc.MiscSetDraft(accId, chatId, strptr("draft text"), nil, nil, nil, nil))

		draft, err = rpc.GetDraft(accId, chatId)
		require.Nil(t, err)
		require.NotNil(t, draft)

		require.Nil(t, rpc.RemoveDraft(accId, chatId))

		draft, err = rpc.GetDraft(accId, chatId)
		require.Nil(t, err)
		require.Nil(t, draft)
	})
}

func TestMsg_GetFreshMsgCnt(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		count, err := rpc.GetFreshMsgCnt(accId, chatId)
		require.Nil(t, err)
		require.Equal(t, uint(0), count)
	})
}

func TestMsg_GetMessages(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		msgId, err := rpc.MiscSendTextMessage(accId, chatId, "test message")
		require.Nil(t, err)

		msgs, err := rpc.GetMessages(accId, []uint32{msgId})
		require.Nil(t, err)
		require.NotEmpty(t, msgs)

		rpc.Transport.(*IOTransport).Close()
		_, err = rpc.GetMessages(accId, []uint32{msgId})
		require.NotNil(t, err)
	})
}

func TestMsg_GetMessageInfo(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		msgId, err := rpc.MiscSendTextMessage(accId, chatId, "test message")
		require.Nil(t, err)

		info, err := rpc.GetMessageInfo(accId, msgId)
		require.Nil(t, err)
		require.NotEmpty(t, info)

		infoObj, err := rpc.GetMessageInfoObject(accId, msgId)
		require.Nil(t, err)
		require.NotNil(t, infoObj)
	})
}

func TestMsg_GetMessageListItems(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		_, err := rpc.MiscSendTextMessage(accId, chatId, "test message")
		require.Nil(t, err)

		items, err := rpc.GetMessageListItems(accId, chatId, false, false)
		require.Nil(t, err)
		require.NotEmpty(t, items)

		rpc.Transport.(*IOTransport).Close()
		_, err = rpc.GetMessageListItems(accId, chatId, false, false)
		require.NotNil(t, err)
	})
}

func TestMsg_GetExistingMsgIds(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		msgId, err := rpc.MiscSendTextMessage(accId, chatId, "test message")
		require.Nil(t, err)

		ids, err := rpc.GetExistingMsgIds(accId, []uint32{msgId, 99999})
		require.Nil(t, err)
		require.Contains(t, ids, msgId)
		require.NotContains(t, ids, uint32(99999))
	})
}

func TestMsg_GetMessageHtml(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		msgId, err := rpc.SendMsg(accId, chatId, MessageData{Html: strptr("test")})
		require.Nil(t, err)

		html, err := rpc.GetMessageHtml(accId, msgId)
		require.Nil(t, err)
		require.NotNil(t, html)
		require.Equal(t, "test", *html)
	})
}

func TestMsg_GetMessageNotificationInfo(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		msgId, err := rpc.MiscSendTextMessage(accId, chatId, "notification test")
		require.Nil(t, err)

		info, err := rpc.GetMessageNotificationInfo(accId, msgId)
		require.Nil(t, err)
		require.NotNil(t, info)
	})
}

func TestMsg_MiscSendMsg(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		result, err := rpc.MiscSendMsg(accId, chatId, strptr("misc send test"), nil, nil, nil, nil)
		require.Nil(t, err)
		require.NotEqual(t, uint32(0), result.First)
		require.Equal(t, "misc send test", result.Second.Text)
	})
}

func TestMsg_ForwardMessages(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		chatId2, err := rpc.CreateGroupChat(accId, "target group", false)
		require.Nil(t, err)

		msgId, err := rpc.MiscSendTextMessage(accId, chatId, "forward me")
		require.Nil(t, err)

		require.Nil(t, rpc.ForwardMessages(accId, []uint32{msgId}, chatId2))

		ids, err := rpc.GetMessageIds(accId, chatId2, false, false)
		require.Nil(t, err)
		require.NotEmpty(t, ids)
	})
}

func TestMsg_SendEditRequest(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		msgId, err := rpc.MiscSendTextMessage(accId, chatId, "original text")
		require.Nil(t, err)

		require.Nil(t, rpc.SendEditRequest(accId, msgId, "edited text"))
	})
}

func TestMsg_MiscGetStickers(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		folder, err := rpc.MiscGetStickerFolder(accId)
		require.Nil(t, err)
		require.NotEmpty(t, folder)

		stickers, err := rpc.MiscGetStickers(accId)
		require.Nil(t, err)
		require.NotNil(t, stickers)
	})
}

func TestChat_CreateBroadcast(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		chatId, err := rpc.CreateBroadcast(accId, "test broadcast")
		require.Nil(t, err)
		require.NotEqual(t, uint32(0), chatId)
	})
}

func TestMsg_Webxdc(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		webxdcPath := acfactory.TestWebxdc()
		msgId, err := rpc.SendMsg(accId, chatId, MessageData{File: &webxdcPath})
		require.Nil(t, err)

		info, err := rpc.GetWebxdcInfo(accId, msgId)
		require.Nil(t, err)
		require.Equal(t, "TestApp", info.Name)

		updates, err := rpc.GetWebxdcStatusUpdates(accId, msgId, 0)
		require.Nil(t, err)
		require.NotEmpty(t, updates)

		require.Nil(t, rpc.SendWebxdcStatusUpdate(accId, msgId, `{"payload":"test"}`, nil))
	})
}

func TestRpc_Sleep(t *testing.T) {
	t.Parallel()
	acfactory.WithRpc(func(rpc *Rpc) {
		require.Nil(t, rpc.Sleep(0))
	})
}

func TestAccount_GetPushState(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		state, err := rpc.GetPushState(accId)
		require.Nil(t, err)
		require.NotEmpty(t, string(state))
	})
}

func TestChat_LeaveGroup(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		_, err := rpc.MiscSendTextMessage(accId, chatId, "promote group")
		require.Nil(t, err)

		require.Nil(t, rpc.LeaveGroup(accId, chatId))
	})
}

func TestRpc_GetNextEventBatch(t *testing.T) {
	t.Parallel()
	acfactory.WithRpc(func(rpc *Rpc) {
		// GetNextEventBatch blocks until an event arrives; use a short-lived context
		// to exercise the code path without hanging indefinitely.
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		timedRpc := &Rpc{Context: ctx, Transport: rpc.Transport}
		_, _ = timedRpc.GetNextEventBatch() // ignore timeout error
	})
}

func TestRpc_MigrateAccount(t *testing.T) {
	t.Parallel()
	acfactory.WithRpc(func(rpc *Rpc) {
		_, err := rpc.MigrateAccount("/nonexistent/path/to.db")
		require.NotNil(t, err)
	})
}

func TestRpc_BackgroundFetch(t *testing.T) {
	t.Parallel()
	acfactory.WithRpc(func(rpc *Rpc) {
		require.Nil(t, rpc.BackgroundFetch(0))
	})
}

func TestRpc_StopBackgroundFetch(t *testing.T) {
	t.Parallel()
	acfactory.WithRpc(func(rpc *Rpc) {
		require.Nil(t, rpc.StopBackgroundFetch())
	})
}

func TestRpc_CopyToBlobDir(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		dest, err := rpc.CopyToBlobDir(accId, acfactory.TestFile("test.txt", 1))
		require.Nil(t, err)
		require.NotEmpty(t, dest)
	})
}

func TestRpc_CheckQr(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		qr, err := rpc.CheckQr(accId, "https://example.com")
		require.Nil(t, err)
		require.NotNil(t, qr)

		rpc.Transport.(*IOTransport).Close()
		_, err = rpc.CheckQr(accId, "")
		require.NotNil(t, err)
	})
}

func TestRpc_SetStockStrings(t *testing.T) {
	t.Parallel()
	acfactory.WithRpc(func(rpc *Rpc) {
		require.Nil(t, rpc.SetStockStrings(map[string]string{}))
	})
}

func TestRpc_AddOrUpdateTransport(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		transports, err := rpc.ListTransports(accId)
		require.Nil(t, err)
		require.NotEmpty(t, transports)
		require.Nil(t, rpc.AddOrUpdateTransport(accId, transports[0]))
	})
}

func TestRpc_AddTransport(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		transports, err := rpc.ListTransports(accId)
		require.Nil(t, err)
		require.NotEmpty(t, transports)
		require.Nil(t, rpc.AddTransport(accId, transports[0]))
	})
}

func TestRpc_ListTransports(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		transports, err := rpc.ListTransports(accId)
		require.Nil(t, err)
		require.NotEmpty(t, transports)
	})
}

func TestRpc_DeleteTransport(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		// First add a second transport, you can't have zero transports
		if err := rpc.AddTransportFromQr(accId, acfactory.ConfigQr); err != nil {
			panic(err)
		}
		transports, err := rpc.ListTransports(accId)
		require.Nil(t, err)
		require.Equal(t, 2, len(transports))
		addr := transports[1].Addr
		require.Nil(t, rpc.DeleteTransport(accId, addr))
		transports, err = rpc.ListTransports(accId)
		require.Nil(t, err)
		require.Equal(t, 1, len(transports))
	})
}

func TestRpc_StopOngoingProcess(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		require.Nil(t, rpc.StopOngoingProcess(accId))
	})
}

func TestRpc_ExportAndImportSelfKeys(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		dir := acfactory.MkdirTemp()
		require.Nil(t, rpc.ExportSelfKeys(accId, dir, nil))
		// FIXME: this shouldn't throw error, see https://github.com/chatmail/core/issues/7960
		require.NotNil(t, rpc.ImportSelfKeys(accId, dir, nil))
	})
}

func TestRpc_WaitNextMsgs(t *testing.T) {
	t.Parallel()
	acfactory.WithRunningBot(func(bot *Bot, botAcc uint32) {
		acfactory.WithOnlineAccount(func(accRpc *Rpc, accId uint32) {
			chatWithBot := acfactory.CreateChat(accRpc, accId, bot.Rpc, botAcc)
			_, err := accRpc.MiscSendTextMessage(accId, chatWithBot, "test")
			require.Nil(t, err)

			msgs, err := bot.Rpc.WaitNextMsgs(botAcc)
			require.Nil(t, err)
			require.NotNil(t, msgs)
		})
	})
}

func TestRpc_SecureJoinWithUxInfo(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc1 *Rpc, accId1 uint32) {
		acfactory.WithOnlineAccount(func(rpc2 *Rpc, accId2 uint32) {
			qrdata, err := rpc1.GetChatSecurejoinQrCode(accId1, nil)
			require.Nil(t, err)

			_, err = rpc2.SecureJoinWithUxInfo(accId2, qrdata, nil, nil)
			require.Nil(t, err)
			acfactory.WaitForEvent(rpc1, accId1, &EventTypeSecurejoinInviterProgress{})
			acfactory.WaitForEvent(rpc2, accId2, &EventTypeSecurejoinJoinerProgress{})
		})
	})
}

func TestRpc_DeleteMessagesForAll(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		msgId, err := rpc.MiscSendTextMessage(accId, chatId, "test")
		require.Nil(t, err)
		require.Nil(t, rpc.DeleteMessagesForAll(accId, []uint32{msgId}))
	})
}

func TestRpc_GetMessageReadReceiptCountAndReceipts(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		msgId, err := rpc.MiscSendTextMessage(accId, chatId, "check read receipts")
		require.Nil(t, err)

		count, err := rpc.GetMessageReadReceiptCount(accId, msgId)
		require.Nil(t, err)
		require.Equal(t, uint(0), count)

		receipts, err := rpc.GetMessageReadReceipts(accId, msgId)
		require.Nil(t, err)
		require.NotNil(t, receipts)
	})
}

func TestRpc_DownloadFullMessage(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc1 *Rpc, accId1 uint32) {
		acfactory.WithOnlineAccount(func(rpc2 *Rpc, accId2 uint32) {
			// Set a tiny download limit so that attachment are not auto-downloaded
			require.Nil(t, rpc1.SetConfig(accId1, "download_limit", strptr("1")))

			chatId := acfactory.CreateChat(rpc2, accId2, rpc1, accId1)
			file := acfactory.TestFile("test.txt", 500*1024) // if not big enough a pre-message will not be sent
			_, err := rpc2.SendMsg(accId2, chatId, MessageData{File: &file})
			require.Nil(t, err)

			msg := acfactory.NextMsg(rpc1, accId1)
			require.Nil(t, rpc1.DownloadFullMessage(accId1, msg.Id))
		})
	})
}

func TestRpc_MessageIdsToSearchResults(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		msgId, err := rpc.MiscSendTextMessage(accId, chatId, "searchable message")
		require.Nil(t, err)
		results, err := rpc.MessageIdsToSearchResults(accId, []uint32{msgId})
		require.Nil(t, err)
		require.NotNil(t, results)
	})
}

func TestRpc_SaveMsgs(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		msgId, err := rpc.MiscSendTextMessage(accId, chatId, "save me")
		require.Nil(t, err)
		require.Nil(t, rpc.SaveMsgs(accId, []uint32{msgId}))
	})
}

func TestRpc_ParseVcard(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		vcard, err := rpc.MakeVcard(accId, []uint32{ContactSelf})
		require.Nil(t, err)

		dir := acfactory.MkdirTemp()
		path := filepath.Join(dir, "contact.vcf")
		require.Nil(t, os.WriteFile(path, []byte(vcard), 0600))

		contacts, err := rpc.ParseVcard(path)
		require.Nil(t, err)
		require.NotEmpty(t, contacts)
	})
}

func TestRpc_ImportVcard(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc1 *Rpc, accId1 uint32) {
		vcard, err := rpc1.MakeVcard(accId1, []uint32{ContactSelf})
		require.Nil(t, err)

		acfactory.WithOnlineAccount(func(rpc2 *Rpc, accId2 uint32) {
			dir := acfactory.MkdirTemp()
			path := filepath.Join(dir, "contact.vcf")
			require.Nil(t, os.WriteFile(path, []byte(vcard), 0600))

			ids, err := rpc2.ImportVcard(accId2, path)
			require.Nil(t, err)
			require.NotEmpty(t, ids)
		})
	})
}

func TestRpc_SetDraftVcard(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		// FIXME: this shouldn't throw error, see https://github.com/chatmail/core/issues/7960
		err := rpc.SetDraftVcard(accId, chatId, []uint32{ContactSelf})
		require.NotNil(t, err)
		require.Equal(t, "Wrong viewtype for vCard: Text", err.(*jrpc2.Error).Message)
	})
}

func TestRpc_GetChatMedia(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		msgId, err := rpc.SendMsg(accId, chatId, MessageData{File: strptr(acfactory.TestImage())})
		require.Nil(t, err)

		msgs, err := rpc.GetChatMedia(accId, &chatId, ViewtypeImage, nil, nil)
		require.Nil(t, err)
		require.NotNil(t, msgs)
		require.Contains(t, msgs, msgId)
	})
}

func TestRpc_GetLocations(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		locs, err := rpc.GetLocations(accId, &chatId, nil, 0, 0)
		require.Nil(t, err)
		require.NotNil(t, locs)
	})
}

func TestRpc_SendWebxdcRealtimeDataAndAdvertisement(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc1 *Rpc, accId1 uint32) {
		acfactory.WithOnlineAccount(func(rpc2 *Rpc, accId2 uint32) {
			chatId := acfactory.CreateChat(rpc1, accId1, rpc2, accId2)
			msgId, err := rpc1.SendMsg(accId1, chatId, MessageData{File: strptr(acfactory.TestWebxdc())})
			require.Nil(t, err)

			require.Nil(t, rpc1.SendWebxdcRealtimeAdvertisement(accId1, msgId))
			require.Nil(t, rpc1.SendWebxdcRealtimeData(accId1, msgId, []int{1, 2, 3}))
			require.Nil(t, rpc1.LeaveWebxdcRealtime(accId1, msgId))
		})
	})
}

func TestRpc_GetWebxdcHrefAndBlob(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		msgId, err := rpc.SendMsg(accId, chatId, MessageData{File: strptr(acfactory.TestWebxdc())})
		require.Nil(t, err)

		// this is supposed to be used on info-messages, not normal messages
		_, err = rpc.GetWebxdcHref(accId, msgId)
		require.Nil(t, err)

		_, err = rpc.GetWebxdcBlob(accId, msgId, "index.html")
		require.Nil(t, err)
	})
}

func TestRpc_SetAndInitWebxdcIntegration(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		require.Nil(t, rpc.SetWebxdcIntegration(accId, acfactory.TestWebxdc()))
		_, err := rpc.InitWebxdcIntegration(accId, &chatId)
		require.Nil(t, err)
	})
}

func TestRpc_IceServers(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		_, err := rpc.IceServers(accId)
		require.Nil(t, err)
	})
}

func TestRpc_ForwardMessagesToAccount(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId1 uint32, chatId1 uint32) {
		msgId, err := rpc.MiscSendTextMessage(accId1, chatId1, "to forward")
		require.Nil(t, err)

		// add a second profile
		accId2, err := rpc.AddAccount()
		require.Nil(t, err)
		if err := rpc.AddTransportFromQr(accId2, acfactory.ConfigQr); err != nil {
			panic(err)
		}

		// create a chat in second profile to forward messages into it
		chatId2, err := rpc.CreateChatByContactId(accId2, ContactSelf)
		require.Nil(t, err)

		require.Nil(t, rpc.ForwardMessagesToAccount(accId1, []uint32{msgId}, accId2, chatId2))
	})
}

func TestRpc_ResendMessages(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		msgId, err := rpc.MiscSendTextMessage(accId, chatId, "resend me")
		require.Nil(t, err)
		require.Nil(t, rpc.ResendMessages(accId, []uint32{msgId}))
	})
}

func TestRpc_SendSticker(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		_, err := rpc.SendSticker(accId, chatId, acfactory.TestFile("test.webp", 1))
		require.Nil(t, err)
	})
}

func TestRpc_SaveMsgFile(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		msgId, err := rpc.SendMsg(accId, chatId, MessageData{File: strptr(acfactory.TestImage())})
		require.Nil(t, err)
		destPath := filepath.Join(acfactory.MkdirTemp(), "saved.jpg")
		require.Nil(t, rpc.SaveMsgFile(accId, msgId, destPath))
	})
}

func TestRpc_MiscSaveSticker(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		msgId, err := rpc.SendSticker(accId, chatId, acfactory.TestFile("test.webp", 1))
		require.Nil(t, err)
		require.Nil(t, rpc.MiscSaveSticker(accId, msgId, "Saved"))
	})
}

func TestRpc_GetHttpResponse(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		resp, err := rpc.GetHttpResponse(accId, "https://delta.chat/robots.txt")
		require.Nil(t, err)
		require.NotEmpty(t, resp.Blob)
	})
}

func TestRpc_Calls(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc1 *Rpc, accId1 uint32) {
		acfactory.WithOnlineAccount(func(rpc2 *Rpc, accId2 uint32) {
			// calls will not trigger for contact request, introduce each other
			chatId1, _ := acfactory.IntroduceEachOther(rpc1, accId1, rpc2, accId2)

			// start call
			msgId1, err := rpc1.PlaceOutgoingCall(accId1, chatId1, "fake-data", false)
			require.Nil(t, err)
			info1, err := rpc1.CallInfo(accId1, msgId1)
			require.Nil(t, err)
			require.NotNil(t, "Alerting", info1.State.GetKind())

			// wait for the incoming call on the other side
			event2 := acfactory.WaitForEvent(rpc2, accId2, &EventTypeIncomingCall{}).(*EventTypeIncomingCall)
			info2, err := rpc2.CallInfo(accId2, event2.MsgId)
			require.Nil(t, err)
			require.NotNil(t, "Alerting", info2.State.GetKind())

			// accept incoming call
			require.Nil(t, rpc2.AcceptIncomingCall(accId2, event2.MsgId, "fake-data"))
			info2, err = rpc2.CallInfo(accId2, event2.MsgId)
			require.Nil(t, err)
			require.NotNil(t, "Active", info2.State.GetKind())

			// wait for call response
			require.NotNil(t, acfactory.WaitForEvent(rpc1, accId1, &EventTypeOutgoingCallAccepted{}).(*EventTypeOutgoingCallAccepted))
			info1, err = rpc1.CallInfo(accId1, msgId1)
			require.Nil(t, err)
			require.NotNil(t, "Active", info1.State.GetKind())

			// end call
			require.Nil(t, rpc1.EndCall(accId1, msgId1))
			info1, err = rpc1.CallInfo(accId1, msgId1)
			require.Nil(t, err)
			require.NotNil(t, "Completed", info1.State.GetKind())
		})
	})
}
