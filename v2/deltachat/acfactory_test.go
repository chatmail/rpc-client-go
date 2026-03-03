package deltachat

import (
	"testing"
)

func TestAcFactory_TearDown(t *testing.T) {
	t.Parallel()
	acf := &AcFactory{}
	acf.TearUp()
	acf.TearDown()
}

func TestAcFactory_getChatId(t *testing.T) {
	t.Parallel()
	getChatId(&EventTypeIncomingMsg{})
	getChatId(&EventTypeMsgsNoticed{})
	getChatId(&EventTypeMsgDelivered{})
	getChatId(&EventTypeMsgFailed{})
	getChatId(&EventTypeMsgRead{})
	getChatId(&EventTypeChatModified{})
}
