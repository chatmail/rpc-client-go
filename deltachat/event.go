package deltachat

type eventType string

const (
	eventTypeUnknown                     eventType = "UnknownEvent"
	eventTypeInfo                        eventType = "Info"
	eventTypeSmtpConnected               eventType = "SmtpConnected"
	eventTypeImapConnected               eventType = "ImapConnected"
	eventTypeSmtpMessageSent             eventType = "SmtpMessageSent"
	eventTypeImapMessageDeleted          eventType = "ImapMessageDeleted"
	eventTypeImapMessageMoved            eventType = "ImapMessageMoved"
	eventTypeImapInboxIdle               eventType = "ImapInboxIdle"
	eventTypeNewBlobFile                 eventType = "NewBlobFile"
	eventTypeDeletedBlobFile             eventType = "DeletedBlobFile"
	eventTypeWarning                     eventType = "Warning"
	eventTypeError                       eventType = "Error"
	eventTypeErrorSelfNotInGroup         eventType = "ErrorSelfNotInGroup"
	eventTypeMsgsChanged                 eventType = "MsgsChanged"
	eventTypeReactionsChanged            eventType = "ReactionsChanged"
	eventTypeIncomingMsg                 eventType = "IncomingMsg"
	eventTypeIncomingMsgBunch            eventType = "IncomingMsgBunch"
	eventTypeMsgsNoticed                 eventType = "MsgsNoticed"
	eventTypeMsgDelivered                eventType = "MsgDelivered"
	eventTypeMsgFailed                   eventType = "MsgFailed"
	eventTypeMsgRead                     eventType = "MsgRead"
	eventTypeMsgDeleted                  eventType = "MsgDeleted"
	eventTypeChatModified                eventType = "ChatModified"
	eventTypeChatEphemeralTimerModified  eventType = "ChatEphemeralTimerModified"
	eventTypeContactsChanged             eventType = "ContactsChanged"
	eventTypeLocationChanged             eventType = "LocationChanged"
	eventTypeConfigureProgress           eventType = "ConfigureProgress"
	eventTypeImexProgress                eventType = "ImexProgress"
	eventTypeImexFileWritten             eventType = "ImexFileWritten"
	eventTypeSecurejoinInviterProgress   eventType = "SecurejoinInviterProgress"
	eventTypeSecurejoinJoinerProgress    eventType = "SecurejoinJoinerProgress"
	eventTypeConnectivityChanged         eventType = "ConnectivityChanged"
	eventTypeSelfavatarChanged           eventType = "SelfavatarChanged"
	eventTypeConfigSynced                eventType = "ConfigSynced"
	eventTypeWebxdcStatusUpdate          eventType = "WebxdcStatusUpdate"
	eventTypeWebxdcInstanceDeleted       eventType = "WebxdcInstanceDeleted"
	eventTypeAccountsBackgroundFetchDone eventType = "AccountsBackgroundFetchDone"
)

type _Event struct {
	ContextId AccountId
	Event     *_EventData
}

type _EventData struct {
	Kind               eventType
	Msg                string
	File               string
	ChatId             ChatId
	MsgId              MsgId
	ContactId          ContactId
	MsgIds             []MsgId
	Timer              int
	Progress           uint
	Comment            string
	Path               string
	StatusUpdateSerial uint
	Key                string
}

func (eventData *_EventData) ToEvent() Event {
	var event Event
	switch eventData.Kind {
	case eventTypeInfo:
		event = EventInfo{Msg: eventData.Msg}
	case eventTypeSmtpConnected:
		event = EventSmtpConnected{Msg: eventData.Msg}
	case eventTypeImapConnected:
		event = EventImapConnected{Msg: eventData.Msg}
	case eventTypeSmtpMessageSent:
		event = EventSmtpMessageSent{Msg: eventData.Msg}
	case eventTypeImapMessageDeleted:
		event = EventImapMessageDeleted{Msg: eventData.Msg}
	case eventTypeImapMessageMoved:
		event = EventImapMessageMoved{Msg: eventData.Msg}
	case eventTypeImapInboxIdle:
		event = EventImapInboxIdle{}
	case eventTypeNewBlobFile:
		event = EventNewBlobFile{File: eventData.File}
	case eventTypeDeletedBlobFile:
		event = EventDeletedBlobFile{File: eventData.File}
	case eventTypeWarning:
		event = EventWarning{Msg: eventData.Msg}
	case eventTypeError:
		event = EventError{Msg: eventData.Msg}
	case eventTypeErrorSelfNotInGroup:
		event = EventErrorSelfNotInGroup{Msg: eventData.Msg}
	case eventTypeMsgsChanged:
		event = EventMsgsChanged{ChatId: eventData.ChatId, MsgId: eventData.MsgId}
	case eventTypeReactionsChanged:
		event = EventReactionsChanged{
			ChatId:    eventData.ChatId,
			MsgId:     eventData.MsgId,
			ContactId: eventData.ContactId,
		}
	case eventTypeIncomingMsg:
		event = EventIncomingMsg{ChatId: eventData.ChatId, MsgId: eventData.MsgId}
	case eventTypeIncomingMsgBunch:
		event = EventIncomingMsgBunch{}
	case eventTypeMsgsNoticed:
		event = EventMsgsNoticed{ChatId: eventData.ChatId}
	case eventTypeMsgDelivered:
		event = EventMsgDelivered{ChatId: eventData.ChatId, MsgId: eventData.MsgId}
	case eventTypeMsgFailed:
		event = EventMsgFailed{ChatId: eventData.ChatId, MsgId: eventData.MsgId}
	case eventTypeMsgRead:
		event = EventMsgRead{ChatId: eventData.ChatId, MsgId: eventData.MsgId}
	case eventTypeMsgDeleted:
		event = EventMsgDeleted{ChatId: eventData.ChatId, MsgId: eventData.MsgId}
	case eventTypeChatModified:
		event = EventChatModified{ChatId: eventData.ChatId}
	case eventTypeChatEphemeralTimerModified:
		event = EventChatEphemeralTimerModified{
			ChatId: eventData.ChatId,
			Timer:  eventData.Timer,
		}
	case eventTypeContactsChanged:
		event = EventContactsChanged{ContactId: eventData.ContactId}
	case eventTypeLocationChanged:
		event = EventLocationChanged{ContactId: eventData.ContactId}
	case eventTypeConfigureProgress:
		event = EventConfigureProgress{Progress: eventData.Progress, Comment: eventData.Comment}
	case eventTypeImexProgress:
		event = EventImexProgress{Progress: eventData.Progress}
	case eventTypeImexFileWritten:
		event = EventImexFileWritten{Path: eventData.Path}
	case eventTypeSecurejoinInviterProgress:
		event = EventSecurejoinInviterProgress{
			ContactId: eventData.ContactId,
			Progress:  eventData.Progress,
		}
	case eventTypeSecurejoinJoinerProgress:
		event = EventSecurejoinJoinerProgress{
			ContactId: eventData.ContactId,
			Progress:  eventData.Progress,
		}
	case eventTypeConnectivityChanged:
		event = EventConnectivityChanged{}
	case eventTypeSelfavatarChanged:
		event = EventSelfavatarChanged{}
	case eventTypeConfigSynced:
		event = EventConfigSynced{Key: eventData.Key}
	case eventTypeWebxdcStatusUpdate:
		event = EventWebxdcStatusUpdate{
			MsgId:              eventData.MsgId,
			StatusUpdateSerial: eventData.StatusUpdateSerial,
		}
	case eventTypeWebxdcInstanceDeleted:
		event = EventWebxdcInstanceDeleted{MsgId: eventData.MsgId}
	case eventTypeAccountsBackgroundFetchDone:
		event = EventAccountsBackgroundFetchDone{}
	default:
		event = UnknownEvent{Kind: eventData.Kind}
	}
	return event
}

// Delta Chat core Event
type Event interface {
	eventType() eventType
}

// Unknown event from a newer unsupported core version
type UnknownEvent struct {
	Kind eventType
}

func (event UnknownEvent) eventType() eventType {
	return eventTypeUnknown
}

// The library-user may write an informational string to the log.
//
// This event should *not* be reported to the end-user using a popup or something like
// that.
type EventInfo struct {
	Msg string
}

func (event EventInfo) eventType() eventType {
	return eventTypeInfo
}

// Emitted when SMTP connection is established and login was successful.
type EventSmtpConnected struct {
	Msg string
}

func (event EventSmtpConnected) eventType() eventType {
	return eventTypeSmtpConnected
}

// Emitted when IMAP connection is established and login was successful.
type EventImapConnected struct {
	Msg string
}

func (event EventImapConnected) eventType() eventType {
	return eventTypeImapConnected
}

// Emitted when a message was successfully sent to the SMTP server.
type EventSmtpMessageSent struct {
	Msg string
}

func (event EventSmtpMessageSent) eventType() eventType {
	return eventTypeSmtpMessageSent
}

// Emitted when an IMAP message has been marked as deleted
type EventImapMessageDeleted struct {
	Msg string
}

func (event EventImapMessageDeleted) eventType() eventType {
	return eventTypeImapMessageDeleted
}

// Emitted when an IMAP message has been moved
type EventImapMessageMoved struct {
	Msg string
}

func (event EventImapMessageMoved) eventType() eventType {
	return eventTypeImapMessageMoved
}

// Emitted before going into IDLE on the Inbox folder.
type EventImapInboxIdle struct{}

func (event EventImapInboxIdle) eventType() eventType {
	return eventTypeImapInboxIdle
}

// Emitted when an new file in the $BLOBDIR was created
type EventNewBlobFile struct {
	File string
}

func (event EventNewBlobFile) eventType() eventType {
	return eventTypeNewBlobFile
}

// Emitted when an file in the $BLOBDIR was deleted
type EventDeletedBlobFile struct {
	File string
}

func (event EventDeletedBlobFile) eventType() eventType {
	return eventTypeDeletedBlobFile
}

// The library-user should write a warning string to the log.
//
// This event should *not* be reported to the end-user using a popup or something like
// that.
type EventWarning struct {
	Msg string
}

func (event EventWarning) eventType() eventType {
	return eventTypeWarning
}

// The library-user should report an error to the end-user.
//
// As most things are asynchronous, things may go wrong at any time and the user
// should not be disturbed by a dialog or so.  Instead, use a bubble or so.
//
// However, for ongoing processes (eg. Account.Configure())
// or for functions that are expected to fail (eg. Message.AutocryptContinueKeyTransfer())
// it might be better to delay showing these events until the function has really
// failed (returned false). It should be sufficient to report only the *last* error
// in a messasge box then.
type EventError struct {
	Msg string
}

func (event EventError) eventType() eventType {
	return eventTypeError
}

// An action cannot be performed because the user is not in the group.
// Reported eg. after a call to
// Chat.SetName(), Chat.SetImage(),
// Chat.AddContact(), Chat.RemoveContact(),
// and messages sending functions.
type EventErrorSelfNotInGroup struct {
	Msg string
}

func (event EventErrorSelfNotInGroup) eventType() eventType {
	return eventTypeErrorSelfNotInGroup
}

// Messages or chats changed.  One or more messages or chats changed for various
// reasons in the database:
// - Messages sent, received or removed
// - Chats created, deleted or archived
// - A draft has been set
//
// ChatId is set if only a single chat is affected by the changes, otherwise 0.
// MsgId is set if only a single message is affected by the changes, otherwise 0.
type EventMsgsChanged struct {
	ChatId ChatId
	MsgId  MsgId
}

func (event EventMsgsChanged) eventType() eventType {
	return eventTypeMsgsChanged
}

// Reactions for the message changed.
type EventReactionsChanged struct {
	ChatId    ChatId
	MsgId     MsgId
	ContactId ContactId
}

func (event EventReactionsChanged) eventType() eventType {
	return eventTypeReactionsChanged
}

// There is a fresh message. Typically, the user will show an notification
// when receiving this message.
//
// There is no extra EventMsgsChanged event send together with this event.
type EventIncomingMsg struct {
	ChatId ChatId
	MsgId  MsgId
}

func (event EventIncomingMsg) eventType() eventType {
	return eventTypeIncomingMsg
}

// Downloading a bunch of messages just finished. This is an experimental
// event to allow the UI to only show one notification per message bunch,
// instead of cluttering the user with many notifications.
//
// msg_ids contains the message ids.
type EventIncomingMsgBunch struct {
}

func (event EventIncomingMsgBunch) eventType() eventType {
	return eventTypeIncomingMsgBunch
}

// Messages were seen or noticed.
// chat id is always set.
type EventMsgsNoticed struct {
	ChatId ChatId
}

func (event EventMsgsNoticed) eventType() eventType {
	return eventTypeMsgsNoticed
}

// A single message is sent successfully. State changed from  MsgStateOutPending to
// MsgStateOutDelivered.
type EventMsgDelivered struct {
	ChatId ChatId
	MsgId  MsgId
}

func (event EventMsgDelivered) eventType() eventType {
	return eventTypeMsgDelivered
}

// A single message could not be sent. State changed from MsgStateOutPending or MsgStateOutDelivered to
// MsgStateOutFailed.
type EventMsgFailed struct {
	ChatId ChatId
	MsgId  MsgId
}

func (event EventMsgFailed) eventType() eventType {
	return eventTypeMsgFailed
}

// A single message is read by the receiver. State changed from MsgStateOutDelivered to
// MsgStateOutMdnRcvd.
type EventMsgRead struct {
	ChatId ChatId
	MsgId  MsgId
}

func (event EventMsgRead) eventType() eventType {
	return eventTypeMsgRead
}

// A single message is deleted.
type EventMsgDeleted struct {
	ChatId ChatId
	MsgId  MsgId
}

func (event EventMsgDeleted) eventType() eventType {
	return eventTypeMsgDeleted
}

// Chat changed.  The name or the image of a chat group was changed or members were added or removed.
// Or the verify state of a chat has changed.
// See Chat.SetName(), Chat.SetImage(), Chat.AddContact()
// and Chat.RemoveContact().
//
// This event does not include ephemeral timer modification, which
// is a separate event.
type EventChatModified struct {
	ChatId ChatId
}

func (event EventChatModified) eventType() eventType {
	return eventTypeChatModified
}

// Chat ephemeral timer changed.
type EventChatEphemeralTimerModified struct {
	ChatId ChatId
	Timer  int
}

func (event EventChatEphemeralTimerModified) eventType() eventType {
	return eventTypeChatEphemeralTimerModified
}

// Contact(s) created, renamed, blocked or deleted.
type EventContactsChanged struct {
	// The id of contact that has changed, or zero if several contacts have changed.
	ContactId ContactId
}

func (event EventContactsChanged) eventType() eventType {
	return eventTypeContactsChanged
}

// Location of one or more contact has changed.
type EventLocationChanged struct {
	// The id of contact for which the location has changed, or zero if the locations of several contacts have been changed.
	ContactId ContactId
}

func (event EventLocationChanged) eventType() eventType {
	return eventTypeLocationChanged
}

// Inform about the configuration progress started by Account.Configure().
type EventConfigureProgress struct {
	// Progress.
	// 0=error, 1-999=progress in permille, 1000=success and done
	Progress uint

	// Optional progress comment or error, something to display to the user.
	Comment string
}

func (event EventConfigureProgress) eventType() eventType {
	return eventTypeConfigureProgress
}

// Inform about the import/export progress.
type EventImexProgress struct {
	// Progress.
	// (usize) 0=error, 1-999=progress in permille, 1000=success and done
	Progress uint
}

func (event EventImexProgress) eventType() eventType {
	return eventTypeImexProgress
}

// A file has been exported.
// This event may be sent after a call to Account.ExportBackup() or Account.ExportSelfKeys().
//
// A typical purpose for a handler of this event may be to make the file public to some system
// services.
type EventImexFileWritten struct {
	Path string
}

func (event EventImexFileWritten) eventType() eventType {
	return eventTypeImexFileWritten
}

// Progress information of a secure-join handshake from the view of the inviter
// (Alice, the person who shows the QR code).
//
// These events are typically sent after a joiner has scanned the QR code
// generated by Account.QrCode() or Chat.QrCode().
type EventSecurejoinInviterProgress struct {
	// ID of the contact that wants to join.
	ContactId ContactId

	// Progress as:
	// 300=vg-/vc-request received, typically shown as "bob@addr joins".
	// 600=vg-/vc-request-with-auth received, vg-member-added/vc-contact-confirm sent, typically shown as "bob@addr verified".
	// 800=vg-member-added-received received, shown as "bob@addr securely joined GROUP", only sent for the verified-group-protocol.
	// 1000=Protocol finished for this contact.
	Progress uint
}

func (event EventSecurejoinInviterProgress) eventType() eventType {
	return eventTypeSecurejoinInviterProgress
}

// Progress information of a secure-join handshake from the view of the joiner
// (Bob, the person who scans the QR code).
// The events are typically sent while Account.SecureJoin(), which
// may take some time, is executed.
type EventSecurejoinJoinerProgress struct {
	// ID of the inviting contact.
	ContactId ContactId

	// Progress as:
	// 400=vg-/vc-request-with-auth sent, typically shown as "alice@addr verified, introducing myself."
	// (Bob has verified alice and waits until Alice does the same for him)
	Progress uint
}

func (event EventSecurejoinJoinerProgress) eventType() eventType {
	return eventTypeSecurejoinJoinerProgress
}

// The connectivity to the server changed.
// This means that you should refresh the connectivity view
// and possibly the connectivtiy HTML; see Account.Connectivity() and
// Account.ConnectivityHtml() for details.
type EventConnectivityChanged struct{}

func (event EventConnectivityChanged) eventType() eventType {
	return eventTypeConnectivityChanged
}

// The user's avatar changed.
type EventSelfavatarChanged struct{}

func (event EventSelfavatarChanged) eventType() eventType {
	return eventTypeSelfavatarChanged
}

// A multi-device synced config value changed. Maybe the app needs to refresh smth. For
// uniformity this is emitted on the source device too. The value isn't here, otherwise it
// would be logged which might not be good for privacy.
type EventConfigSynced struct {
	Key string
}

func (event EventConfigSynced) eventType() eventType {
	return eventTypeConfigSynced
}

// Webxdc status update received.
type EventWebxdcStatusUpdate struct {
	MsgId              MsgId
	StatusUpdateSerial uint
}

func (event EventWebxdcStatusUpdate) eventType() eventType {
	return eventTypeWebxdcStatusUpdate
}

// Inform that a message containing a webxdc instance has been deleted
type EventWebxdcInstanceDeleted struct {
	MsgId MsgId
}

func (event EventWebxdcInstanceDeleted) eventType() eventType {
	return eventTypeWebxdcInstanceDeleted
}

// Tells that the Background fetch was completed (or timed out).
// This event acts as a marker, when you reach this event you can be sure
// that all events emitted during the background fetch were processed.
//
// This event is only emitted by the account manager
type EventAccountsBackgroundFetchDone struct{}

func (event EventAccountsBackgroundFetchDone) eventType() eventType {
	return eventTypeAccountsBackgroundFetchDone
}
