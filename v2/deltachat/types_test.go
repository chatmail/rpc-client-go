package deltachat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPair_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	var p Pair[string, int]
	require.Nil(t, json.Unmarshal([]byte(`["hello", 42]`), &p))
	require.Equal(t, "hello", p.First)
	require.Equal(t, 42, p.Second)

	require.NotNil(t, json.Unmarshal([]byte(`"notarray"`), &p))
	require.NotNil(t, json.Unmarshal([]byte(`[1, 2]`), &p))
	require.NotNil(t, json.Unmarshal([]byte(`["hello", "notint"]`), &p))
}

func TestUnmarshalAccount(t *testing.T) {
	t.Parallel()

	var acc Account
	require.Nil(t, unmarshalAccount(json.RawMessage(`{"kind":"Configured","id":1,"color":"#fff"}`), &acc))
	require.Equal(t, "Configured", acc.GetKind())

	require.Nil(t, unmarshalAccount(json.RawMessage(`{"kind":"Unconfigured","id":2}`), &acc))
	require.Equal(t, "Unconfigured", acc.GetKind())

	require.NotNil(t, unmarshalAccount(json.RawMessage(`{"kind":"Unknown"}`), &acc))
	require.NotNil(t, unmarshalAccount(json.RawMessage(`notjson`), &acc))
}

func TestAccount_MarshalJSON(t *testing.T) {
	t.Parallel()

	configured := &AccountConfigured{Id: 1, Color: "#fff"}
	data, err := json.Marshal(configured)
	require.Nil(t, err)
	require.Contains(t, string(data), `"kind":"Configured"`)
	require.Contains(t, string(data), `"id":1`)

	unconfigured := &AccountUnconfigured{Id: 2}
	data, err = json.Marshal(unconfigured)
	require.Nil(t, err)
	require.Contains(t, string(data), `"kind":"Unconfigured"`)
	require.Contains(t, string(data), `"id":2`)
}

func TestCallInfo_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	variants := []struct {
		kind    string
		payload string
	}{
		{"Alerting", `{"kind":"Alerting"}`},
		{"Active", `{"kind":"Active"}`},
		{"Completed", `{"kind":"Completed","duration":42}`},
		{"Missed", `{"kind":"Missed"}`},
		{"Declined", `{"kind":"Declined"}`},
		{"Canceled", `{"kind":"Canceled"}`},
	}
	for _, v := range variants {
		var ci CallInfo
		err := json.Unmarshal([]byte(`{"hasVideo":true,"sdpOffer":"offer","state":`+v.payload+`}`), &ci)
		require.Nil(t, err, "variant %s", v.kind)
		require.Equal(t, v.kind, ci.State.GetKind(), "variant %s", v.kind)
	}

	var ci CallInfo
	require.NotNil(t, json.Unmarshal([]byte(`notjson`), &ci))
	require.NotNil(t, json.Unmarshal([]byte(`{"state":{"kind":"Unknown"}}`), &ci))
}

func TestCallState_MarshalJSON(t *testing.T) {
	t.Parallel()

	states := []CallState{
		&CallStateAlerting{},
		&CallStateActive{},
		&CallStateCompleted{Duration: 10},
		&CallStateMissed{},
		&CallStateDeclined{},
		&CallStateCanceled{},
	}
	kinds := []string{"Alerting", "Active", "Completed", "Missed", "Declined", "Canceled"}

	for i, state := range states {
		data, err := json.Marshal(state)
		require.Nil(t, err)
		require.Contains(t, string(data), `"kind":"`+kinds[i]+`"`)
	}
}

func TestUnmarshalCallState(t *testing.T) {
	t.Parallel()

	variants := map[string]string{
		"Alerting":  `{"kind":"Alerting"}`,
		"Active":    `{"kind":"Active"}`,
		"Completed": `{"kind":"Completed","duration":5}`,
		"Missed":    `{"kind":"Missed"}`,
		"Declined":  `{"kind":"Declined"}`,
		"Canceled":  `{"kind":"Canceled"}`,
	}
	for kind, payload := range variants {
		var out CallState
		err := unmarshalCallState(json.RawMessage(payload), &out)
		require.Nil(t, err, "variant %s", kind)
		require.Equal(t, kind, out.GetKind())
	}

	var out CallState
	require.NotNil(t, unmarshalCallState(json.RawMessage(`notjson`), &out))
	require.NotNil(t, unmarshalCallState(json.RawMessage(`{"kind":"Unknown"}`), &out))
}

func TestChatListItemFetchResult_MarshalJSON(t *testing.T) {
	t.Parallel()

	item := &ChatListItemFetchResultChatListItem{Id: 1, Name: "test"}
	data, err := json.Marshal(item)
	require.Nil(t, err)
	require.Contains(t, string(data), `"kind":"ChatListItem"`)

	archive := &ChatListItemFetchResultArchiveLink{FreshMessageCounter: 3}
	data, err = json.Marshal(archive)
	require.Nil(t, err)
	require.Contains(t, string(data), `"kind":"ArchiveLink"`)

	errItem := &ChatListItemFetchResultError{Id: 2, Error: "err"}
	data, err = json.Marshal(errItem)
	require.Nil(t, err)
	require.Contains(t, string(data), `"kind":"Error"`)
}

func TestUnmarshalChatListItemFetchResult(t *testing.T) {
	t.Parallel()

	var out ChatListItemFetchResult
	require.Nil(t, unmarshalChatListItemFetchResult(json.RawMessage(`{"kind":"ChatListItem","id":1,"name":"test","color":"#fff","summaryStatus":0,"summaryText1":"","summaryText2":"","timestamp":0,"isProtected":false,"isContactRequest":false,"isSelfTalk":false,"isDeviceChat":false,"isMuted":false,"isArchived":false,"archived":false,"pinned":false,"wasSeenRecently":false}`), &out))
	require.Equal(t, "ChatListItem", out.GetKind())

	require.Nil(t, unmarshalChatListItemFetchResult(json.RawMessage(`{"kind":"ArchiveLink","freshMessageCounter":1}`), &out))
	require.Equal(t, "ArchiveLink", out.GetKind())

	require.Nil(t, unmarshalChatListItemFetchResult(json.RawMessage(`{"kind":"Error","id":1,"error":"oops"}`), &out))
	require.Equal(t, "Error", out.GetKind())

	require.NotNil(t, unmarshalChatListItemFetchResult(json.RawMessage(`notjson`), &out))
	require.NotNil(t, unmarshalChatListItemFetchResult(json.RawMessage(`{"kind":"Unknown"}`), &out))
}

func TestEphemeralTimer_MarshalJSON(t *testing.T) {
	t.Parallel()

	disabled := &EphemeralTimerDisabled{}
	data, err := json.Marshal(disabled)
	require.Nil(t, err)
	require.Contains(t, string(data), `"kind":"disabled"`)

	enabled := &EphemeralTimerEnabled{Duration: 300}
	data, err = json.Marshal(enabled)
	require.Nil(t, err)
	require.Contains(t, string(data), `"kind":"enabled"`)
	require.Contains(t, string(data), `"duration":300`)
}

func TestUnmarshalEphemeralTimer(t *testing.T) {
	t.Parallel()

	var out EphemeralTimer
	require.Nil(t, unmarshalEphemeralTimer(json.RawMessage(`{"kind":"disabled"}`), &out))
	require.Equal(t, "disabled", out.GetKind())

	require.Nil(t, unmarshalEphemeralTimer(json.RawMessage(`{"kind":"enabled","duration":60}`), &out))
	require.Equal(t, "enabled", out.GetKind())

	require.NotNil(t, unmarshalEphemeralTimer(json.RawMessage(`notjson`), &out))
	require.NotNil(t, unmarshalEphemeralTimer(json.RawMessage(`{"kind":"Unknown"}`), &out))
}

func TestEventType_MarshalJSON(t *testing.T) {
	t.Parallel()

	events := []EventType{
		&EventTypeInfo{Msg: "info"},
		&EventTypeSmtpConnected{Msg: "smtp"},
		&EventTypeImapConnected{Msg: "imap"},
		&EventTypeSmtpMessageSent{Msg: "sent"},
		&EventTypeImapMessageDeleted{Msg: "deleted"},
		&EventTypeImapMessageMoved{Msg: "moved"},
		&EventTypeImapInboxIdle{},
		&EventTypeNewBlobFile{File: "file.jpg"},
		&EventTypeDeletedBlobFile{File: "file.jpg"},
		&EventTypeWarning{Msg: "warn"},
		&EventTypeError{Msg: "error"},
		&EventTypeErrorSelfNotInGroup{Msg: "not in group"},
		&EventTypeMsgsChanged{ChatId: 1, MsgId: 2},
		&EventTypeReactionsChanged{ChatId: 1, ContactId: 2, MsgId: 3},
		&EventTypeIncomingReaction{ChatId: 1, ContactId: 2, MsgId: 3, Reaction: "👍"},
		&EventTypeIncomingWebxdcNotify{ChatId: 1, ContactId: 2, MsgId: 3, Text: "update"},
		&EventTypeIncomingMsg{ChatId: 1, MsgId: 2},
		&EventTypeIncomingMsgBunch{},
		&EventTypeMsgsNoticed{ChatId: 1},
		&EventTypeMsgDelivered{ChatId: 1, MsgId: 2},
		&EventTypeMsgFailed{ChatId: 1, MsgId: 2},
		&EventTypeMsgRead{ChatId: 1, MsgId: 2},
		&EventTypeMsgDeleted{ChatId: 1, MsgId: 2},
		&EventTypeChatModified{ChatId: 1},
		&EventTypeChatEphemeralTimerModified{ChatId: 1, Timer: 60},
		&EventTypeChatDeleted{ChatId: 1},
		&EventTypeContactsChanged{},
		&EventTypeLocationChanged{},
		&EventTypeConfigureProgress{Progress: 500},
		&EventTypeImexProgress{Progress: 500},
		&EventTypeImexFileWritten{Path: "/tmp/keys.zip"},
		&EventTypeSecurejoinInviterProgress{ChatId: 1, ContactId: 2, Progress: 1000},
		&EventTypeSecurejoinJoinerProgress{ContactId: 1, Progress: 1000},
		&EventTypeConnectivityChanged{},
		&EventTypeSelfavatarChanged{},
		&EventTypeConfigSynced{Key: "selfstatus"},
		&EventTypeWebxdcStatusUpdate{MsgId: 1, StatusUpdateSerial: 2},
		&EventTypeWebxdcRealtimeData{MsgId: 1, Data: []int{1, 2, 3}},
		&EventTypeWebxdcRealtimeAdvertisementReceived{MsgId: 1},
		&EventTypeWebxdcInstanceDeleted{MsgId: 1},
		&EventTypeAccountsBackgroundFetchDone{},
		&EventTypeChatlistChanged{},
		&EventTypeChatlistItemChanged{},
		&EventTypeAccountsChanged{},
		&EventTypeAccountsItemChanged{},
		&EventTypeEventChannelOverflow{N: 5},
		&EventTypeIncomingCall{ChatId: 1, MsgId: 2, HasVideo: true},
		&EventTypeIncomingCallAccepted{ChatId: 1, MsgId: 2},
		&EventTypeOutgoingCallAccepted{ChatId: 1, MsgId: 2, AcceptCallInfo: "info"},
		&EventTypeCallEnded{ChatId: 1, MsgId: 2},
		&EventTypeTransportsModified{},
	}

	for _, event := range events {
		data, err := json.Marshal(event)
		require.Nil(t, err, "event %T", event)
		require.Contains(t, string(data), `"kind":"`+event.GetKind()+`"`, "event %T", event)
	}
}

func TestUnmarshalEventType(t *testing.T) {
	t.Parallel()

	variants := []struct {
		kind    string
		payload string
	}{
		{"Info", `{"kind":"Info","msg":"hello"}`},
		{"SmtpConnected", `{"kind":"SmtpConnected","msg":"ok"}`},
		{"ImapConnected", `{"kind":"ImapConnected","msg":"ok"}`},
		{"SmtpMessageSent", `{"kind":"SmtpMessageSent","msg":"ok"}`},
		{"ImapMessageDeleted", `{"kind":"ImapMessageDeleted","msg":"ok"}`},
		{"ImapMessageMoved", `{"kind":"ImapMessageMoved","msg":"ok"}`},
		{"ImapInboxIdle", `{"kind":"ImapInboxIdle"}`},
		{"NewBlobFile", `{"kind":"NewBlobFile","file":"x.jpg"}`},
		{"DeletedBlobFile", `{"kind":"DeletedBlobFile","file":"x.jpg"}`},
		{"Warning", `{"kind":"Warning","msg":"warn"}`},
		{"Error", `{"kind":"Error","msg":"err"}`},
		{"ErrorSelfNotInGroup", `{"kind":"ErrorSelfNotInGroup","msg":"not in group"}`},
		{"MsgsChanged", `{"kind":"MsgsChanged","chatId":1,"msgId":2}`},
		{"ReactionsChanged", `{"kind":"ReactionsChanged","chatId":1,"contactId":2,"msgId":3}`},
		{"IncomingReaction", `{"kind":"IncomingReaction","chatId":1,"contactId":2,"msgId":3,"reaction":"👍"}`},
		{"IncomingWebxdcNotify", `{"kind":"IncomingWebxdcNotify","chatId":1,"contactId":2,"msgId":3,"text":"update"}`},
		{"IncomingMsg", `{"kind":"IncomingMsg","chatId":1,"msgId":2}`},
		{"IncomingMsgBunch", `{"kind":"IncomingMsgBunch"}`},
		{"MsgsNoticed", `{"kind":"MsgsNoticed","chatId":1}`},
		{"MsgDelivered", `{"kind":"MsgDelivered","chatId":1,"msgId":2}`},
		{"MsgFailed", `{"kind":"MsgFailed","chatId":1,"msgId":2}`},
		{"MsgRead", `{"kind":"MsgRead","chatId":1,"msgId":2}`},
		{"MsgDeleted", `{"kind":"MsgDeleted","chatId":1,"msgId":2}`},
		{"ChatModified", `{"kind":"ChatModified","chatId":1}`},
		{"ChatEphemeralTimerModified", `{"kind":"ChatEphemeralTimerModified","chatId":1,"timer":60}`},
		{"ChatDeleted", `{"kind":"ChatDeleted","chat_id":1}`},
		{"ContactsChanged", `{"kind":"ContactsChanged"}`},
		{"LocationChanged", `{"kind":"LocationChanged"}`},
		{"ConfigureProgress", `{"kind":"ConfigureProgress","progress":500}`},
		{"ImexProgress", `{"kind":"ImexProgress","progress":500}`},
		{"ImexFileWritten", `{"kind":"ImexFileWritten","path":"/tmp/keys.zip"}`},
		{"SecurejoinInviterProgress", `{"kind":"SecurejoinInviterProgress","chatId":1,"chatType":"Single","contactId":2,"progress":1000}`},
		{"SecurejoinJoinerProgress", `{"kind":"SecurejoinJoinerProgress","contactId":1,"progress":1000}`},
		{"ConnectivityChanged", `{"kind":"ConnectivityChanged"}`},
		{"SelfavatarChanged", `{"kind":"SelfavatarChanged"}`},
		{"ConfigSynced", `{"kind":"ConfigSynced","key":"selfstatus"}`},
		{"WebxdcStatusUpdate", `{"kind":"WebxdcStatusUpdate","msgId":1,"statusUpdateSerial":2}`},
		{"WebxdcRealtimeData", `{"kind":"WebxdcRealtimeData","data":[1,2],"msgId":1}`},
		{"WebxdcRealtimeAdvertisementReceived", `{"kind":"WebxdcRealtimeAdvertisementReceived","msgId":1}`},
		{"WebxdcInstanceDeleted", `{"kind":"WebxdcInstanceDeleted","msgId":1}`},
		{"AccountsBackgroundFetchDone", `{"kind":"AccountsBackgroundFetchDone"}`},
		{"ChatlistChanged", `{"kind":"ChatlistChanged"}`},
		{"ChatlistItemChanged", `{"kind":"ChatlistItemChanged"}`},
		{"AccountsChanged", `{"kind":"AccountsChanged"}`},
		{"AccountsItemChanged", `{"kind":"AccountsItemChanged"}`},
		{"EventChannelOverflow", `{"kind":"EventChannelOverflow","n":5}`},
		{"IncomingCall", `{"kind":"IncomingCall","chat_id":1,"msg_id":2,"has_video":true,"place_call_info":"info"}`},
		{"IncomingCallAccepted", `{"kind":"IncomingCallAccepted","chat_id":1,"msg_id":2}`},
		{"OutgoingCallAccepted", `{"kind":"OutgoingCallAccepted","accept_call_info":"info","chat_id":1,"msg_id":2}`},
		{"CallEnded", `{"kind":"CallEnded","chat_id":1,"msg_id":2}`},
		{"TransportsModified", `{"kind":"TransportsModified"}`},
	}

	for _, v := range variants {
		var out EventType
		err := unmarshalEventType(json.RawMessage(v.payload), &out)
		require.Nil(t, err, "variant %s", v.kind)
		require.Equal(t, v.kind, out.GetKind(), "variant %s", v.kind)
	}

	var out EventType
	require.NotNil(t, unmarshalEventType(json.RawMessage(`notjson`), &out))
	require.NotNil(t, unmarshalEventType(json.RawMessage(`{"kind":"Unknown"}`), &out))
}

func TestMessageListItem_MarshalJSON(t *testing.T) {
	t.Parallel()

	msg := &MessageListItemMessage{MsgId: 42}
	data, err := json.Marshal(msg)
	require.Nil(t, err)
	require.Contains(t, string(data), `"kind":"message"`)

	day := &MessageListItemDayMarker{Timestamp: 1234567890000}
	data, err = json.Marshal(day)
	require.Nil(t, err)
	require.Contains(t, string(data), `"kind":"dayMarker"`)
}

func TestUnmarshalMessageListItem(t *testing.T) {
	t.Parallel()

	var out MessageListItem
	require.Nil(t, unmarshalMessageListItem(json.RawMessage(`{"kind":"message","msg_id":42}`), &out))
	require.Equal(t, "message", out.GetKind())

	require.Nil(t, unmarshalMessageListItem(json.RawMessage(`{"kind":"dayMarker","timestamp":1234}`), &out))
	require.Equal(t, "dayMarker", out.GetKind())

	require.NotNil(t, unmarshalMessageListItem(json.RawMessage(`notjson`), &out))
	require.NotNil(t, unmarshalMessageListItem(json.RawMessage(`{"kind":"Unknown"}`), &out))
}

func TestMessageLoadResult_MarshalJSON(t *testing.T) {
	t.Parallel()

	errResult := &MessageLoadResultLoadingError{Error: "something went wrong"}
	data, err := json.Marshal(errResult)
	require.Nil(t, err)
	require.Contains(t, string(data), `"kind":"loadingError"`)

	msgResult := &MessageLoadResultMessage{Id: 1, Text: "hello"}
	data, err = json.Marshal(msgResult)
	require.Nil(t, err)
	require.Contains(t, string(data), `"kind":"message"`)
}

func TestUnmarshalMessageLoadResult(t *testing.T) {
	t.Parallel()

	var out MessageLoadResult
	require.Nil(t, unmarshalMessageLoadResult(json.RawMessage(`{"kind":"message","id":1,"chatId":0,"dimensionsHeight":0,"dimensionsWidth":0,"downloadState":"Done","duration":0,"fileBytes":0,"fromId":0,"hasDeviatingTimestamp":false,"hasHtml":false,"hasLocation":false,"isBot":false,"isEdited":false,"isForwarded":false,"isInfo":false,"isSetupmessage":false,"receivedTimestamp":0,"showPadlock":false,"sortTimestamp":0,"state":0,"subject":"","systemMessageType":"Unknown","text":"hi","timestamp":0,"viewType":"Text","sender":{"address":"","authName":"","color":"","id":0,"isBlocked":false,"isKeyContact":false,"isVerified":false,"lastSeen":0,"name":"","nameAndAddr":"","profileImage":null,"status":"","verifiedBy":0}}`), &out))
	require.Equal(t, "message", out.GetKind())

	require.Nil(t, unmarshalMessageLoadResult(json.RawMessage(`{"kind":"loadingError","error":"oops"}`), &out))
	require.Equal(t, "loadingError", out.GetKind())

	require.NotNil(t, unmarshalMessageLoadResult(json.RawMessage(`notjson`), &out))
	require.NotNil(t, unmarshalMessageLoadResult(json.RawMessage(`{"kind":"Unknown"}`), &out))
}

func TestMessageQuote_MarshalJSON(t *testing.T) {
	t.Parallel()

	justText := &MessageQuoteJustText{Text: "quoted text"}
	data, err := json.Marshal(justText)
	require.Nil(t, err)
	require.Contains(t, string(data), `"kind":"JustText"`)

	withMsg := &MessageQuoteWithMessage{Text: "replied text", MessageId: 5}
	data, err = json.Marshal(withMsg)
	require.Nil(t, err)
	require.Contains(t, string(data), `"kind":"WithMessage"`)
}

func TestUnmarshalMessageQuote(t *testing.T) {
	t.Parallel()

	var out MessageQuote
	require.Nil(t, unmarshalMessageQuote(json.RawMessage(`{"kind":"JustText","text":"hi"}`), &out))
	require.Equal(t, "JustText", out.GetKind())

	require.Nil(t, unmarshalMessageQuote(json.RawMessage(`{"kind":"WithMessage","authorDisplayColor":"#fff","authorDisplayName":"Alice","chatId":1,"isForwarded":false,"messageId":2,"text":"hi","viewType":"Text"}`), &out))
	require.Equal(t, "WithMessage", out.GetKind())

	require.NotNil(t, unmarshalMessageQuote(json.RawMessage(`notjson`), &out))
	require.NotNil(t, unmarshalMessageQuote(json.RawMessage(`{"kind":"Unknown"}`), &out))
}

func TestMessage_UnmarshalJSON_WithQuote(t *testing.T) {
	t.Parallel()

	var msg Message
	require.Nil(t, json.Unmarshal([]byte(`{"chatId":0,"dimensionsHeight":0,"dimensionsWidth":0,"downloadState":"Done","duration":0,"fileBytes":0,"fromId":0,"hasDeviatingTimestamp":false,"hasHtml":false,"hasLocation":false,"id":1,"isBot":false,"isEdited":false,"isForwarded":false,"isInfo":false,"receivedTimestamp":0,"showPadlock":false,"sortTimestamp":0,"state":0,"subject":"","systemMessageType":"Unknown","text":"hello","timestamp":0,"viewType":"Text","quote":{"kind":"JustText","text":"quoted text"},"sender":{"address":"","authName":"","color":"","id":0,"isBlocked":false,"isKeyContact":false,"isVerified":false,"lastSeen":0,"name":"","nameAndAddr":"","profileImage":null,"status":"","verifiedBy":0}}`), &msg))
	require.NotNil(t, msg.Quote)
	require.Equal(t, "JustText", (*msg.Quote).GetKind())
}

func TestMuteDuration_MarshalJSON(t *testing.T) {
	t.Parallel()

	notMuted := &MuteDurationNotMuted{}
	data, err := json.Marshal(notMuted)
	require.Nil(t, err)
	require.Contains(t, string(data), `"kind":"NotMuted"`)
	require.Equal(t, "NotMuted", notMuted.GetKind())

	forever := &MuteDurationForever{}
	data, err = json.Marshal(forever)
	require.Nil(t, err)
	require.Contains(t, string(data), `"kind":"Forever"`)
	require.Equal(t, "Forever", forever.GetKind())

	until := &MuteDurationUntil{Duration: 3600}
	data, err = json.Marshal(until)
	require.Nil(t, err)
	require.Contains(t, string(data), `"kind":"Until"`)
	require.Equal(t, "Until", until.GetKind())
}

func TestQr_MarshalJSON(t *testing.T) {
	t.Parallel()

	qrs := []Qr{
		&QrAskVerifyContact{Authcode: "abc", ContactId: 1, Fingerprint: "fp", Invitenumber: "inv"},
		&QrAskVerifyGroup{Authcode: "abc", ContactId: 1, Fingerprint: "fp", Grpid: "grp", Grpname: "Group", Invitenumber: "inv"},
		&QrAskJoinBroadcast{Authcode: "abc", ContactId: 1, Fingerprint: "fp", Grpid: "grp", Invitenumber: "inv", Name: "channel"},
		&QrFprOk{ContactId: 1},
		&QrFprMismatch{},
		&QrFprWithoutAddr{Fingerprint: "fp"},
		&QrAccount{Domain: "example.com"},
		&QrBackup2{AuthToken: "tok", NodeAddr: "addr"},
		&QrBackupTooNew{},
		&QrWebrtcInstance{Domain: "example.com", InstancePattern: "pattern"},
		&QrProxy{Host: "proxy.example.com", Port: 8080, Url: "http://proxy.example.com:8080"},
		&QrAddr{ContactId: 1},
		&QrUrl{Url: "https://example.com"},
		&QrText{Text: "hello"},
		&QrWithdrawVerifyContact{Authcode: "abc", ContactId: 1, Fingerprint: "fp", Invitenumber: "inv"},
		&QrWithdrawVerifyGroup{Authcode: "abc", ContactId: 1, Fingerprint: "fp", Grpid: "grp", Grpname: "Group", Invitenumber: "inv"},
		&QrWithdrawJoinBroadcast{Authcode: "abc", ContactId: 1, Fingerprint: "fp", Grpid: "grp", Invitenumber: "inv", Name: "ch"},
		&QrReviveVerifyContact{Authcode: "abc", ContactId: 1, Fingerprint: "fp", Invitenumber: "inv"},
		&QrReviveVerifyGroup{Authcode: "abc", ContactId: 1, Fingerprint: "fp", Grpid: "grp", Grpname: "Group", Invitenumber: "inv"},
		&QrReviveJoinBroadcast{Authcode: "abc", ContactId: 1, Fingerprint: "fp", Grpid: "grp", Invitenumber: "inv", Name: "ch"},
		&QrLogin{Address: "user@example.com"},
	}

	for _, qr := range qrs {
		data, err := json.Marshal(qr)
		require.Nil(t, err, "qr %T", qr)
		require.Contains(t, string(data), `"kind":"`+qr.GetKind()+`"`, "qr %T", qr)
	}
}

func TestUnmarshalQr(t *testing.T) {
	t.Parallel()

	variants := []struct {
		kind    string
		payload string
	}{
		{"askVerifyContact", `{"kind":"askVerifyContact","authcode":"a","contact_id":1,"fingerprint":"fp","invitenumber":"inv"}`},
		{"askVerifyGroup", `{"kind":"askVerifyGroup","authcode":"a","contact_id":1,"fingerprint":"fp","grpid":"g","grpname":"G","invitenumber":"inv"}`},
		{"askJoinBroadcast", `{"kind":"askJoinBroadcast","authcode":"a","contact_id":1,"fingerprint":"fp","grpid":"g","invitenumber":"inv","name":"ch"}`},
		{"fprOk", `{"kind":"fprOk","contact_id":1}`},
		{"fprMismatch", `{"kind":"fprMismatch"}`},
		{"fprWithoutAddr", `{"kind":"fprWithoutAddr","fingerprint":"fp"}`},
		{"account", `{"kind":"account","domain":"example.com"}`},
		{"backup2", `{"kind":"backup2","auth_token":"tok","node_addr":"addr"}`},
		{"backupTooNew", `{"kind":"backupTooNew"}`},
		{"webrtcInstance", `{"kind":"webrtcInstance","domain":"example.com","instance_pattern":"p"}`},
		{"proxy", `{"kind":"proxy","host":"proxy.example.com","port":8080,"url":"http://proxy.example.com:8080"}`},
		{"addr", `{"kind":"addr","contact_id":1}`},
		{"url", `{"kind":"url","url":"https://example.com"}`},
		{"text", `{"kind":"text","text":"hello"}`},
		{"withdrawVerifyContact", `{"kind":"withdrawVerifyContact","authcode":"a","contact_id":1,"fingerprint":"fp","invitenumber":"inv"}`},
		{"withdrawVerifyGroup", `{"kind":"withdrawVerifyGroup","authcode":"a","contact_id":1,"fingerprint":"fp","grpid":"g","grpname":"G","invitenumber":"inv"}`},
		{"withdrawJoinBroadcast", `{"kind":"withdrawJoinBroadcast","authcode":"a","contact_id":1,"fingerprint":"fp","grpid":"g","invitenumber":"inv","name":"ch"}`},
		{"reviveVerifyContact", `{"kind":"reviveVerifyContact","authcode":"a","contact_id":1,"fingerprint":"fp","invitenumber":"inv"}`},
		{"reviveVerifyGroup", `{"kind":"reviveVerifyGroup","authcode":"a","contact_id":1,"fingerprint":"fp","grpid":"g","grpname":"G","invitenumber":"inv"}`},
		{"reviveJoinBroadcast", `{"kind":"reviveJoinBroadcast","authcode":"a","contact_id":1,"fingerprint":"fp","grpid":"g","invitenumber":"inv","name":"ch"}`},
		{"login", `{"kind":"login","address":"user@example.com"}`},
	}

	for _, v := range variants {
		var out Qr
		err := unmarshalQr(json.RawMessage(v.payload), &out)
		require.Nil(t, err, "variant %s", v.kind)
		require.Equal(t, v.kind, out.GetKind(), "variant %s", v.kind)
	}

	var out Qr
	require.NotNil(t, unmarshalQr(json.RawMessage(`notjson`), &out))
	require.NotNil(t, unmarshalQr(json.RawMessage(`{"kind":"Unknown"}`), &out))
}
