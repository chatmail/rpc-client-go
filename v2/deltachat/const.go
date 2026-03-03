package deltachat

type ChatListFlag uint

type ContactFlag uint

type MsgState uint

const (
	//Special contact ids
	ContactSelf        uint32 = 1
	ContactInfo        uint32 = 2
	ContactDevice      uint32 = 5
	ContactLastSpecial uint32 = 9

	// Chatlist Flags
	ChatListFlagArchivedOnly   ChatListFlag = 0x01
	ChatListFlagNoSpecials     ChatListFlag = 0x02
	ChatListFlagAddAlldoneHint ChatListFlag = 0x04
	ChatListFlagForForwarding  ChatListFlag = 0x08

	// Contact Flags
	ContactFlagVerifiedOnly ContactFlag = 0x01
	ContactFlagAddSelf      ContactFlag = 0x02

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
)
