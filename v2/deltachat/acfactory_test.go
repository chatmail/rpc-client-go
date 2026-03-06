package deltachat

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAcFactory_TearDown(t *testing.T) {
	t.Parallel()
	acf := &AcFactory{}
	acf.TearUp()
	acf.TearDown()
}

func TestAcFactory_getChatId(t *testing.T) {
	t.Parallel()
	getChatId(&EventTypeMsgsChanged{})
	getChatId(&EventTypeReactionsChanged{})
	getChatId(&EventTypeIncomingMsg{})
	getChatId(&EventTypeMsgsNoticed{})
	getChatId(&EventTypeMsgDelivered{})
	getChatId(&EventTypeMsgFailed{})
	getChatId(&EventTypeMsgRead{})
	getChatId(&EventTypeMsgDeleted{})
	getChatId(&EventTypeChatModified{})
	getChatId(&EventTypeChatEphemeralTimerModified{})
	require.Equal(t, uint32(0), getChatId(&EventTypeInfo{})) // default case: returns 0
}

func TestAcFactory_IntroduceEachOther(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc1 *Rpc, accId1 uint32) {
		acfactory.WithOnlineAccount(func(rpc2 *Rpc, accId2 uint32) {
			acfactory.IntroduceEachOther(rpc1, accId1, rpc2, accId2)
		})
	})
}

func TestAcFactory_WaitForEventInChat(t *testing.T) {
	t.Parallel()
	acfactory.WithGroup(func(rpc *Rpc, accId uint32, chatId uint32) {
		_, err := rpc.MiscSendTextMessage(accId, chatId, "test")
		require.Nil(t, err)
		require.NotNil(t, acfactory.WaitForEventInChat(rpc, accId, chatId, &EventTypeMsgsChanged{}))
	})
}

func TestAcFactory_WaitForEventDebugPath(t *testing.T) {
	t.Parallel()
	// Enable debug mode to exercise the debug print path in WaitForEvent.
	acf := &AcFactory{Debug: true}
	acf.TearUp()
	defer acf.TearDown()

	acf.WithOnlineAccount(func(rpc *Rpc, accId uint32) {
		acf.WithOnlineAccount(func(rpc2 *Rpc, accId2 uint32) {
			chatId := acf.CreateChat(rpc2, accId2, rpc, accId)
			_, err := rpc2.MiscSendTextMessage(accId2, chatId, "trigger debug event")
			require.Nil(t, err)
			// Waiting for an event in accId will print debug info for every event received.
			acf.WaitForEvent(rpc, accId, &EventTypeIncomingMsg{})
		})
	})
}
