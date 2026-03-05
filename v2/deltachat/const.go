package deltachat

const (
	//Special contact ids
	ContactSelf        uint32 = 1
	ContactInfo        uint32 = 2
	ContactDevice      uint32 = 5
	ContactLastSpecial uint32 = 9

	// Chatlist Flags
	ChatListFlagArchivedOnly   uint32 = 0x01
	ChatListFlagNoSpecials     uint32 = 0x02
	ChatListFlagAddAlldoneHint uint32 = 0x04
	ChatListFlagForForwarding  uint32 = 0x08

	// Contact Flags
	ContactFlagVerifiedOnly uint32 = 0x01
	ContactFlagAddSelf      uint32 = 0x02
	ContactFlagAddress      uint32 = 0x04

	// Message State
	MsgStateUndefined    uint32 = 0  // Message just created.
	MsgStateInFresh      uint32 = 10 // Incoming fresh message.
	MsgStateInNoticed    uint32 = 13 // Incoming noticed message.
	MsgStateInSeen       uint32 = 16 // Incoming seen message.
	MsgStateOutPreparing uint32 = 18 // Outgoing message being prepared.
	MsgStateOutDraft     uint32 = 19 // Outgoing message drafted.
	MsgStateOutPending   uint32 = 20 // Outgoing message waiting to be sent.
	MsgStateOutFailed    uint32 = 24 // Outgoing message failed sending.
	MsgStateOutDelivered uint32 = 26 // Outgoing message sent.
	MsgStateOutMdnRcvd   uint32 = 28 // Outgoing message sent and seen by recipients(s).
)
