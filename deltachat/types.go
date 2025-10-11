package deltachat

import (
	"github.com/chatmail/rpc-client-go/deltachat/option"
)

type AccountId uint64

type ContactId uint64

type MsgId uint64

type ChatId uint64

// Values in const.go
type ChatType string

type Account struct {
	// Configured
	Id           AccountId
	DisplayName  option.Option[string]
	Addr         option.Option[string]
	ProfileImage option.Option[string]
	Color        string

	// Unconfigured
	// Id AccountId
}

// Delta Chat Contact snapshot.
type ContactSnapshot struct {
	Address           string
	Color             string
	AuthName          string
	Status            string
	DisplayName       string
	Id                ContactId
	Name              string
	ProfileImage      string
	NameAndAddr       string
	IsBlocked         bool
	IsVerified        bool
	IsProfileVerified bool
	VerifierId        ContactId
	LastSeen          Timestamp
	WasSeenRecently   bool
	IsBot             bool
}

// Full chat snapshot.
type FullChatSnapshot struct {
	Id                  ChatId
	Name                string
	IsProtected         bool
	ProfileImage        string
	Archived            bool
	ChatType            ChatType
	IsUnpromoted        bool
	IsSelfTalk          bool
	Contacts            []*ContactSnapshot
	ContactIds          []ContactId
	Color               string
	FreshMessageCounter uint
	IsContactRequest    bool
	IsProtectionBroken  bool
	IsDeviceChat        bool
	SelfInGroup         bool
	IsMuted             bool
	EphemeralTimer      uint
	CanSend             bool
	WasSeenRecently     bool
	MailingListAddress  string
}

// Cheaper version of FullChatSnapshot.
type BasicChatSnapshot struct {
	Id                 ChatId
	Name               string
	IsProtected        bool
	ProfileImage       string
	Archived           bool
	ChatType           ChatType
	IsUnpromoted       bool
	IsSelfTalk         bool
	Color              string
	IsContactRequest   bool
	IsProtectionBroken bool
	IsDeviceChat       bool
	IsMuted            bool
}

// Chat list item snapshot
type ChatListItem struct {
	Id                  ChatId
	Name                string
	AvatarPath          string
	Color               string
	LastUpdated         Timestamp
	SummaryText1        string
	SummaryText2        string
	SummaryStatus       uint32
	IsProtected         bool
	IsGroup             bool
	FreshMessageCounter uint
	IsSelfTalk          bool
	IsDeviceTalk        bool
	IsSendingLocation   bool
	IsSelfInGroup       bool
	IsArchived          bool
	IsPinned            bool
	IsMuted             bool
	IsContactRequest    bool
	IsBroadcast         bool
	DmChatContact       ContactId
	WasSeenRecently     bool

	// ArchiveLink
	// FreshMessageCounter uint

	// Error
	// Id uint64
	Error string
}

// Message data provided to Chat.SendMsg()
type MsgData struct {
	Text               string      `json:"text,omitempty"`
	Html               string      `json:"html,omitempty"`
	ViewType           MsgType     `json:"viewtype,omitempty"`
	File               string      `json:"file,omitempty"`
	Location           *[2]float64 `json:"location,omitempty"`
	OverrideSenderName string      `json:"overrideSenderName,omitempty"`
	QuotedMessageId    MsgId       `json:"quotedMessageId,omitempty"`
}

// Message quote. Only the Text property is warrantied to be present, all other fields are optional.
type MsgQuote struct {
	Text               string
	MessageId          MsgId
	AuthorDisplayName  string
	AuthorDisplayColor string
	OverrideSenderName string
	Image              string
	IsForwarded        bool
	ViewType           MsgType
}

// Message search result.
type MsgSearchResult struct {
	Id                 MsgId
	AuthorProfileImage string
	AuthorName         string
	AuthorColor        string
	ChatName           string
	Message            string
	Timestamp          Timestamp
}

// Message snapshot.
type MsgSnapshot struct {
	Id                    MsgId
	ChatId                ChatId
	FromId                ContactId
	Quote                 *MsgQuote
	ParentId              MsgId
	Text                  string
	HasLocation           bool
	HasHtml               bool
	ViewType              MsgType
	State                 MsgState
	Error                 string
	Timestamp             Timestamp
	SortTimestamp         Timestamp
	ReceivedTimestamp     Timestamp
	HasDeviatingTimestamp bool
	Subject               string
	ShowPadlock           bool
	IsSetupmessage        bool
	IsInfo                bool
	IsForwarded           bool
	IsBot                 bool
	SystemMessageType     SysmsgType
	Duration              int
	DimensionsHeight      int
	DimensionsWidth       int
	VideochatType         int
	VideochatUrl          string
	OverrideSenderName    string
	Sender                *ContactSnapshot
	SetupCodeBegin        string
	File                  string
	FileMime              string
	FileBytes             uint64
	FileName              string
	WebxdcInfo            *WebxdcMsgInfo
	DownloadState         DownloadState
	Reactions             *Reactions
}

type WebxdcMsgInfo struct {
	Name           string
	Icon           string
	Document       string
	Summary        string
	SourceCodeUrl  string
	InternetAccess bool
}

type Reaction struct {
	Emoji      string
	Count      uint32
	IsFromSelf bool
}

type Reactions struct {
	ReactionsByContact map[ContactId][]string
	Reactions          []Reaction
}
