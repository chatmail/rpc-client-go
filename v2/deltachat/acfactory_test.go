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
