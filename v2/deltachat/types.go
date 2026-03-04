package deltachat

import (
	"encoding/json"
	"fmt"
)

// Pair is a generic two-element tuple used for RPC methods that return two values.
type Pair[A, B any] struct {
	First  A
	Second B
}

func (p *Pair[A, B]) UnmarshalJSON(data []byte) error {
	var raw [2]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[0], &p.First); err != nil {
		return err
	}
	return json.Unmarshal(raw[1], &p.Second)
}

type Account interface {
	isAccountVariant()
	GetKind() string
}

type AccountConfigured struct {
	Addr        *string `json:"addr,omitempty"`
	Color       string  `json:"color"`
	DisplayName *string `json:"displayName,omitempty"`
	Id          uint32  `json:"id"`
	// Optional tag as "Work", "Family". Meant to help profile owner to differ between profiles with similar names.
	PrivateTag   *string `json:"privateTag,omitempty"`
	ProfileImage *string `json:"profileImage,omitempty"`
}

func (*AccountConfigured) isAccountVariant() {}
func (*AccountConfigured) GetKind() string   { return "Configured" }
func (v *AccountConfigured) MarshalJSON() ([]byte, error) {
	type alias AccountConfigured
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "Configured", alias: alias(*v)})
}

type AccountUnconfigured struct {
	Id uint32 `json:"id"`
}

func (*AccountUnconfigured) isAccountVariant() {}
func (*AccountUnconfigured) GetKind() string   { return "Unconfigured" }
func (v *AccountUnconfigured) MarshalJSON() ([]byte, error) {
	type alias AccountUnconfigured
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "Unconfigured", alias: alias(*v)})
}

func unmarshalAccount(data json.RawMessage, out *Account) error {
	var header struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return err
	}
	switch header.Kind {
	case "Configured":
		var v AccountConfigured
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "Unconfigured":
		var v AccountUnconfigured
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	default:
		return fmt.Errorf("unknown Account variant: %q", header.Kind)
	}
	return nil
}

// cheaper version of fullchat, omits: - contact_ids - fresh_message_counter - ephemeral_timer - self_in_group - was_seen_recently - can_send
//
// used when you only need the basic metadata of a chat like type, name, profile picture
type BasicChat struct {
	Archived         bool     `json:"archived"`
	ChatType         ChatType `json:"chatType"`
	Color            string   `json:"color"`
	Id               uint32   `json:"id"`
	IsContactRequest bool     `json:"isContactRequest"`
	IsDeviceChat     bool     `json:"isDeviceChat"`
	// True if the chat is encrypted. This means that all messages in the chat are encrypted, and all contacts in the chat are "key-contacts", i.e. identified by the PGP key fingerprint.
	//
	// False if the chat is unencrypted. This means that all messages in the chat are unencrypted, and all contacts in the chat are "address-contacts", i.e. identified by the email address. The UI should mark this chat e.g. with a mail-letter icon.
	//
	// Unencrypted groups are called "ad-hoc groups" and the user can't add/remove members, create a QR invite code, or set an avatar. These options should therefore be disabled in the UI.
	//
	// Note that it can happen that an encrypted chat contains unencrypted messages that were received in core <= v1.159.* and vice versa.
	//
	// See also `is_key_contact` on `Contact`.
	IsEncrypted  bool    `json:"isEncrypted"`
	IsMuted      bool    `json:"isMuted"`
	IsSelfTalk   bool    `json:"isSelfTalk"`
	IsUnpromoted bool    `json:"isUnpromoted"`
	Name         string  `json:"name"`
	Pinned       bool    `json:"pinned"`
	ProfileImage *string `json:"profileImage,omitempty"`
}

type CallInfo struct {
	// True if the call is started as a video call.
	HasVideo bool `json:"hasVideo"`
	// SDP offer.
	//
	// Can be used to manually answer the call even if incoming call event was missed.
	SdpOffer string `json:"sdpOffer"`
	// Call state.
	//
	// For example, if the call is accepted, active, canceled, declined etc.
	State CallState `json:"-"`
}

func (s *CallInfo) UnmarshalJSON(data []byte) error {
	var raw struct {
		HasVideo bool            `json:"hasVideo"`
		SdpOffer string          `json:"sdpOffer"`
		State    json.RawMessage `json:"state"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.HasVideo = raw.HasVideo
	s.SdpOffer = raw.SdpOffer
	err := unmarshalCallState(raw.State, &s.State)
	if err != nil {
		return err
	}
	return nil
}

type CallState interface {
	isCallStateVariant()
	GetKind() string
}

// Fresh incoming or outgoing call that is still ringing.
//
// There is no separate state for outgoing call that has been dialled but not ringing on the other side yet as we don't know whether the other side received our call.
type CallStateAlerting struct {
}

func (*CallStateAlerting) isCallStateVariant() {}
func (*CallStateAlerting) GetKind() string     { return "Alerting" }
func (v *CallStateAlerting) MarshalJSON() ([]byte, error) {
	type alias CallStateAlerting
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "Alerting", alias: alias(*v)})
}

// Active call.
type CallStateActive struct {
}

func (*CallStateActive) isCallStateVariant() {}
func (*CallStateActive) GetKind() string     { return "Active" }
func (v *CallStateActive) MarshalJSON() ([]byte, error) {
	type alias CallStateActive
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "Active", alias: alias(*v)})
}

// Completed call that was once active and then was terminated for any reason.
type CallStateCompleted struct {
	// Call duration in seconds.
	Duration int64 `json:"duration"`
}

func (*CallStateCompleted) isCallStateVariant() {}
func (*CallStateCompleted) GetKind() string     { return "Completed" }
func (v *CallStateCompleted) MarshalJSON() ([]byte, error) {
	type alias CallStateCompleted
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "Completed", alias: alias(*v)})
}

// Incoming call that was not picked up within a timeout or was explicitly ended by the caller before we picked up.
type CallStateMissed struct {
}

func (*CallStateMissed) isCallStateVariant() {}
func (*CallStateMissed) GetKind() string     { return "Missed" }
func (v *CallStateMissed) MarshalJSON() ([]byte, error) {
	type alias CallStateMissed
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "Missed", alias: alias(*v)})
}

// Incoming call that was explicitly ended on our side before picking up or outgoing call that was declined before the timeout.
type CallStateDeclined struct {
}

func (*CallStateDeclined) isCallStateVariant() {}
func (*CallStateDeclined) GetKind() string     { return "Declined" }
func (v *CallStateDeclined) MarshalJSON() ([]byte, error) {
	type alias CallStateDeclined
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "Declined", alias: alias(*v)})
}

// Outgoing call that has been canceled on our side before receiving a response.
//
// Incoming calls cannot be canceled, on the receiver side canceled calls usually result in missed calls.
type CallStateCanceled struct {
}

func (*CallStateCanceled) isCallStateVariant() {}
func (*CallStateCanceled) GetKind() string     { return "Canceled" }
func (v *CallStateCanceled) MarshalJSON() ([]byte, error) {
	type alias CallStateCanceled
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "Canceled", alias: alias(*v)})
}

func unmarshalCallState(data json.RawMessage, out *CallState) error {
	var header struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return err
	}
	switch header.Kind {
	case "Alerting":
		var v CallStateAlerting
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "Active":
		var v CallStateActive
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "Completed":
		var v CallStateCompleted
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "Missed":
		var v CallStateMissed
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "Declined":
		var v CallStateDeclined
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "Canceled":
		var v CallStateCanceled
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	default:
		return fmt.Errorf("unknown CallState variant: %q", header.Kind)
	}
	return nil
}

type ChatListItemFetchResult interface {
	isChatListItemFetchResultVariant()
	GetKind() string
}

type ChatListItemFetchResultChatListItem struct {
	AvatarPath *string  `json:"avatarPath,omitempty"`
	ChatType   ChatType `json:"chatType"`
	Color      string   `json:"color"`
	// contact id if this is a dm chat (for view profile entry in context menu)
	DmChatContact       *uint32 `json:"dmChatContact,omitempty"`
	FreshMessageCounter uint    `json:"freshMessageCounter"`
	Id                  uint32  `json:"id"`
	IsArchived          bool    `json:"isArchived"`
	IsContactRequest    bool    `json:"isContactRequest"`
	IsDeviceTalk        bool    `json:"isDeviceTalk"`
	// True if the chat is encrypted. This means that all messages in the chat are encrypted, and all contacts in the chat are "key-contacts", i.e. identified by the PGP key fingerprint.
	//
	// False if the chat is unencrypted. This means that all messages in the chat are unencrypted, and all contacts in the chat are "address-contacts", i.e. identified by the email address. The UI should mark this chat e.g. with a mail-letter icon.
	//
	// Unencrypted groups are called "ad-hoc groups" and the user can't add/remove members, create a QR invite code, or set an avatar. These options should therefore be disabled in the UI.
	//
	// Note that it can happen that an encrypted chat contains unencrypted messages that were received in core <= v1.159.* and vice versa.
	//
	// See also `is_key_contact` on `Contact`.
	IsEncrypted bool `json:"isEncrypted"`
	// deprecated 2025-07, use chat_type instead
	IsGroup           bool      `json:"isGroup"`
	IsMuted           bool      `json:"isMuted"`
	IsPinned          bool      `json:"isPinned"`
	IsSelfInGroup     bool      `json:"isSelfInGroup"`
	IsSelfTalk        bool      `json:"isSelfTalk"`
	IsSendingLocation bool      `json:"isSendingLocation"`
	LastMessageId     *uint32   `json:"lastMessageId,omitempty"`
	LastMessageType   *Viewtype `json:"lastMessageType,omitempty"`
	LastUpdated       *int64    `json:"lastUpdated,omitempty"`
	Name              string    `json:"name"`
	// showing preview if last chat message is image
	SummaryPreviewImage *string `json:"summaryPreviewImage,omitempty"`
	SummaryStatus       uint32  `json:"summaryStatus"`
	SummaryText1        string  `json:"summaryText1"`
	SummaryText2        string  `json:"summaryText2"`
	WasSeenRecently     bool    `json:"wasSeenRecently"`
}

func (*ChatListItemFetchResultChatListItem) isChatListItemFetchResultVariant() {}
func (*ChatListItemFetchResultChatListItem) GetKind() string                   { return "ChatListItem" }
func (v *ChatListItemFetchResultChatListItem) MarshalJSON() ([]byte, error) {
	type alias ChatListItemFetchResultChatListItem
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "ChatListItem", alias: alias(*v)})
}

type ChatListItemFetchResultArchiveLink struct {
	FreshMessageCounter uint `json:"freshMessageCounter"`
}

func (*ChatListItemFetchResultArchiveLink) isChatListItemFetchResultVariant() {}
func (*ChatListItemFetchResultArchiveLink) GetKind() string                   { return "ArchiveLink" }
func (v *ChatListItemFetchResultArchiveLink) MarshalJSON() ([]byte, error) {
	type alias ChatListItemFetchResultArchiveLink
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "ArchiveLink", alias: alias(*v)})
}

type ChatListItemFetchResultError struct {
	Error string `json:"error"`
	Id    uint32 `json:"id"`
}

func (*ChatListItemFetchResultError) isChatListItemFetchResultVariant() {}
func (*ChatListItemFetchResultError) GetKind() string                   { return "Error" }
func (v *ChatListItemFetchResultError) MarshalJSON() ([]byte, error) {
	type alias ChatListItemFetchResultError
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "Error", alias: alias(*v)})
}

func unmarshalChatListItemFetchResult(data json.RawMessage, out *ChatListItemFetchResult) error {
	var header struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return err
	}
	switch header.Kind {
	case "ChatListItem":
		var v ChatListItemFetchResultChatListItem
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "ArchiveLink":
		var v ChatListItemFetchResultArchiveLink
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "Error":
		var v ChatListItemFetchResultError
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	default:
		return fmt.Errorf("unknown ChatListItemFetchResult variant: %q", header.Kind)
	}
	return nil
}

type ChatType string

const (
	ChatTypeSingle       ChatType = "Single"
	ChatTypeGroup        ChatType = "Group"
	ChatTypeMailinglist  ChatType = "Mailinglist"
	ChatTypeOutBroadcast ChatType = "OutBroadcast"
	ChatTypeInBroadcast  ChatType = "InBroadcast"
)

type ChatVisibility string

const (
	ChatVisibilityNormal   ChatVisibility = "Normal"
	ChatVisibilityArchived ChatVisibility = "Archived"
	ChatVisibilityPinned   ChatVisibility = "Pinned"
)

type Contact struct {
	Address     string `json:"address"`
	AuthName    string `json:"authName"`
	Color       string `json:"color"`
	DisplayName string `json:"displayName"`
	// Is encryption available for this contact.
	//
	// This can only be true for key-contacts. However, it is possible to have a key-contact for which encryption is not available because we don't have a key yet, e.g. if we just scanned the fingerprint from a QR code.
	E2eeAvail bool   `json:"e2eeAvail"`
	Id        uint32 `json:"id"`
	IsBlocked bool   `json:"isBlocked"`
	// If the contact is a bot.
	IsBot bool `json:"isBot"`
	// Is the contact a key contact.
	IsKeyContact bool `json:"isKeyContact"`
	// True if the contact can be added to protected chats because SELF and contact have verified their fingerprints in both directions.
	//
	// See [`Self::verifier_id`]/`Contact.verifierId` for a guidance how to display these information.
	IsVerified bool `json:"isVerified"`
	// the contact's last seen timestamp
	LastSeen     int64   `json:"lastSeen"`
	Name         string  `json:"name"`
	NameAndAddr  string  `json:"nameAndAddr"`
	ProfileImage *string `json:"profileImage,omitempty"`
	Status       string  `json:"status"`
	// The contact ID that verified a contact.
	//
	// As verifier may be unknown, use [`Self::is_verified`]/`Contact.isVerified` to check if a contact can be added to a protected chat.
	//
	// UI should display the information in the contact's profile as follows:
	//
	// - If `verifierId` != 0, display text "Introduced by ..." with the name of the contact. Prefix the text by a green checkmark.
	//
	// - If `verifierId` == 0 and `isVerified` != 0, display "Introduced" prefixed by a green checkmark.
	//
	// - if `verifierId` == 0 and `isVerified` == 0, display nothing
	//
	// This contains the contact ID of the verifier. If it is `DC_CONTACT_ID_SELF`, we verified the contact ourself. If it is None/Null, we don't have verifier information or the contact is not verified.
	VerifierId      *uint32 `json:"verifierId,omitempty"`
	WasSeenRecently bool    `json:"wasSeenRecently"`
}

type DownloadState string

const (
	DownloadStateDone           DownloadState = "Done"
	DownloadStateAvailable      DownloadState = "Available"
	DownloadStateFailure        DownloadState = "Failure"
	DownloadStateUndecipherable DownloadState = "Undecipherable"
	DownloadStateInProgress     DownloadState = "InProgress"
)

type EnteredCertificateChecks string

const (
	// `Automatic` means that provider database setting should be taken. If there is no provider database setting for certificate checks, check certificates strictly.
	EnteredCertificateChecksAutomatic EnteredCertificateChecks = "automatic"
	// Ensure that TLS certificate is valid for the server hostname.
	EnteredCertificateChecksStrict EnteredCertificateChecks = "strict"
	// Accept certificates that are expired, self-signed or otherwise not valid for the server hostname.
	EnteredCertificateChecksAcceptInvalidCertificates EnteredCertificateChecks = "acceptInvalidCertificates"
)

// Login parameters entered by the user.
//
// Usually it will be enough to only set `addr` and `password`, and all the other settings will be autoconfigured.
type EnteredLoginParam struct {
	// Email address.
	Addr string `json:"addr"`
	// TLS options: whether to allow invalid certificates and/or invalid hostnames. Default: Automatic
	CertificateChecks *EnteredCertificateChecks `json:"certificateChecks,omitempty"`
	// Imap server port.
	ImapPort *uint16 `json:"imapPort,omitempty"`
	// Imap socket security.
	ImapSecurity *Socket `json:"imapSecurity,omitempty"`
	// Imap server hostname or IP address.
	ImapServer *string `json:"imapServer,omitempty"`
	// Imap username.
	ImapUser *string `json:"imapUser,omitempty"`
	// If true, login via OAUTH2 (not recommended anymore). Default: false
	Oauth2 *bool `json:"oauth2,omitempty"`
	// Password.
	Password string `json:"password"`
	// SMTP Password.
	//
	// Only needs to be specified if different than IMAP password.
	SmtpPassword *string `json:"smtpPassword,omitempty"`
	// SMTP server port.
	SmtpPort *uint16 `json:"smtpPort,omitempty"`
	// SMTP socket security.
	SmtpSecurity *Socket `json:"smtpSecurity,omitempty"`
	// SMTP server hostname or IP address.
	SmtpServer *string `json:"smtpServer,omitempty"`
	// SMTP username.
	SmtpUser *string `json:"smtpUser,omitempty"`
}

type EphemeralTimer interface {
	isEphemeralTimerVariant()
	GetKind() string
}

// Timer is disabled.
type EphemeralTimerDisabled struct {
}

func (*EphemeralTimerDisabled) isEphemeralTimerVariant() {}
func (*EphemeralTimerDisabled) GetKind() string          { return "disabled" }
func (v *EphemeralTimerDisabled) MarshalJSON() ([]byte, error) {
	type alias EphemeralTimerDisabled
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "disabled", alias: alias(*v)})
}

// Timer is enabled.
type EphemeralTimerEnabled struct {
	// Timer duration in seconds.
	//
	// The value cannot be 0.
	Duration uint32 `json:"duration"`
}

func (*EphemeralTimerEnabled) isEphemeralTimerVariant() {}
func (*EphemeralTimerEnabled) GetKind() string          { return "enabled" }
func (v *EphemeralTimerEnabled) MarshalJSON() ([]byte, error) {
	type alias EphemeralTimerEnabled
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "enabled", alias: alias(*v)})
}

func unmarshalEphemeralTimer(data json.RawMessage, out *EphemeralTimer) error {
	var header struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return err
	}
	switch header.Kind {
	case "disabled":
		var v EphemeralTimerDisabled
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "enabled":
		var v EphemeralTimerEnabled
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	default:
		return fmt.Errorf("unknown EphemeralTimer variant: %q", header.Kind)
	}
	return nil
}

type Event struct {
	// Account ID.
	ContextId uint32 `json:"contextId"`
	// Event payload.
	Event EventType `json:"-"`
}

func (s *Event) UnmarshalJSON(data []byte) error {
	var raw struct {
		ContextId uint32          `json:"contextId"`
		Event     json.RawMessage `json:"event"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.ContextId = raw.ContextId
	err := unmarshalEventType(raw.Event, &s.Event)
	if err != nil {
		return err
	}
	return nil
}

type EventType interface {
	isEventTypeVariant()
	GetKind() string
}

// The library-user may write an informational string to the log.
//
// This event should *not* be reported to the end-user using a popup or something like that.
type EventTypeInfo struct {
	Msg string `json:"msg"`
}

func (*EventTypeInfo) isEventTypeVariant() {}
func (*EventTypeInfo) GetKind() string     { return "Info" }
func (v *EventTypeInfo) MarshalJSON() ([]byte, error) {
	type alias EventTypeInfo
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "Info", alias: alias(*v)})
}

// Emitted when SMTP connection is established and login was successful.
type EventTypeSmtpConnected struct {
	Msg string `json:"msg"`
}

func (*EventTypeSmtpConnected) isEventTypeVariant() {}
func (*EventTypeSmtpConnected) GetKind() string     { return "SmtpConnected" }
func (v *EventTypeSmtpConnected) MarshalJSON() ([]byte, error) {
	type alias EventTypeSmtpConnected
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "SmtpConnected", alias: alias(*v)})
}

// Emitted when IMAP connection is established and login was successful.
type EventTypeImapConnected struct {
	Msg string `json:"msg"`
}

func (*EventTypeImapConnected) isEventTypeVariant() {}
func (*EventTypeImapConnected) GetKind() string     { return "ImapConnected" }
func (v *EventTypeImapConnected) MarshalJSON() ([]byte, error) {
	type alias EventTypeImapConnected
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "ImapConnected", alias: alias(*v)})
}

// Emitted when a message was successfully sent to the SMTP server.
type EventTypeSmtpMessageSent struct {
	Msg string `json:"msg"`
}

func (*EventTypeSmtpMessageSent) isEventTypeVariant() {}
func (*EventTypeSmtpMessageSent) GetKind() string     { return "SmtpMessageSent" }
func (v *EventTypeSmtpMessageSent) MarshalJSON() ([]byte, error) {
	type alias EventTypeSmtpMessageSent
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "SmtpMessageSent", alias: alias(*v)})
}

// Emitted when an IMAP message has been marked as deleted
type EventTypeImapMessageDeleted struct {
	Msg string `json:"msg"`
}

func (*EventTypeImapMessageDeleted) isEventTypeVariant() {}
func (*EventTypeImapMessageDeleted) GetKind() string     { return "ImapMessageDeleted" }
func (v *EventTypeImapMessageDeleted) MarshalJSON() ([]byte, error) {
	type alias EventTypeImapMessageDeleted
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "ImapMessageDeleted", alias: alias(*v)})
}

// Emitted when an IMAP message has been moved
type EventTypeImapMessageMoved struct {
	Msg string `json:"msg"`
}

func (*EventTypeImapMessageMoved) isEventTypeVariant() {}
func (*EventTypeImapMessageMoved) GetKind() string     { return "ImapMessageMoved" }
func (v *EventTypeImapMessageMoved) MarshalJSON() ([]byte, error) {
	type alias EventTypeImapMessageMoved
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "ImapMessageMoved", alias: alias(*v)})
}

// Emitted before going into IDLE on the Inbox folder.
type EventTypeImapInboxIdle struct {
}

func (*EventTypeImapInboxIdle) isEventTypeVariant() {}
func (*EventTypeImapInboxIdle) GetKind() string     { return "ImapInboxIdle" }
func (v *EventTypeImapInboxIdle) MarshalJSON() ([]byte, error) {
	type alias EventTypeImapInboxIdle
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "ImapInboxIdle", alias: alias(*v)})
}

// Emitted when an new file in the $BLOBDIR was created
type EventTypeNewBlobFile struct {
	File string `json:"file"`
}

func (*EventTypeNewBlobFile) isEventTypeVariant() {}
func (*EventTypeNewBlobFile) GetKind() string     { return "NewBlobFile" }
func (v *EventTypeNewBlobFile) MarshalJSON() ([]byte, error) {
	type alias EventTypeNewBlobFile
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "NewBlobFile", alias: alias(*v)})
}

// Emitted when an file in the $BLOBDIR was deleted
type EventTypeDeletedBlobFile struct {
	File string `json:"file"`
}

func (*EventTypeDeletedBlobFile) isEventTypeVariant() {}
func (*EventTypeDeletedBlobFile) GetKind() string     { return "DeletedBlobFile" }
func (v *EventTypeDeletedBlobFile) MarshalJSON() ([]byte, error) {
	type alias EventTypeDeletedBlobFile
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "DeletedBlobFile", alias: alias(*v)})
}

// The library-user should write a warning string to the log.
//
// This event should *not* be reported to the end-user using a popup or something like that.
type EventTypeWarning struct {
	Msg string `json:"msg"`
}

func (*EventTypeWarning) isEventTypeVariant() {}
func (*EventTypeWarning) GetKind() string     { return "Warning" }
func (v *EventTypeWarning) MarshalJSON() ([]byte, error) {
	type alias EventTypeWarning
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "Warning", alias: alias(*v)})
}

// The library-user should report an error to the end-user.
//
// As most things are asynchronous, things may go wrong at any time and the user should not be disturbed by a dialog or so.  Instead, use a bubble or so.
//
// However, for ongoing processes (eg. configure()) or for functions that are expected to fail (eg. autocryptContinueKeyTransfer()) it might be better to delay showing these events until the function has really failed (returned false). It should be sufficient to report only the *last* error in a message box then.
type EventTypeError struct {
	Msg string `json:"msg"`
}

func (*EventTypeError) isEventTypeVariant() {}
func (*EventTypeError) GetKind() string     { return "Error" }
func (v *EventTypeError) MarshalJSON() ([]byte, error) {
	type alias EventTypeError
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "Error", alias: alias(*v)})
}

// An action cannot be performed because the user is not in the group. Reported eg. after a call to setChatName(), setChatProfileImage(), addContactToChat(), removeContactFromChat(), and messages sending functions.
type EventTypeErrorSelfNotInGroup struct {
	Msg string `json:"msg"`
}

func (*EventTypeErrorSelfNotInGroup) isEventTypeVariant() {}
func (*EventTypeErrorSelfNotInGroup) GetKind() string     { return "ErrorSelfNotInGroup" }
func (v *EventTypeErrorSelfNotInGroup) MarshalJSON() ([]byte, error) {
	type alias EventTypeErrorSelfNotInGroup
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "ErrorSelfNotInGroup", alias: alias(*v)})
}

// Messages or chats changed.  One or more messages or chats changed for various reasons in the database: - Messages sent, received or removed - Chats created, deleted or archived - A draft has been set
type EventTypeMsgsChanged struct {
	// Set if only a single chat is affected by the changes, otherwise 0.
	ChatId uint32 `json:"chatId"`
	// Set if only a single message is affected by the changes, otherwise 0.
	MsgId uint32 `json:"msgId"`
}

func (*EventTypeMsgsChanged) isEventTypeVariant() {}
func (*EventTypeMsgsChanged) GetKind() string     { return "MsgsChanged" }
func (v *EventTypeMsgsChanged) MarshalJSON() ([]byte, error) {
	type alias EventTypeMsgsChanged
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "MsgsChanged", alias: alias(*v)})
}

// Reactions for the message changed.
type EventTypeReactionsChanged struct {
	// ID of the chat which the message belongs to.
	ChatId uint32 `json:"chatId"`
	// ID of the contact whose reaction set is changed.
	ContactId uint32 `json:"contactId"`
	// ID of the message for which reactions were changed.
	MsgId uint32 `json:"msgId"`
}

func (*EventTypeReactionsChanged) isEventTypeVariant() {}
func (*EventTypeReactionsChanged) GetKind() string     { return "ReactionsChanged" }
func (v *EventTypeReactionsChanged) MarshalJSON() ([]byte, error) {
	type alias EventTypeReactionsChanged
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "ReactionsChanged", alias: alias(*v)})
}

// A reaction to one's own sent message received. Typically, the UI will show a notification for that.
//
// In addition to this event, ReactionsChanged is emitted.
type EventTypeIncomingReaction struct {
	// ID of the chat which the message belongs to.
	ChatId uint32 `json:"chatId"`
	// ID of the contact whose reaction set is changed.
	ContactId uint32 `json:"contactId"`
	// ID of the message for which reactions were changed.
	MsgId uint32 `json:"msgId"`
	// The reaction.
	Reaction string `json:"reaction"`
}

func (*EventTypeIncomingReaction) isEventTypeVariant() {}
func (*EventTypeIncomingReaction) GetKind() string     { return "IncomingReaction" }
func (v *EventTypeIncomingReaction) MarshalJSON() ([]byte, error) {
	type alias EventTypeIncomingReaction
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "IncomingReaction", alias: alias(*v)})
}

// Incoming webxdc info or summary update, should be notified.
type EventTypeIncomingWebxdcNotify struct {
	// ID of the chat.
	ChatId uint32 `json:"chatId"`
	// ID of the contact sending.
	ContactId uint32 `json:"contactId"`
	// Link assigned to this notification, if any.
	Href *string `json:"href,omitempty"`
	// ID of the added info message or webxdc instance in case of summary change.
	MsgId uint32 `json:"msgId"`
	// Text to notify.
	Text string `json:"text"`
}

func (*EventTypeIncomingWebxdcNotify) isEventTypeVariant() {}
func (*EventTypeIncomingWebxdcNotify) GetKind() string     { return "IncomingWebxdcNotify" }
func (v *EventTypeIncomingWebxdcNotify) MarshalJSON() ([]byte, error) {
	type alias EventTypeIncomingWebxdcNotify
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "IncomingWebxdcNotify", alias: alias(*v)})
}

// There is a fresh message. Typically, the user will show a notification when receiving this message.
//
// There is no extra #DC_EVENT_MSGS_CHANGED event sent together with this event.
type EventTypeIncomingMsg struct {
	// ID of the chat where the message is assigned.
	ChatId uint32 `json:"chatId"`
	// ID of the message.
	MsgId uint32 `json:"msgId"`
}

func (*EventTypeIncomingMsg) isEventTypeVariant() {}
func (*EventTypeIncomingMsg) GetKind() string     { return "IncomingMsg" }
func (v *EventTypeIncomingMsg) MarshalJSON() ([]byte, error) {
	type alias EventTypeIncomingMsg
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "IncomingMsg", alias: alias(*v)})
}

// Downloading a bunch of messages just finished. This is an event to allow the UI to only show one notification per message bunch, instead of cluttering the user with many notifications.
type EventTypeIncomingMsgBunch struct {
}

func (*EventTypeIncomingMsgBunch) isEventTypeVariant() {}
func (*EventTypeIncomingMsgBunch) GetKind() string     { return "IncomingMsgBunch" }
func (v *EventTypeIncomingMsgBunch) MarshalJSON() ([]byte, error) {
	type alias EventTypeIncomingMsgBunch
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "IncomingMsgBunch", alias: alias(*v)})
}

// Messages were seen or noticed. chat id is always set.
type EventTypeMsgsNoticed struct {
	ChatId uint32 `json:"chatId"`
}

func (*EventTypeMsgsNoticed) isEventTypeVariant() {}
func (*EventTypeMsgsNoticed) GetKind() string     { return "MsgsNoticed" }
func (v *EventTypeMsgsNoticed) MarshalJSON() ([]byte, error) {
	type alias EventTypeMsgsNoticed
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "MsgsNoticed", alias: alias(*v)})
}

// A single message is sent successfully. State changed from  DC_STATE_OUT_PENDING to DC_STATE_OUT_DELIVERED, see `Message.state`.
type EventTypeMsgDelivered struct {
	// ID of the chat which the message belongs to.
	ChatId uint32 `json:"chatId"`
	// ID of the message that was successfully sent.
	MsgId uint32 `json:"msgId"`
}

func (*EventTypeMsgDelivered) isEventTypeVariant() {}
func (*EventTypeMsgDelivered) GetKind() string     { return "MsgDelivered" }
func (v *EventTypeMsgDelivered) MarshalJSON() ([]byte, error) {
	type alias EventTypeMsgDelivered
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "MsgDelivered", alias: alias(*v)})
}

// A single message could not be sent. State changed from DC_STATE_OUT_PENDING or DC_STATE_OUT_DELIVERED to DC_STATE_OUT_FAILED, see `Message.state`.
type EventTypeMsgFailed struct {
	// ID of the chat which the message belongs to.
	ChatId uint32 `json:"chatId"`
	// ID of the message that could not be sent.
	MsgId uint32 `json:"msgId"`
}

func (*EventTypeMsgFailed) isEventTypeVariant() {}
func (*EventTypeMsgFailed) GetKind() string     { return "MsgFailed" }
func (v *EventTypeMsgFailed) MarshalJSON() ([]byte, error) {
	type alias EventTypeMsgFailed
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "MsgFailed", alias: alias(*v)})
}

// A single message is read by the receiver. State changed from DC_STATE_OUT_DELIVERED to DC_STATE_OUT_MDN_RCVD, see `Message.state`.
type EventTypeMsgRead struct {
	// ID of the chat which the message belongs to.
	ChatId uint32 `json:"chatId"`
	// ID of the message that was read.
	MsgId uint32 `json:"msgId"`
}

func (*EventTypeMsgRead) isEventTypeVariant() {}
func (*EventTypeMsgRead) GetKind() string     { return "MsgRead" }
func (v *EventTypeMsgRead) MarshalJSON() ([]byte, error) {
	type alias EventTypeMsgRead
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "MsgRead", alias: alias(*v)})
}

// A single message was deleted.
//
// This event means that the message will no longer appear in the messagelist. UI should remove the message from the messagelist in response to this event if the message is currently displayed.
//
// The message may have been explicitly deleted by the user or expired. Internally the message may have been removed from the database, moved to the trash chat or hidden.
//
// This event does not indicate the message deletion from the server.
type EventTypeMsgDeleted struct {
	// ID of the chat where the message was prior to deletion. Never 0.
	ChatId uint32 `json:"chatId"`
	// ID of the deleted message. Never 0.
	MsgId uint32 `json:"msgId"`
}

func (*EventTypeMsgDeleted) isEventTypeVariant() {}
func (*EventTypeMsgDeleted) GetKind() string     { return "MsgDeleted" }
func (v *EventTypeMsgDeleted) MarshalJSON() ([]byte, error) {
	type alias EventTypeMsgDeleted
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "MsgDeleted", alias: alias(*v)})
}

// Chat changed.  The name or the image of a chat group was changed or members were added or removed. See setChatName(), setChatProfileImage(), addContactToChat() and removeContactFromChat().
//
// This event does not include ephemeral timer modification, which is a separate event.
type EventTypeChatModified struct {
	ChatId uint32 `json:"chatId"`
}

func (*EventTypeChatModified) isEventTypeVariant() {}
func (*EventTypeChatModified) GetKind() string     { return "ChatModified" }
func (v *EventTypeChatModified) MarshalJSON() ([]byte, error) {
	type alias EventTypeChatModified
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "ChatModified", alias: alias(*v)})
}

// Chat ephemeral timer changed.
type EventTypeChatEphemeralTimerModified struct {
	// Chat ID.
	ChatId uint32 `json:"chatId"`
	// New ephemeral timer value.
	Timer uint32 `json:"timer"`
}

func (*EventTypeChatEphemeralTimerModified) isEventTypeVariant() {}
func (*EventTypeChatEphemeralTimerModified) GetKind() string     { return "ChatEphemeralTimerModified" }
func (v *EventTypeChatEphemeralTimerModified) MarshalJSON() ([]byte, error) {
	type alias EventTypeChatEphemeralTimerModified
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "ChatEphemeralTimerModified", alias: alias(*v)})
}

// Chat deleted.
type EventTypeChatDeleted struct {
	// Chat ID.
	Chat_id uint32 `json:"chat_id"`
}

func (*EventTypeChatDeleted) isEventTypeVariant() {}
func (*EventTypeChatDeleted) GetKind() string     { return "ChatDeleted" }
func (v *EventTypeChatDeleted) MarshalJSON() ([]byte, error) {
	type alias EventTypeChatDeleted
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "ChatDeleted", alias: alias(*v)})
}

// Contact(s) created, renamed, blocked or deleted.
type EventTypeContactsChanged struct {
	// If set, this is the contact_id of an added contact that should be selected.
	ContactId *uint32 `json:"contactId,omitempty"`
}

func (*EventTypeContactsChanged) isEventTypeVariant() {}
func (*EventTypeContactsChanged) GetKind() string     { return "ContactsChanged" }
func (v *EventTypeContactsChanged) MarshalJSON() ([]byte, error) {
	type alias EventTypeContactsChanged
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "ContactsChanged", alias: alias(*v)})
}

// Location of one or more contact has changed.
type EventTypeLocationChanged struct {
	// contact_id of the contact for which the location has changed. If the locations of several contacts have been changed, this parameter is set to `None`.
	ContactId *uint32 `json:"contactId,omitempty"`
}

func (*EventTypeLocationChanged) isEventTypeVariant() {}
func (*EventTypeLocationChanged) GetKind() string     { return "LocationChanged" }
func (v *EventTypeLocationChanged) MarshalJSON() ([]byte, error) {
	type alias EventTypeLocationChanged
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "LocationChanged", alias: alias(*v)})
}

// Inform about the configuration progress started by configure().
type EventTypeConfigureProgress struct {
	// Progress comment or error, something to display to the user.
	Comment *string `json:"comment,omitempty"`
	// Progress.
	//
	// 0=error, 1-999=progress in permille, 1000=success and done
	Progress uint16 `json:"progress"`
}

func (*EventTypeConfigureProgress) isEventTypeVariant() {}
func (*EventTypeConfigureProgress) GetKind() string     { return "ConfigureProgress" }
func (v *EventTypeConfigureProgress) MarshalJSON() ([]byte, error) {
	type alias EventTypeConfigureProgress
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "ConfigureProgress", alias: alias(*v)})
}

// Inform about the import/export progress started by imex().
type EventTypeImexProgress struct {
	// 0=error, 1-999=progress in permille, 1000=success and done
	Progress uint16 `json:"progress"`
}

func (*EventTypeImexProgress) isEventTypeVariant() {}
func (*EventTypeImexProgress) GetKind() string     { return "ImexProgress" }
func (v *EventTypeImexProgress) MarshalJSON() ([]byte, error) {
	type alias EventTypeImexProgress
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "ImexProgress", alias: alias(*v)})
}

// A file has been exported. A file has been written by imex(). This event may be sent multiple times by a single call to imex().
//
// A typical purpose for a handler of this event may be to make the file public to some system services.
//
// @param data2 0
type EventTypeImexFileWritten struct {
	Path string `json:"path"`
}

func (*EventTypeImexFileWritten) isEventTypeVariant() {}
func (*EventTypeImexFileWritten) GetKind() string     { return "ImexFileWritten" }
func (v *EventTypeImexFileWritten) MarshalJSON() ([]byte, error) {
	type alias EventTypeImexFileWritten
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "ImexFileWritten", alias: alias(*v)})
}

// Progress event sent when SecureJoin protocol has finished from the view of the inviter (Alice, the person who shows the QR code).
//
// These events are typically sent after a joiner has scanned the QR code generated by getChatSecurejoinQrCodeSvg().
type EventTypeSecurejoinInviterProgress struct {
	// ID of the chat in case of success.
	ChatId uint32 `json:"chatId"`
	// The type of the joined chat. This can take the same values as `BasicChat.chatType` ([`crate::api::types::chat::BasicChat::chat_type`]).
	ChatType ChatType `json:"chatType"`
	// ID of the contact that wants to join.
	ContactId uint32 `json:"contactId"`
	// Progress, always 1000.
	Progress uint16 `json:"progress"`
}

func (*EventTypeSecurejoinInviterProgress) isEventTypeVariant() {}
func (*EventTypeSecurejoinInviterProgress) GetKind() string     { return "SecurejoinInviterProgress" }
func (v *EventTypeSecurejoinInviterProgress) MarshalJSON() ([]byte, error) {
	type alias EventTypeSecurejoinInviterProgress
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "SecurejoinInviterProgress", alias: alias(*v)})
}

// Progress information of a secure-join handshake from the view of the joiner (Bob, the person who scans the QR code). The events are typically sent while secureJoin(), which may take some time, is executed.
type EventTypeSecurejoinJoinerProgress struct {
	// ID of the inviting contact.
	ContactId uint32 `json:"contactId"`
	// Progress as: 400=vg-/vc-request-with-auth sent, typically shown as "alice@addr verified, introducing myself." (Bob has verified alice and waits until Alice does the same for him) 1000=vg-member-added/vc-contact-confirm received
	Progress uint16 `json:"progress"`
}

func (*EventTypeSecurejoinJoinerProgress) isEventTypeVariant() {}
func (*EventTypeSecurejoinJoinerProgress) GetKind() string     { return "SecurejoinJoinerProgress" }
func (v *EventTypeSecurejoinJoinerProgress) MarshalJSON() ([]byte, error) {
	type alias EventTypeSecurejoinJoinerProgress
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "SecurejoinJoinerProgress", alias: alias(*v)})
}

// The connectivity to the server changed. This means that you should refresh the connectivity view and possibly the connectivtiy HTML; see getConnectivity() and getConnectivityHtml() for details.
type EventTypeConnectivityChanged struct {
}

func (*EventTypeConnectivityChanged) isEventTypeVariant() {}
func (*EventTypeConnectivityChanged) GetKind() string     { return "ConnectivityChanged" }
func (v *EventTypeConnectivityChanged) MarshalJSON() ([]byte, error) {
	type alias EventTypeConnectivityChanged
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "ConnectivityChanged", alias: alias(*v)})
}

// Deprecated by `ConfigSynced`.
type EventTypeSelfavatarChanged struct {
}

func (*EventTypeSelfavatarChanged) isEventTypeVariant() {}
func (*EventTypeSelfavatarChanged) GetKind() string     { return "SelfavatarChanged" }
func (v *EventTypeSelfavatarChanged) MarshalJSON() ([]byte, error) {
	type alias EventTypeSelfavatarChanged
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "SelfavatarChanged", alias: alias(*v)})
}

// A multi-device synced config value changed. Maybe the app needs to refresh smth. For uniformity this is emitted on the source device too. The value isn't here, otherwise it would be logged which might not be good for privacy.
type EventTypeConfigSynced struct {
	// Configuration key.
	Key string `json:"key"`
}

func (*EventTypeConfigSynced) isEventTypeVariant() {}
func (*EventTypeConfigSynced) GetKind() string     { return "ConfigSynced" }
func (v *EventTypeConfigSynced) MarshalJSON() ([]byte, error) {
	type alias EventTypeConfigSynced
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "ConfigSynced", alias: alias(*v)})
}

type EventTypeWebxdcStatusUpdate struct {
	// Message ID.
	MsgId uint32 `json:"msgId"`
	// Status update ID.
	StatusUpdateSerial uint32 `json:"statusUpdateSerial"`
}

func (*EventTypeWebxdcStatusUpdate) isEventTypeVariant() {}
func (*EventTypeWebxdcStatusUpdate) GetKind() string     { return "WebxdcStatusUpdate" }
func (v *EventTypeWebxdcStatusUpdate) MarshalJSON() ([]byte, error) {
	type alias EventTypeWebxdcStatusUpdate
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "WebxdcStatusUpdate", alias: alias(*v)})
}

// Data received over an ephemeral peer channel.
type EventTypeWebxdcRealtimeData struct {
	// Realtime data.
	Data []int `json:"data"`
	// Message ID.
	MsgId uint32 `json:"msgId"`
}

func (*EventTypeWebxdcRealtimeData) isEventTypeVariant() {}
func (*EventTypeWebxdcRealtimeData) GetKind() string     { return "WebxdcRealtimeData" }
func (v *EventTypeWebxdcRealtimeData) MarshalJSON() ([]byte, error) {
	type alias EventTypeWebxdcRealtimeData
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "WebxdcRealtimeData", alias: alias(*v)})
}

// Advertisement received over an ephemeral peer channel. This can be used by bots to initiate peer-to-peer communication from their side.
type EventTypeWebxdcRealtimeAdvertisementReceived struct {
	// Message ID of the webxdc instance.
	MsgId uint32 `json:"msgId"`
}

func (*EventTypeWebxdcRealtimeAdvertisementReceived) isEventTypeVariant() {}
func (*EventTypeWebxdcRealtimeAdvertisementReceived) GetKind() string {
	return "WebxdcRealtimeAdvertisementReceived"
}
func (v *EventTypeWebxdcRealtimeAdvertisementReceived) MarshalJSON() ([]byte, error) {
	type alias EventTypeWebxdcRealtimeAdvertisementReceived
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "WebxdcRealtimeAdvertisementReceived", alias: alias(*v)})
}

// Inform that a message containing a webxdc instance has been deleted
type EventTypeWebxdcInstanceDeleted struct {
	// ID of the deleted message.
	MsgId uint32 `json:"msgId"`
}

func (*EventTypeWebxdcInstanceDeleted) isEventTypeVariant() {}
func (*EventTypeWebxdcInstanceDeleted) GetKind() string     { return "WebxdcInstanceDeleted" }
func (v *EventTypeWebxdcInstanceDeleted) MarshalJSON() ([]byte, error) {
	type alias EventTypeWebxdcInstanceDeleted
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "WebxdcInstanceDeleted", alias: alias(*v)})
}

// Tells that the Background fetch was completed (or timed out). This event acts as a marker, when you reach this event you can be sure that all events emitted during the background fetch were processed.
//
// This event is only emitted by the account manager
type EventTypeAccountsBackgroundFetchDone struct {
}

func (*EventTypeAccountsBackgroundFetchDone) isEventTypeVariant() {}
func (*EventTypeAccountsBackgroundFetchDone) GetKind() string     { return "AccountsBackgroundFetchDone" }
func (v *EventTypeAccountsBackgroundFetchDone) MarshalJSON() ([]byte, error) {
	type alias EventTypeAccountsBackgroundFetchDone
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "AccountsBackgroundFetchDone", alias: alias(*v)})
}

// Inform that set of chats or the order of the chats in the chatlist has changed.
//
// Sometimes this is emitted together with `UIChatlistItemChanged`.
type EventTypeChatlistChanged struct {
}

func (*EventTypeChatlistChanged) isEventTypeVariant() {}
func (*EventTypeChatlistChanged) GetKind() string     { return "ChatlistChanged" }
func (v *EventTypeChatlistChanged) MarshalJSON() ([]byte, error) {
	type alias EventTypeChatlistChanged
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "ChatlistChanged", alias: alias(*v)})
}

// Inform that a single chat list item changed and needs to be rerendered. If `chat_id` is set to None, then all currently visible chats need to be rerendered, and all not-visible items need to be cleared from cache if the UI has a cache.
type EventTypeChatlistItemChanged struct {
	// ID of the changed chat
	ChatId *uint32 `json:"chatId,omitempty"`
}

func (*EventTypeChatlistItemChanged) isEventTypeVariant() {}
func (*EventTypeChatlistItemChanged) GetKind() string     { return "ChatlistItemChanged" }
func (v *EventTypeChatlistItemChanged) MarshalJSON() ([]byte, error) {
	type alias EventTypeChatlistItemChanged
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "ChatlistItemChanged", alias: alias(*v)})
}

// Inform that the list of accounts has changed (an account removed or added or (not yet implemented) the account order changes)
//
// This event is only emitted by the account manager
type EventTypeAccountsChanged struct {
}

func (*EventTypeAccountsChanged) isEventTypeVariant() {}
func (*EventTypeAccountsChanged) GetKind() string     { return "AccountsChanged" }
func (v *EventTypeAccountsChanged) MarshalJSON() ([]byte, error) {
	type alias EventTypeAccountsChanged
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "AccountsChanged", alias: alias(*v)})
}

// Inform that an account property that might be shown in the account list changed, namely: - is_configured (see is_configured()) - displayname - selfavatar - private_tag
//
// This event is emitted from the account whose property changed.
type EventTypeAccountsItemChanged struct {
}

func (*EventTypeAccountsItemChanged) isEventTypeVariant() {}
func (*EventTypeAccountsItemChanged) GetKind() string     { return "AccountsItemChanged" }
func (v *EventTypeAccountsItemChanged) MarshalJSON() ([]byte, error) {
	type alias EventTypeAccountsItemChanged
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "AccountsItemChanged", alias: alias(*v)})
}

// Inform than some events have been skipped due to event channel overflow.
type EventTypeEventChannelOverflow struct {
	// Number of events skipped.
	N uint64 `json:"n"`
}

func (*EventTypeEventChannelOverflow) isEventTypeVariant() {}
func (*EventTypeEventChannelOverflow) GetKind() string     { return "EventChannelOverflow" }
func (v *EventTypeEventChannelOverflow) MarshalJSON() ([]byte, error) {
	type alias EventTypeEventChannelOverflow
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "EventChannelOverflow", alias: alias(*v)})
}

// Incoming call.
type EventTypeIncomingCall struct {
	// ID of the chat which the message belongs to.
	Chat_id uint32 `json:"chat_id"`
	// True if incoming call is a video call.
	Has_video bool `json:"has_video"`
	// ID of the info message referring to the call.
	Msg_id uint32 `json:"msg_id"`
	// User-defined info as passed to place_outgoing_call()
	Place_call_info string `json:"place_call_info"`
}

func (*EventTypeIncomingCall) isEventTypeVariant() {}
func (*EventTypeIncomingCall) GetKind() string     { return "IncomingCall" }
func (v *EventTypeIncomingCall) MarshalJSON() ([]byte, error) {
	type alias EventTypeIncomingCall
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "IncomingCall", alias: alias(*v)})
}

// Incoming call accepted. This is esp. interesting to stop ringing on other devices.
type EventTypeIncomingCallAccepted struct {
	// ID of the chat which the message belongs to.
	Chat_id uint32 `json:"chat_id"`
	// ID of the info message referring to the call.
	Msg_id uint32 `json:"msg_id"`
}

func (*EventTypeIncomingCallAccepted) isEventTypeVariant() {}
func (*EventTypeIncomingCallAccepted) GetKind() string     { return "IncomingCallAccepted" }
func (v *EventTypeIncomingCallAccepted) MarshalJSON() ([]byte, error) {
	type alias EventTypeIncomingCallAccepted
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "IncomingCallAccepted", alias: alias(*v)})
}

// Outgoing call accepted.
type EventTypeOutgoingCallAccepted struct {
	// User-defined info passed to dc_accept_incoming_call(
	Accept_call_info string `json:"accept_call_info"`
	// ID of the chat which the message belongs to.
	Chat_id uint32 `json:"chat_id"`
	// ID of the info message referring to the call.
	Msg_id uint32 `json:"msg_id"`
}

func (*EventTypeOutgoingCallAccepted) isEventTypeVariant() {}
func (*EventTypeOutgoingCallAccepted) GetKind() string     { return "OutgoingCallAccepted" }
func (v *EventTypeOutgoingCallAccepted) MarshalJSON() ([]byte, error) {
	type alias EventTypeOutgoingCallAccepted
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "OutgoingCallAccepted", alias: alias(*v)})
}

// Call ended.
type EventTypeCallEnded struct {
	// ID of the chat which the message belongs to.
	Chat_id uint32 `json:"chat_id"`
	// ID of the info message referring to the call.
	Msg_id uint32 `json:"msg_id"`
}

func (*EventTypeCallEnded) isEventTypeVariant() {}
func (*EventTypeCallEnded) GetKind() string     { return "CallEnded" }
func (v *EventTypeCallEnded) MarshalJSON() ([]byte, error) {
	type alias EventTypeCallEnded
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "CallEnded", alias: alias(*v)})
}

// One or more transports has changed.
//
// UI should update the list.
//
// This event is emitted when transport synchronization messages arrives, but not when the UI modifies the transport list by itself.
type EventTypeTransportsModified struct {
}

func (*EventTypeTransportsModified) isEventTypeVariant() {}
func (*EventTypeTransportsModified) GetKind() string     { return "TransportsModified" }
func (v *EventTypeTransportsModified) MarshalJSON() ([]byte, error) {
	type alias EventTypeTransportsModified
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "TransportsModified", alias: alias(*v)})
}

func unmarshalEventType(data json.RawMessage, out *EventType) error {
	var header struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return err
	}
	switch header.Kind {
	case "Info":
		var v EventTypeInfo
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "SmtpConnected":
		var v EventTypeSmtpConnected
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "ImapConnected":
		var v EventTypeImapConnected
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "SmtpMessageSent":
		var v EventTypeSmtpMessageSent
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "ImapMessageDeleted":
		var v EventTypeImapMessageDeleted
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "ImapMessageMoved":
		var v EventTypeImapMessageMoved
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "ImapInboxIdle":
		var v EventTypeImapInboxIdle
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "NewBlobFile":
		var v EventTypeNewBlobFile
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "DeletedBlobFile":
		var v EventTypeDeletedBlobFile
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "Warning":
		var v EventTypeWarning
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "Error":
		var v EventTypeError
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "ErrorSelfNotInGroup":
		var v EventTypeErrorSelfNotInGroup
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "MsgsChanged":
		var v EventTypeMsgsChanged
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "ReactionsChanged":
		var v EventTypeReactionsChanged
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "IncomingReaction":
		var v EventTypeIncomingReaction
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "IncomingWebxdcNotify":
		var v EventTypeIncomingWebxdcNotify
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "IncomingMsg":
		var v EventTypeIncomingMsg
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "IncomingMsgBunch":
		var v EventTypeIncomingMsgBunch
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "MsgsNoticed":
		var v EventTypeMsgsNoticed
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "MsgDelivered":
		var v EventTypeMsgDelivered
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "MsgFailed":
		var v EventTypeMsgFailed
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "MsgRead":
		var v EventTypeMsgRead
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "MsgDeleted":
		var v EventTypeMsgDeleted
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "ChatModified":
		var v EventTypeChatModified
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "ChatEphemeralTimerModified":
		var v EventTypeChatEphemeralTimerModified
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "ChatDeleted":
		var v EventTypeChatDeleted
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "ContactsChanged":
		var v EventTypeContactsChanged
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "LocationChanged":
		var v EventTypeLocationChanged
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "ConfigureProgress":
		var v EventTypeConfigureProgress
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "ImexProgress":
		var v EventTypeImexProgress
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "ImexFileWritten":
		var v EventTypeImexFileWritten
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "SecurejoinInviterProgress":
		var v EventTypeSecurejoinInviterProgress
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "SecurejoinJoinerProgress":
		var v EventTypeSecurejoinJoinerProgress
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "ConnectivityChanged":
		var v EventTypeConnectivityChanged
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "SelfavatarChanged":
		var v EventTypeSelfavatarChanged
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "ConfigSynced":
		var v EventTypeConfigSynced
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "WebxdcStatusUpdate":
		var v EventTypeWebxdcStatusUpdate
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "WebxdcRealtimeData":
		var v EventTypeWebxdcRealtimeData
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "WebxdcRealtimeAdvertisementReceived":
		var v EventTypeWebxdcRealtimeAdvertisementReceived
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "WebxdcInstanceDeleted":
		var v EventTypeWebxdcInstanceDeleted
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "AccountsBackgroundFetchDone":
		var v EventTypeAccountsBackgroundFetchDone
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "ChatlistChanged":
		var v EventTypeChatlistChanged
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "ChatlistItemChanged":
		var v EventTypeChatlistItemChanged
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "AccountsChanged":
		var v EventTypeAccountsChanged
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "AccountsItemChanged":
		var v EventTypeAccountsItemChanged
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "EventChannelOverflow":
		var v EventTypeEventChannelOverflow
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "IncomingCall":
		var v EventTypeIncomingCall
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "IncomingCallAccepted":
		var v EventTypeIncomingCallAccepted
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "OutgoingCallAccepted":
		var v EventTypeOutgoingCallAccepted
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "CallEnded":
		var v EventTypeCallEnded
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "TransportsModified":
		var v EventTypeTransportsModified
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	default:
		return fmt.Errorf("unknown EventType variant: %q", header.Kind)
	}
	return nil
}

type FullChat struct {
	Archived            bool     `json:"archived"`
	CanSend             bool     `json:"canSend"`
	ChatType            ChatType `json:"chatType"`
	Color               string   `json:"color"`
	ContactIds          []uint32 `json:"contactIds"`
	EphemeralTimer      uint32   `json:"ephemeralTimer"`
	FreshMessageCounter uint     `json:"freshMessageCounter"`
	Id                  uint32   `json:"id"`
	IsContactRequest    bool     `json:"isContactRequest"`
	IsDeviceChat        bool     `json:"isDeviceChat"`
	// True if the chat is encrypted. This means that all messages in the chat are encrypted, and all contacts in the chat are "key-contacts", i.e. identified by the PGP key fingerprint.
	//
	// False if the chat is unencrypted. This means that all messages in the chat are unencrypted, and all contacts in the chat are "address-contacts", i.e. identified by the email address. The UI should mark this chat e.g. with a mail-letter icon.
	//
	// Unencrypted groups are called "ad-hoc groups" and the user can't add/remove members, create a QR invite code, or set an avatar. These options should therefore be disabled in the UI.
	//
	// Note that it can happen that an encrypted chat contains unencrypted messages that were received in core <= v1.159.* and vice versa.
	//
	// See also `is_key_contact` on `Contact`.
	IsEncrypted        bool    `json:"isEncrypted"`
	IsMuted            bool    `json:"isMuted"`
	IsSelfTalk         bool    `json:"isSelfTalk"`
	IsUnpromoted       bool    `json:"isUnpromoted"`
	MailingListAddress *string `json:"mailingListAddress,omitempty"`
	Name               string  `json:"name"`
	// Contact IDs of the past chat members.
	PastContactIds []uint32 `json:"pastContactIds"`
	Pinned         bool     `json:"pinned"`
	ProfileImage   *string  `json:"profileImage,omitempty"`
	// Note that this is different from [`ChatListItem::is_self_in_group`](`crate::api::types::chat_list::ChatListItemFetchResult::ChatListItem::is_self_in_group`). This property should only be accessed when [`FullChat::chat_type`] is [`Chattype::Group`].
	SelfInGroup     bool `json:"selfInGroup"`
	WasSeenRecently bool `json:"wasSeenRecently"`
}

type HttpResponse struct {
	// base64-encoded response body.
	Blob string `json:"blob"`
	// Encoding, e.g. "utf-8".
	Encoding *string `json:"encoding,omitempty"`
	// MIME type, e.g. "text/plain" or "text/html".
	Mimetype *string `json:"mimetype,omitempty"`
}

type Location struct {
	Accuracy      float64 `json:"accuracy"`
	ChatId        uint32  `json:"chatId"`
	ContactId     uint32  `json:"contactId"`
	IsIndependent bool    `json:"isIndependent"`
	Latitude      float64 `json:"latitude"`
	LocationId    uint32  `json:"locationId"`
	Longitude     float64 `json:"longitude"`
	Marker        *string `json:"marker,omitempty"`
	MsgId         uint32  `json:"msgId"`
	Timestamp     int64   `json:"timestamp"`
}

type Message struct {
	ChatId           uint32        `json:"chatId"`
	DimensionsHeight int32         `json:"dimensionsHeight"`
	DimensionsWidth  int32         `json:"dimensionsWidth"`
	DownloadState    DownloadState `json:"downloadState"`
	Duration         int32         `json:"duration"`
	// An error text, if there is one.
	Error *string `json:"error,omitempty"`
	File  *string `json:"file,omitempty"`
	// The size of the file in bytes, if applicable. If message is a pre-message, then this is the size of the file to be downloaded.
	FileBytes             uint64  `json:"fileBytes"`
	FileMime              *string `json:"fileMime,omitempty"`
	FileName              *string `json:"fileName,omitempty"`
	FromId                uint32  `json:"fromId"`
	HasDeviatingTimestamp bool    `json:"hasDeviatingTimestamp"`
	HasHtml               bool    `json:"hasHtml"`
	// Check if a message has a POI location bound to it. These locations are also returned by `get_locations` method. The UI may decide to display a special icon beside such messages.
	HasLocation bool   `json:"hasLocation"`
	Id          uint32 `json:"id"`
	// if is_info is set, this refers to the contact profile that should be opened when the info message is tapped.
	InfoContactId *uint32 `json:"infoContactId,omitempty"`
	// True if the message was sent by a bot.
	IsBot              bool          `json:"isBot"`
	IsEdited           bool          `json:"isEdited"`
	IsForwarded        bool          `json:"isForwarded"`
	IsInfo             bool          `json:"isInfo"`
	IsSetupmessage     bool          `json:"isSetupmessage"`
	OriginalMsgId      *uint32       `json:"originalMsgId,omitempty"`
	OverrideSenderName *string       `json:"overrideSenderName,omitempty"`
	ParentId           *uint32       `json:"parentId,omitempty"`
	Quote              *MessageQuote `json:"-"`
	Reactions          *Reactions    `json:"reactions,omitempty"`
	ReceivedTimestamp  int64         `json:"receivedTimestamp"`
	SavedMessageId     *uint32       `json:"savedMessageId,omitempty"`
	Sender             Contact       `json:"sender"`
	SetupCodeBegin     *string       `json:"setupCodeBegin,omitempty"`
	// True if the message was correctly encrypted&signed, false otherwise. Historically, UIs showed a small padlock on the message then.
	//
	// Today, the UIs should instead show a small email-icon on the message if `show_padlock` is `false`, and nothing if it is `true`.
	ShowPadlock   bool   `json:"showPadlock"`
	SortTimestamp int64  `json:"sortTimestamp"`
	State         uint32 `json:"state"`
	Subject       string `json:"subject"`
	// when is_info is true this describes what type of system message it is
	SystemMessageType SystemMessageType `json:"systemMessageType"`
	Text              string            `json:"text"`
	Timestamp         int64             `json:"timestamp"`
	VcardContact      *VcardContact     `json:"vcardContact,omitempty"`
	ViewType          Viewtype          `json:"viewType"`
	WebxdcHref        *string           `json:"webxdcHref,omitempty"`
}

func (s *Message) UnmarshalJSON(data []byte) error {
	var raw struct {
		ChatId                uint32            `json:"chatId"`
		DimensionsHeight      int32             `json:"dimensionsHeight"`
		DimensionsWidth       int32             `json:"dimensionsWidth"`
		DownloadState         DownloadState     `json:"downloadState"`
		Duration              int32             `json:"duration"`
		Error                 *string           `json:"error,omitempty"`
		File                  *string           `json:"file,omitempty"`
		FileBytes             uint64            `json:"fileBytes"`
		FileMime              *string           `json:"fileMime,omitempty"`
		FileName              *string           `json:"fileName,omitempty"`
		FromId                uint32            `json:"fromId"`
		HasDeviatingTimestamp bool              `json:"hasDeviatingTimestamp"`
		HasHtml               bool              `json:"hasHtml"`
		HasLocation           bool              `json:"hasLocation"`
		Id                    uint32            `json:"id"`
		InfoContactId         *uint32           `json:"infoContactId,omitempty"`
		IsBot                 bool              `json:"isBot"`
		IsEdited              bool              `json:"isEdited"`
		IsForwarded           bool              `json:"isForwarded"`
		IsInfo                bool              `json:"isInfo"`
		IsSetupmessage        bool              `json:"isSetupmessage"`
		OriginalMsgId         *uint32           `json:"originalMsgId,omitempty"`
		OverrideSenderName    *string           `json:"overrideSenderName,omitempty"`
		ParentId              *uint32           `json:"parentId,omitempty"`
		Quote                 json.RawMessage   `json:"quote"`
		Reactions             *Reactions        `json:"reactions,omitempty"`
		ReceivedTimestamp     int64             `json:"receivedTimestamp"`
		SavedMessageId        *uint32           `json:"savedMessageId,omitempty"`
		Sender                Contact           `json:"sender"`
		SetupCodeBegin        *string           `json:"setupCodeBegin,omitempty"`
		ShowPadlock           bool              `json:"showPadlock"`
		SortTimestamp         int64             `json:"sortTimestamp"`
		State                 uint32            `json:"state"`
		Subject               string            `json:"subject"`
		SystemMessageType     SystemMessageType `json:"systemMessageType"`
		Text                  string            `json:"text"`
		Timestamp             int64             `json:"timestamp"`
		VcardContact          *VcardContact     `json:"vcardContact,omitempty"`
		ViewType              Viewtype          `json:"viewType"`
		WebxdcHref            *string           `json:"webxdcHref,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.ChatId = raw.ChatId
	s.DimensionsHeight = raw.DimensionsHeight
	s.DimensionsWidth = raw.DimensionsWidth
	s.DownloadState = raw.DownloadState
	s.Duration = raw.Duration
	s.Error = raw.Error
	s.File = raw.File
	s.FileBytes = raw.FileBytes
	s.FileMime = raw.FileMime
	s.FileName = raw.FileName
	s.FromId = raw.FromId
	s.HasDeviatingTimestamp = raw.HasDeviatingTimestamp
	s.HasHtml = raw.HasHtml
	s.HasLocation = raw.HasLocation
	s.Id = raw.Id
	s.InfoContactId = raw.InfoContactId
	s.IsBot = raw.IsBot
	s.IsEdited = raw.IsEdited
	s.IsForwarded = raw.IsForwarded
	s.IsInfo = raw.IsInfo
	s.IsSetupmessage = raw.IsSetupmessage
	s.OriginalMsgId = raw.OriginalMsgId
	s.OverrideSenderName = raw.OverrideSenderName
	s.ParentId = raw.ParentId
	s.Reactions = raw.Reactions
	s.ReceivedTimestamp = raw.ReceivedTimestamp
	s.SavedMessageId = raw.SavedMessageId
	s.Sender = raw.Sender
	s.SetupCodeBegin = raw.SetupCodeBegin
	s.ShowPadlock = raw.ShowPadlock
	s.SortTimestamp = raw.SortTimestamp
	s.State = raw.State
	s.Subject = raw.Subject
	s.SystemMessageType = raw.SystemMessageType
	s.Text = raw.Text
	s.Timestamp = raw.Timestamp
	s.VcardContact = raw.VcardContact
	s.ViewType = raw.ViewType
	s.WebxdcHref = raw.WebxdcHref
	if len(raw.Quote) > 0 && string(raw.Quote) != "null" {
		if err := unmarshalMessageQuote(raw.Quote, s.Quote); err != nil {
			return err
		}
	}
	return nil
}

type MessageData struct {
	File               *string                 `json:"file,omitempty"`
	Filename           *string                 `json:"filename,omitempty"`
	Html               *string                 `json:"html,omitempty"`
	Location           *Pair[float64, float64] `json:"location,omitempty"`
	OverrideSenderName *string                 `json:"overrideSenderName,omitempty"`
	// Quoted message id. Takes preference over `quoted_text` (see below).
	QuotedMessageId *uint32   `json:"quotedMessageId,omitempty"`
	QuotedText      *string   `json:"quotedText,omitempty"`
	Text            *string   `json:"text,omitempty"`
	Viewtype        *Viewtype `json:"viewtype,omitempty"`
}

type MessageInfo struct {
	EphemeralTimer EphemeralTimer `json:"-"`
	// When message is ephemeral this contains the timestamp of the message expiry
	EphemeralTimestamp *int64   `json:"ephemeralTimestamp,omitempty"`
	Error              *string  `json:"error,omitempty"`
	HopInfo            string   `json:"hopInfo"`
	Rfc724Mid          string   `json:"rfc724Mid"`
	ServerUrls         []string `json:"serverUrls"`
}

func (s *MessageInfo) UnmarshalJSON(data []byte) error {
	var raw struct {
		EphemeralTimer     json.RawMessage `json:"ephemeralTimer"`
		EphemeralTimestamp *int64          `json:"ephemeralTimestamp,omitempty"`
		Error              *string         `json:"error,omitempty"`
		HopInfo            string          `json:"hopInfo"`
		Rfc724Mid          string          `json:"rfc724Mid"`
		ServerUrls         []string        `json:"serverUrls"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.EphemeralTimestamp = raw.EphemeralTimestamp
	s.Error = raw.Error
	s.HopInfo = raw.HopInfo
	s.Rfc724Mid = raw.Rfc724Mid
	s.ServerUrls = raw.ServerUrls
	err := unmarshalEphemeralTimer(raw.EphemeralTimer, &s.EphemeralTimer)
	if err != nil {
		return err
	}
	return nil
}

type MessageListItem interface {
	isMessageListItemVariant()
	GetKind() string
}

type MessageListItemMessage struct {
	Msg_id uint32 `json:"msg_id"`
}

func (*MessageListItemMessage) isMessageListItemVariant() {}
func (*MessageListItemMessage) GetKind() string           { return "message" }
func (v *MessageListItemMessage) MarshalJSON() ([]byte, error) {
	type alias MessageListItemMessage
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "message", alias: alias(*v)})
}

// Day marker, separating messages that correspond to different days according to local time.
type MessageListItemDayMarker struct {
	// Marker timestamp, for day markers, in unix milliseconds
	Timestamp int64 `json:"timestamp"`
}

func (*MessageListItemDayMarker) isMessageListItemVariant() {}
func (*MessageListItemDayMarker) GetKind() string           { return "dayMarker" }
func (v *MessageListItemDayMarker) MarshalJSON() ([]byte, error) {
	type alias MessageListItemDayMarker
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "dayMarker", alias: alias(*v)})
}

func unmarshalMessageListItem(data json.RawMessage, out *MessageListItem) error {
	var header struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return err
	}
	switch header.Kind {
	case "message":
		var v MessageListItemMessage
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "dayMarker":
		var v MessageListItemDayMarker
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	default:
		return fmt.Errorf("unknown MessageListItem variant: %q", header.Kind)
	}
	return nil
}

type MessageLoadResult interface {
	isMessageLoadResultVariant()
	GetKind() string
}

type MessageLoadResultMessage struct {
	ChatId           uint32        `json:"chatId"`
	DimensionsHeight int32         `json:"dimensionsHeight"`
	DimensionsWidth  int32         `json:"dimensionsWidth"`
	DownloadState    DownloadState `json:"downloadState"`
	Duration         int32         `json:"duration"`
	// An error text, if there is one.
	Error *string `json:"error,omitempty"`
	File  *string `json:"file,omitempty"`
	// The size of the file in bytes, if applicable. If message is a pre-message, then this is the size of the file to be downloaded.
	FileBytes             uint64  `json:"fileBytes"`
	FileMime              *string `json:"fileMime,omitempty"`
	FileName              *string `json:"fileName,omitempty"`
	FromId                uint32  `json:"fromId"`
	HasDeviatingTimestamp bool    `json:"hasDeviatingTimestamp"`
	HasHtml               bool    `json:"hasHtml"`
	// Check if a message has a POI location bound to it. These locations are also returned by `get_locations` method. The UI may decide to display a special icon beside such messages.
	HasLocation bool   `json:"hasLocation"`
	Id          uint32 `json:"id"`
	// if is_info is set, this refers to the contact profile that should be opened when the info message is tapped.
	InfoContactId *uint32 `json:"infoContactId,omitempty"`
	// True if the message was sent by a bot.
	IsBot              bool          `json:"isBot"`
	IsEdited           bool          `json:"isEdited"`
	IsForwarded        bool          `json:"isForwarded"`
	IsInfo             bool          `json:"isInfo"`
	IsSetupmessage     bool          `json:"isSetupmessage"`
	OriginalMsgId      *uint32       `json:"originalMsgId,omitempty"`
	OverrideSenderName *string       `json:"overrideSenderName,omitempty"`
	ParentId           *uint32       `json:"parentId,omitempty"`
	Quote              *MessageQuote `json:"quote,omitempty"`
	Reactions          *Reactions    `json:"reactions,omitempty"`
	ReceivedTimestamp  int64         `json:"receivedTimestamp"`
	SavedMessageId     *uint32       `json:"savedMessageId,omitempty"`
	Sender             Contact       `json:"sender"`
	SetupCodeBegin     *string       `json:"setupCodeBegin,omitempty"`
	// True if the message was correctly encrypted&signed, false otherwise. Historically, UIs showed a small padlock on the message then.
	//
	// Today, the UIs should instead show a small email-icon on the message if `show_padlock` is `false`, and nothing if it is `true`.
	ShowPadlock   bool   `json:"showPadlock"`
	SortTimestamp int64  `json:"sortTimestamp"`
	State         uint32 `json:"state"`
	Subject       string `json:"subject"`
	// when is_info is true this describes what type of system message it is
	SystemMessageType SystemMessageType `json:"systemMessageType"`
	Text              string            `json:"text"`
	Timestamp         int64             `json:"timestamp"`
	VcardContact      *VcardContact     `json:"vcardContact,omitempty"`
	ViewType          Viewtype          `json:"viewType"`
	WebxdcHref        *string           `json:"webxdcHref,omitempty"`
}

func (*MessageLoadResultMessage) isMessageLoadResultVariant() {}
func (*MessageLoadResultMessage) GetKind() string             { return "message" }
func (v *MessageLoadResultMessage) MarshalJSON() ([]byte, error) {
	type alias MessageLoadResultMessage
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "message", alias: alias(*v)})
}

type MessageLoadResultLoadingError struct {
	Error string `json:"error"`
}

func (*MessageLoadResultLoadingError) isMessageLoadResultVariant() {}
func (*MessageLoadResultLoadingError) GetKind() string             { return "loadingError" }
func (v *MessageLoadResultLoadingError) MarshalJSON() ([]byte, error) {
	type alias MessageLoadResultLoadingError
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "loadingError", alias: alias(*v)})
}

func unmarshalMessageLoadResult(data json.RawMessage, out *MessageLoadResult) error {
	var header struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return err
	}
	switch header.Kind {
	case "message":
		var v MessageLoadResultMessage
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "loadingError":
		var v MessageLoadResultLoadingError
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	default:
		return fmt.Errorf("unknown MessageLoadResult variant: %q", header.Kind)
	}
	return nil
}

type MessageNotificationInfo struct {
	AccountId        uint32  `json:"accountId"`
	ChatId           uint32  `json:"chatId"`
	ChatName         string  `json:"chatName"`
	ChatProfileImage *string `json:"chatProfileImage,omitempty"`
	Id               uint32  `json:"id"`
	Image            *string `json:"image,omitempty"`
	ImageMimeType    *string `json:"imageMimeType,omitempty"`
	// also known as summary_text1
	SummaryPrefix *string `json:"summaryPrefix,omitempty"`
	// also known as summary_text2
	SummaryText string `json:"summaryText"`
}

type MessageQuote interface {
	isMessageQuoteVariant()
	GetKind() string
}

type MessageQuoteJustText struct {
	Text string `json:"text"`
}

func (*MessageQuoteJustText) isMessageQuoteVariant() {}
func (*MessageQuoteJustText) GetKind() string        { return "JustText" }
func (v *MessageQuoteJustText) MarshalJSON() ([]byte, error) {
	type alias MessageQuoteJustText
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "JustText", alias: alias(*v)})
}

type MessageQuoteWithMessage struct {
	AuthorDisplayColor string `json:"authorDisplayColor"`
	AuthorDisplayName  string `json:"authorDisplayName"`
	// The quoted message does not always belong to the same chat, e.g. when "Reply Privately" is used.
	ChatId             uint32   `json:"chatId"`
	Image              *string  `json:"image,omitempty"`
	IsForwarded        bool     `json:"isForwarded"`
	MessageId          uint32   `json:"messageId"`
	OverrideSenderName *string  `json:"overrideSenderName,omitempty"`
	Text               string   `json:"text"`
	ViewType           Viewtype `json:"viewType"`
}

func (*MessageQuoteWithMessage) isMessageQuoteVariant() {}
func (*MessageQuoteWithMessage) GetKind() string        { return "WithMessage" }
func (v *MessageQuoteWithMessage) MarshalJSON() ([]byte, error) {
	type alias MessageQuoteWithMessage
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "WithMessage", alias: alias(*v)})
}

func unmarshalMessageQuote(data json.RawMessage, out *MessageQuote) error {
	var header struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return err
	}
	switch header.Kind {
	case "JustText":
		var v MessageQuoteJustText
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "WithMessage":
		var v MessageQuoteWithMessage
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	default:
		return fmt.Errorf("unknown MessageQuote variant: %q", header.Kind)
	}
	return nil
}

type MessageReadReceipt struct {
	ContactId uint32 `json:"contactId"`
	Timestamp int64  `json:"timestamp"`
}

type MessageSearchResult struct {
	AuthorColor string `json:"authorColor"`
	AuthorId    uint32 `json:"authorId"`
	// if sender name if overridden it will show it as ~alias
	AuthorName           string   `json:"authorName"`
	AuthorProfileImage   *string  `json:"authorProfileImage,omitempty"`
	ChatColor            string   `json:"chatColor"`
	ChatId               uint32   `json:"chatId"`
	ChatName             string   `json:"chatName"`
	ChatProfileImage     *string  `json:"chatProfileImage,omitempty"`
	ChatType             ChatType `json:"chatType"`
	Id                   uint32   `json:"id"`
	IsChatArchived       bool     `json:"isChatArchived"`
	IsChatContactRequest bool     `json:"isChatContactRequest"`
	Message              string   `json:"message"`
	Timestamp            int64    `json:"timestamp"`
}

type MuteDuration interface {
	isMuteDurationVariant()
	GetKind() string
}

type MuteDurationNotMuted struct {
}

func (*MuteDurationNotMuted) isMuteDurationVariant() {}
func (*MuteDurationNotMuted) GetKind() string        { return "NotMuted" }
func (v *MuteDurationNotMuted) MarshalJSON() ([]byte, error) {
	type alias MuteDurationNotMuted
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "NotMuted", alias: alias(*v)})
}

type MuteDurationForever struct {
}

func (*MuteDurationForever) isMuteDurationVariant() {}
func (*MuteDurationForever) GetKind() string        { return "Forever" }
func (v *MuteDurationForever) MarshalJSON() ([]byte, error) {
	type alias MuteDurationForever
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "Forever", alias: alias(*v)})
}

type MuteDurationUntil struct {
	Duration int64 `json:"duration"`
}

func (*MuteDurationUntil) isMuteDurationVariant() {}
func (*MuteDurationUntil) GetKind() string        { return "Until" }
func (v *MuteDurationUntil) MarshalJSON() ([]byte, error) {
	type alias MuteDurationUntil
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "Until", alias: alias(*v)})
}

type NotifyState string

const (
	// Not subscribed to push notifications.
	NotifyStateNotConnected NotifyState = "NotConnected"
	// Subscribed to heartbeat push notifications.
	NotifyStateHeartbeat NotifyState = "Heartbeat"
	// Subscribed to push notifications for new messages.
	NotifyStateConnected NotifyState = "Connected"
)

type ProviderInfo struct {
	BeforeLoginHint string `json:"beforeLoginHint"`
	// Unique ID, corresponding to provider database filename.
	Id           string `json:"id"`
	OverviewPage string `json:"overviewPage"`
	Status       uint32 `json:"status"`
}

type Qr interface {
	isQrVariant()
	GetKind() string
}

// Ask the user whether to verify the contact.
//
// If the user agrees, pass this QR code to [`crate::securejoin::join_securejoin`].
type QrAskVerifyContact struct {
	// Authentication code.
	Authcode string `json:"authcode"`
	// ID of the contact.
	Contact_id uint32 `json:"contact_id"`
	// Fingerprint of the contact key as scanned from the QR code.
	Fingerprint string `json:"fingerprint"`
	// Invite number.
	Invitenumber string `json:"invitenumber"`
}

func (*QrAskVerifyContact) isQrVariant()    {}
func (*QrAskVerifyContact) GetKind() string { return "askVerifyContact" }
func (v *QrAskVerifyContact) MarshalJSON() ([]byte, error) {
	type alias QrAskVerifyContact
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "askVerifyContact", alias: alias(*v)})
}

// Ask the user whether to join the group.
type QrAskVerifyGroup struct {
	// Authentication code.
	Authcode string `json:"authcode"`
	// ID of the contact.
	Contact_id uint32 `json:"contact_id"`
	// Fingerprint of the contact key as scanned from the QR code.
	Fingerprint string `json:"fingerprint"`
	// Group ID.
	Grpid string `json:"grpid"`
	// Group name.
	Grpname string `json:"grpname"`
	// Invite number.
	Invitenumber string `json:"invitenumber"`
}

func (*QrAskVerifyGroup) isQrVariant()    {}
func (*QrAskVerifyGroup) GetKind() string { return "askVerifyGroup" }
func (v *QrAskVerifyGroup) MarshalJSON() ([]byte, error) {
	type alias QrAskVerifyGroup
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "askVerifyGroup", alias: alias(*v)})
}

// Ask the user whether to join the broadcast channel.
type QrAskJoinBroadcast struct {
	// Authentication code.
	Authcode string `json:"authcode"`
	// ID of the contact who owns the broadcast channel and created the QR code.
	Contact_id uint32 `json:"contact_id"`
	// Fingerprint of the broadcast channel owner's key as scanned from the QR code.
	Fingerprint string `json:"fingerprint"`
	// A string of random characters, uniquely identifying this broadcast channel across all databases/clients. Called `grpid` for historic reasons: The id of multi-user chats is always called `grpid` in the database because groups were once the only multi-user chats.
	Grpid string `json:"grpid"`
	// Invite number.
	Invitenumber string `json:"invitenumber"`
	// The user-visible name of this broadcast channel
	Name string `json:"name"`
}

func (*QrAskJoinBroadcast) isQrVariant()    {}
func (*QrAskJoinBroadcast) GetKind() string { return "askJoinBroadcast" }
func (v *QrAskJoinBroadcast) MarshalJSON() ([]byte, error) {
	type alias QrAskJoinBroadcast
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "askJoinBroadcast", alias: alias(*v)})
}

// Contact fingerprint is verified.
//
// Ask the user if they want to start chatting.
type QrFprOk struct {
	// Contact ID.
	Contact_id uint32 `json:"contact_id"`
}

func (*QrFprOk) isQrVariant()    {}
func (*QrFprOk) GetKind() string { return "fprOk" }
func (v *QrFprOk) MarshalJSON() ([]byte, error) {
	type alias QrFprOk
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "fprOk", alias: alias(*v)})
}

// Scanned fingerprint does not match the last seen fingerprint.
type QrFprMismatch struct {
	// Contact ID.
	Contact_id *uint32 `json:"contact_id,omitempty"`
}

func (*QrFprMismatch) isQrVariant()    {}
func (*QrFprMismatch) GetKind() string { return "fprMismatch" }
func (v *QrFprMismatch) MarshalJSON() ([]byte, error) {
	type alias QrFprMismatch
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "fprMismatch", alias: alias(*v)})
}

// The scanned QR code contains a fingerprint but no e-mail address.
type QrFprWithoutAddr struct {
	// Key fingerprint.
	Fingerprint string `json:"fingerprint"`
}

func (*QrFprWithoutAddr) isQrVariant()    {}
func (*QrFprWithoutAddr) GetKind() string { return "fprWithoutAddr" }
func (v *QrFprWithoutAddr) MarshalJSON() ([]byte, error) {
	type alias QrFprWithoutAddr
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "fprWithoutAddr", alias: alias(*v)})
}

// Ask the user if they want to create an account on the given domain.
type QrAccount struct {
	// Server domain name.
	Domain string `json:"domain"`
}

func (*QrAccount) isQrVariant()    {}
func (*QrAccount) GetKind() string { return "account" }
func (v *QrAccount) MarshalJSON() ([]byte, error) {
	type alias QrAccount
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "account", alias: alias(*v)})
}

// Provides a backup that can be retrieved using iroh-net based backup transfer protocol.
type QrBackup2 struct {
	// Authentication token.
	Auth_token string `json:"auth_token"`
	// Iroh node address.
	Node_addr string `json:"node_addr"`
}

func (*QrBackup2) isQrVariant()    {}
func (*QrBackup2) GetKind() string { return "backup2" }
func (v *QrBackup2) MarshalJSON() ([]byte, error) {
	type alias QrBackup2
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "backup2", alias: alias(*v)})
}

type QrBackupTooNew struct {
}

func (*QrBackupTooNew) isQrVariant()    {}
func (*QrBackupTooNew) GetKind() string { return "backupTooNew" }
func (v *QrBackupTooNew) MarshalJSON() ([]byte, error) {
	type alias QrBackupTooNew
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "backupTooNew", alias: alias(*v)})
}

// Ask the user if they want to use the given service for video chats.
type QrWebrtcInstance struct {
	Domain           string `json:"domain"`
	Instance_pattern string `json:"instance_pattern"`
}

func (*QrWebrtcInstance) isQrVariant()    {}
func (*QrWebrtcInstance) GetKind() string { return "webrtcInstance" }
func (v *QrWebrtcInstance) MarshalJSON() ([]byte, error) {
	type alias QrWebrtcInstance
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "webrtcInstance", alias: alias(*v)})
}

// Ask the user if they want to use the given proxy.
//
// Note that HTTP(S) URLs without a path and query parameters are treated as HTTP(S) proxy URL. UI may want to still offer to open the URL in the browser if QR code contents starts with `http://` or `https://` and the QR code was not scanned from the proxy configuration screen.
type QrProxy struct {
	// Host extracted from the URL to display in the UI.
	Host string `json:"host"`
	// Port extracted from the URL to display in the UI.
	Port uint16 `json:"port"`
	// Proxy URL.
	//
	// This is the URL that is going to be added.
	Url string `json:"url"`
}

func (*QrProxy) isQrVariant()    {}
func (*QrProxy) GetKind() string { return "proxy" }
func (v *QrProxy) MarshalJSON() ([]byte, error) {
	type alias QrProxy
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "proxy", alias: alias(*v)})
}

// Contact address is scanned.
//
// Optionally, a draft message could be provided. Ask the user if they want to start chatting.
type QrAddr struct {
	// Contact ID.
	Contact_id uint32 `json:"contact_id"`
	// Draft message.
	Draft *string `json:"draft,omitempty"`
}

func (*QrAddr) isQrVariant()    {}
func (*QrAddr) GetKind() string { return "addr" }
func (v *QrAddr) MarshalJSON() ([]byte, error) {
	type alias QrAddr
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "addr", alias: alias(*v)})
}

// URL scanned.
//
// Ask the user if they want to open a browser or copy the URL to clipboard.
type QrUrl struct {
	Url string `json:"url"`
}

func (*QrUrl) isQrVariant()    {}
func (*QrUrl) GetKind() string { return "url" }
func (v *QrUrl) MarshalJSON() ([]byte, error) {
	type alias QrUrl
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "url", alias: alias(*v)})
}

// Text scanned.
//
// Ask the user if they want to copy the text to clipboard.
type QrText struct {
	Text string `json:"text"`
}

func (*QrText) isQrVariant()    {}
func (*QrText) GetKind() string { return "text" }
func (v *QrText) MarshalJSON() ([]byte, error) {
	type alias QrText
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "text", alias: alias(*v)})
}

// Ask the user if they want to withdraw their own QR code.
type QrWithdrawVerifyContact struct {
	// Authentication code.
	Authcode string `json:"authcode"`
	// Contact ID.
	Contact_id uint32 `json:"contact_id"`
	// Fingerprint of the contact key as scanned from the QR code.
	Fingerprint string `json:"fingerprint"`
	// Invite number.
	Invitenumber string `json:"invitenumber"`
}

func (*QrWithdrawVerifyContact) isQrVariant()    {}
func (*QrWithdrawVerifyContact) GetKind() string { return "withdrawVerifyContact" }
func (v *QrWithdrawVerifyContact) MarshalJSON() ([]byte, error) {
	type alias QrWithdrawVerifyContact
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "withdrawVerifyContact", alias: alias(*v)})
}

// Ask the user if they want to withdraw their own group invite QR code.
type QrWithdrawVerifyGroup struct {
	// Authentication code.
	Authcode string `json:"authcode"`
	// Contact ID.
	Contact_id uint32 `json:"contact_id"`
	// Fingerprint of the contact key as scanned from the QR code.
	Fingerprint string `json:"fingerprint"`
	// Group ID.
	Grpid string `json:"grpid"`
	// Group name.
	Grpname string `json:"grpname"`
	// Invite number.
	Invitenumber string `json:"invitenumber"`
}

func (*QrWithdrawVerifyGroup) isQrVariant()    {}
func (*QrWithdrawVerifyGroup) GetKind() string { return "withdrawVerifyGroup" }
func (v *QrWithdrawVerifyGroup) MarshalJSON() ([]byte, error) {
	type alias QrWithdrawVerifyGroup
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "withdrawVerifyGroup", alias: alias(*v)})
}

// Ask the user if they want to withdraw their own broadcast channel invite QR code.
type QrWithdrawJoinBroadcast struct {
	// Authentication code.
	Authcode string `json:"authcode"`
	// Contact ID. Always `ContactId::SELF`.
	Contact_id uint32 `json:"contact_id"`
	// Fingerprint of the contact key as scanned from the QR code.
	Fingerprint string `json:"fingerprint"`
	// ID, uniquely identifying this chat. Called grpid for historic reasons.
	Grpid string `json:"grpid"`
	// Invite number.
	Invitenumber string `json:"invitenumber"`
	// Broadcast name.
	Name string `json:"name"`
}

func (*QrWithdrawJoinBroadcast) isQrVariant()    {}
func (*QrWithdrawJoinBroadcast) GetKind() string { return "withdrawJoinBroadcast" }
func (v *QrWithdrawJoinBroadcast) MarshalJSON() ([]byte, error) {
	type alias QrWithdrawJoinBroadcast
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "withdrawJoinBroadcast", alias: alias(*v)})
}

// Ask the user if they want to revive their own QR code.
type QrReviveVerifyContact struct {
	// Authentication code.
	Authcode string `json:"authcode"`
	// Contact ID.
	Contact_id uint32 `json:"contact_id"`
	// Fingerprint of the contact key as scanned from the QR code.
	Fingerprint string `json:"fingerprint"`
	// Invite number.
	Invitenumber string `json:"invitenumber"`
}

func (*QrReviveVerifyContact) isQrVariant()    {}
func (*QrReviveVerifyContact) GetKind() string { return "reviveVerifyContact" }
func (v *QrReviveVerifyContact) MarshalJSON() ([]byte, error) {
	type alias QrReviveVerifyContact
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "reviveVerifyContact", alias: alias(*v)})
}

// Ask the user if they want to revive their own group invite QR code.
type QrReviveVerifyGroup struct {
	// Authentication code.
	Authcode string `json:"authcode"`
	// Contact ID.
	Contact_id uint32 `json:"contact_id"`
	// Fingerprint of the contact key as scanned from the QR code.
	Fingerprint string `json:"fingerprint"`
	// Group ID.
	Grpid string `json:"grpid"`
	// Contact ID.
	Grpname string `json:"grpname"`
	// Invite number.
	Invitenumber string `json:"invitenumber"`
}

func (*QrReviveVerifyGroup) isQrVariant()    {}
func (*QrReviveVerifyGroup) GetKind() string { return "reviveVerifyGroup" }
func (v *QrReviveVerifyGroup) MarshalJSON() ([]byte, error) {
	type alias QrReviveVerifyGroup
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "reviveVerifyGroup", alias: alias(*v)})
}

// Ask the user if they want to revive their own broadcast channel invite QR code.
type QrReviveJoinBroadcast struct {
	// Authentication code.
	Authcode string `json:"authcode"`
	// Contact ID. Always `ContactId::SELF`.
	Contact_id uint32 `json:"contact_id"`
	// Fingerprint of the contact key as scanned from the QR code.
	Fingerprint string `json:"fingerprint"`
	// Globally unique chat ID. Called grpid for historic reasons.
	Grpid string `json:"grpid"`
	// Invite number.
	Invitenumber string `json:"invitenumber"`
	// Broadcast name.
	Name string `json:"name"`
}

func (*QrReviveJoinBroadcast) isQrVariant()    {}
func (*QrReviveJoinBroadcast) GetKind() string { return "reviveJoinBroadcast" }
func (v *QrReviveJoinBroadcast) MarshalJSON() ([]byte, error) {
	type alias QrReviveJoinBroadcast
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "reviveJoinBroadcast", alias: alias(*v)})
}

// `dclogin:` scheme parameters.
//
// Ask the user if they want to login with the email address.
type QrLogin struct {
	Address string `json:"address"`
}

func (*QrLogin) isQrVariant()    {}
func (*QrLogin) GetKind() string { return "login" }
func (v *QrLogin) MarshalJSON() ([]byte, error) {
	type alias QrLogin
	return json.Marshal(struct {
		Kind string `json:"kind"`
		alias
	}{Kind: "login", alias: alias(*v)})
}

func unmarshalQr(data json.RawMessage, out *Qr) error {
	var header struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return err
	}
	switch header.Kind {
	case "askVerifyContact":
		var v QrAskVerifyContact
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "askVerifyGroup":
		var v QrAskVerifyGroup
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "askJoinBroadcast":
		var v QrAskJoinBroadcast
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "fprOk":
		var v QrFprOk
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "fprMismatch":
		var v QrFprMismatch
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "fprWithoutAddr":
		var v QrFprWithoutAddr
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "account":
		var v QrAccount
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "backup2":
		var v QrBackup2
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "backupTooNew":
		var v QrBackupTooNew
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "webrtcInstance":
		var v QrWebrtcInstance
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "proxy":
		var v QrProxy
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "addr":
		var v QrAddr
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "url":
		var v QrUrl
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "text":
		var v QrText
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "withdrawVerifyContact":
		var v QrWithdrawVerifyContact
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "withdrawVerifyGroup":
		var v QrWithdrawVerifyGroup
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "withdrawJoinBroadcast":
		var v QrWithdrawJoinBroadcast
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "reviveVerifyContact":
		var v QrReviveVerifyContact
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "reviveVerifyGroup":
		var v QrReviveVerifyGroup
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "reviveJoinBroadcast":
		var v QrReviveJoinBroadcast
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	case "login":
		var v QrLogin
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*out = &v
	default:
		return fmt.Errorf("unknown Qr variant: %q", header.Kind)
	}
	return nil
}

// A single reaction emoji.
type Reaction struct {
	// Emoji frequency.
	Count uint `json:"count"`
	// Emoji.
	Emoji string `json:"emoji"`
	// True if we reacted with this emoji.
	IsFromSelf bool `json:"isFromSelf"`
}

// Structure representing all reactions to a particular message.
type Reactions struct {
	// Unique reactions and their count, sorted in descending order.
	Reactions []Reaction `json:"reactions"`
	// Map from a contact to it's reaction to message.
	ReactionsByContact map[string][]string `json:"reactionsByContact"`
}

type SecurejoinSource string

const (
	// Because of some problem, it is unknown where the QR code came from.
	SecurejoinSourceUnknown SecurejoinSource = "Unknown"
	// The user opened a link somewhere outside Delta Chat
	SecurejoinSourceExternalLink SecurejoinSource = "ExternalLink"
	// The user clicked on a link in a message inside Delta Chat
	SecurejoinSourceInternalLink SecurejoinSource = "InternalLink"
	// The user clicked "Paste from Clipboard" in the QR scan activity
	SecurejoinSourceClipboard SecurejoinSource = "Clipboard"
	// The user clicked "Load QR code as image" in the QR scan activity
	SecurejoinSourceImageLoaded SecurejoinSource = "ImageLoaded"
	// The user scanned a QR code
	SecurejoinSourceScan SecurejoinSource = "Scan"
)

type SecurejoinUiPath string

const (
	// The UI path is unknown, or the user didn't open the QR code screen at all.
	SecurejoinUiPathUnknown SecurejoinUiPath = "Unknown"
	// The user directly clicked on the QR icon in the main screen
	SecurejoinUiPathQrIcon SecurejoinUiPath = "QrIcon"
	// The user first clicked on the `+` button in the main screen, and then on "New Contact"
	SecurejoinUiPathNewContact SecurejoinUiPath = "NewContact"
)

type Socket string

const (
	// Unspecified socket security, select automatically.
	SocketAutomatic Socket = "automatic"
	// TLS connection.
	SocketSsl Socket = "ssl"
	// STARTTLS connection.
	SocketStarttls Socket = "starttls"
	// No TLS, plaintext connection.
	SocketPlain Socket = "plain"
)

type SystemMessageType string

const (
	SystemMessageTypeUnknown                  SystemMessageType = "Unknown"
	SystemMessageTypeGroupNameChanged         SystemMessageType = "GroupNameChanged"
	SystemMessageTypeGroupDescriptionChanged  SystemMessageType = "GroupDescriptionChanged"
	SystemMessageTypeGroupImageChanged        SystemMessageType = "GroupImageChanged"
	SystemMessageTypeMemberAddedToGroup       SystemMessageType = "MemberAddedToGroup"
	SystemMessageTypeMemberRemovedFromGroup   SystemMessageType = "MemberRemovedFromGroup"
	SystemMessageTypeAutocryptSetupMessage    SystemMessageType = "AutocryptSetupMessage"
	SystemMessageTypeSecurejoinMessage        SystemMessageType = "SecurejoinMessage"
	SystemMessageTypeLocationStreamingEnabled SystemMessageType = "LocationStreamingEnabled"
	SystemMessageTypeLocationOnly             SystemMessageType = "LocationOnly"
	SystemMessageTypeInvalidUnencryptedMail   SystemMessageType = "InvalidUnencryptedMail"
	SystemMessageTypeChatE2ee                 SystemMessageType = "ChatE2ee"
	SystemMessageTypeChatProtectionEnabled    SystemMessageType = "ChatProtectionEnabled"
	SystemMessageTypeChatProtectionDisabled   SystemMessageType = "ChatProtectionDisabled"
	SystemMessageTypeWebxdcStatusUpdate       SystemMessageType = "WebxdcStatusUpdate"
	SystemMessageTypeCallAccepted             SystemMessageType = "CallAccepted"
	SystemMessageTypeCallEnded                SystemMessageType = "CallEnded"
	// 1:1 chats info message telling that SecureJoin has started and the user should wait for it to complete.
	SystemMessageTypeSecurejoinWait SystemMessageType = "SecurejoinWait"
	// 1:1 chats info message telling that SecureJoin is still running, but the user may already send messages.
	SystemMessageTypeSecurejoinWaitTimeout SystemMessageType = "SecurejoinWaitTimeout"
	// Chat ephemeral message timer is changed.
	SystemMessageTypeEphemeralTimerChanged SystemMessageType = "EphemeralTimerChanged"
	// Self-sent-message that contains only json used for multi-device-sync; if possible, we attach that to other messages as for locations.
	SystemMessageTypeMultiDeviceSync SystemMessageType = "MultiDeviceSync"
	// Webxdc info added with `info` set in `send_webxdc_status_update()`.
	SystemMessageTypeWebxdcInfoMessage SystemMessageType = "WebxdcInfoMessage"
	// This message contains a users iroh node address.
	SystemMessageTypeIrohNodeAddr SystemMessageType = "IrohNodeAddr"
)

type VcardContact struct {
	// Email address.
	Addr string `json:"addr"`
	// Contact color as hex string.
	Color string `json:"color"`
	// The contact's name, or the email address if no name was given.
	DisplayName string `json:"displayName"`
	// Public PGP key in Base64.
	Key *string `json:"key,omitempty"`
	// Profile image in Base64.
	ProfileImage *string `json:"profileImage,omitempty"`
	// Last update timestamp.
	Timestamp *int64 `json:"timestamp,omitempty"`
}

type Viewtype string

const (
	ViewtypeUnknown Viewtype = "Unknown"
	// Text message.
	ViewtypeText Viewtype = "Text"
	// Image message. If the image is an animated GIF, the type `Viewtype.Gif` should be used.
	ViewtypeImage Viewtype = "Image"
	// Animated GIF message.
	ViewtypeGif Viewtype = "Gif"
	// Message containing a sticker, similar to image. NB: When sending, the message viewtype may be changed to `Image` by some heuristics like checking for transparent pixels. Use `Message::force_sticker()` to disable them.
	//
	// If possible, the ui should display the image without borders in a transparent way. A click on a sticker will offer to install the sticker set in some future.
	ViewtypeSticker Viewtype = "Sticker"
	// Message containing an Audio file.
	ViewtypeAudio Viewtype = "Audio"
	// A voice message that was directly recorded by the user. For all other audio messages, the type `Viewtype.Audio` should be used.
	ViewtypeVoice Viewtype = "Voice"
	// Video messages.
	ViewtypeVideo Viewtype = "Video"
	// Message containing any file, eg. a PDF.
	ViewtypeFile Viewtype = "File"
	// Message is a call.
	ViewtypeCall Viewtype = "Call"
	// Message is an webxdc instance.
	ViewtypeWebxdc Viewtype = "Webxdc"
	// Message containing shared contacts represented as a vCard (virtual contact file) with email addresses and possibly other fields. Use `parse_vcard()` to retrieve them.
	ViewtypeVcard Viewtype = "Vcard"
)

type WebxdcMessageInfo struct {
	// if the Webxdc represents a document, then this is the name of the document
	Document *string `json:"document,omitempty"`
	// App icon file name. Defaults to an standard icon if nothing is set in the manifest.
	//
	// To get the file, use dc_msg_get_webxdc_blob(). (not yet in jsonrpc, use rust api or cffi for it)
	//
	// App icons should should be square, the implementations will add round corners etc. as needed.
	Icon string `json:"icon"`
	// True if full internet access should be granted to the app.
	InternetAccess bool `json:"internetAccess"`
	// The name of the app.
	//
	// Defaults to the filename if not set in the manifest.
	Name string `json:"name"`
	// Address to be used for `window.webxdc.selfAddr` in JS land.
	SelfAddr string `json:"selfAddr"`
	// Milliseconds to wait before calling `sendUpdate()` again since the last call. Should be exposed to `window.sendUpdateInterval` in JS land.
	SendUpdateInterval uint `json:"sendUpdateInterval"`
	// Maximum number of bytes accepted for a serialized update object. Should be exposed to `window.sendUpdateMaxSize` in JS land.
	SendUpdateMaxSize uint `json:"sendUpdateMaxSize"`
	// URL where the source code of the Webxdc and other information can be found; defaults to an empty string. Implementations may offer an menu or a button to open this URL.
	SourceCodeUrl *string `json:"sourceCodeUrl,omitempty"`
	// short string describing the state of the app, sth. as "2 votes", "Highscore: 123", can be changed by the apps
	Summary *string `json:"summary,omitempty"`
}
