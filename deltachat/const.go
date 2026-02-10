package deltachat

type ChatListFlag uint

type ContactFlag uint

type ChatVisibility string

type DownloadState string

type MsgType string

type MsgState uint

type SysmsgType string

const (
	//Special contact ids
	ContactSelf        ContactId = 1
	ContactInfo        ContactId = 2
	ContactDevice      ContactId = 5
	ContactLastSpecial ContactId = 9

	// Chatlist Flags
	ChatListFlagArchivedOnly   ChatListFlag = 0x01
	ChatListFlagNoSpecials     ChatListFlag = 0x02
	ChatListFlagAddAlldoneHint ChatListFlag = 0x04
	ChatListFlagForForwarding  ChatListFlag = 0x08

	// Contact Flags
	ContactFlagVerifiedOnly ContactFlag = 0x01
	ContactFlagAddSelf      ContactFlag = 0x02

	//Chat types
	ChatUndefined    ChatType = ""
	ChatSingle       ChatType = "Single"
	ChatGroup        ChatType = "Group"
	ChatMailinglist  ChatType = "MailingList"
	ChatOutBroadcast ChatType = "OutBroadcast"
	ChatInBroadcast  ChatType = "InBroadcast"

	// Chat visibility types
	ChatVisibilityNormal   ChatVisibility = "Normal"
	ChatVisibilityArchived ChatVisibility = "Archived"
	ChatVisibilityPinned   ChatVisibility = "Pinned"

	//Message download states
	DownloadDone       DownloadState = "Done"
	DownloadAvailable  DownloadState = "Available"
	DownloadFailure    DownloadState = "Failure"
	DownloadInProgress DownloadState = "InProgress"

	// Message State
	MsgStateUndefined    MsgState = 0  // Message just created.
	MsgStateInFresh      MsgState = 10 // Incoming fresh message.
	MsgStateInNoticed    MsgState = 13 // Incoming noticed message.
	MsgStateInSeen       MsgState = 16 // Incoming seen message.
	MsgStateOutPreparing MsgState = 18 // Outgoing message being prepared.
	MsgStateOutDraft     MsgState = 19 // Outgoing message drafted.
	MsgStateOutPending   MsgState = 20 // Outgoing message waiting to be sent.
	MsgStateOutFailed    MsgState = 24 // Outgoing message failed sending.
	MsgStateOutDelivered MsgState = 26 // Outgoing message sent.
	MsgStateOutMdnRcvd   MsgState = 28 // Outgoing message sent and seen by recipients(s).

	//Message view types
	MsgUnknown             MsgType = "Unknown"
	MsgText                MsgType = "Text"
	MsgImage               MsgType = "Image"
	MsgGif                 MsgType = "Gif"
	MsgSticker             MsgType = "Sticker"
	MsgAudio               MsgType = "Audio"
	MsgVoice               MsgType = "Voice"
	MsgVideo               MsgType = "Video"
	MsgFile                MsgType = "File"
	MsgVideochatInvitation MsgType = "VideochatInvitation"
	MsgWebxdc              MsgType = "Webxdc"
	MsgVcard               MsgType = "Vcard"

	//System message types
	SysmsgUnknown                  SysmsgType = "Unknown"
	SysmsgGroupNameChanged         SysmsgType = "GroupNameChanged"
	SysmsgGroupImageChanged        SysmsgType = "GroupImageChanged"
	SysmsgMemberAddedToGroup       SysmsgType = "MemberAddedToGroup"
	SysmsgMemberRemovedFromGroup   SysmsgType = "MemberRemovedFromGroup"
	SysmsgAutocryptSetupMessage    SysmsgType = "AutocryptSetupMessage"
	SysmsgSecurejoinMessage        SysmsgType = "SecurejoinMessage"
	SysmsgLocationStreamingEnabled SysmsgType = "LocationStreamingEnabled"
	SysmsgLocationOnly             SysmsgType = "LocationOnly"
	SysmsgChatProtectionEnabled    SysmsgType = "ChatProtectionEnabled"
	SysmsgChatProtectionDisabled   SysmsgType = "ChatProtectionDisabled"
	SysmsgWebxdcStatusUpdate       SysmsgType = "WebxdcStatusUpdate"
	SysmsgEphemeralTimerChanged    SysmsgType = "EphemeralTimerChanged"
	SysmsgMultiDeviceSync          SysmsgType = "MultiDeviceSync"
	SysmsgWebxdcInfoMessage        SysmsgType = "WebxdcInfoMessage"
)
