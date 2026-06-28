package chat

import (
	"time"

	"github.com/google/uuid"
)

type ChatThread struct {
	ChatThreadID  ChatThreadID
	Type          ChatThreadsType
	Title         *string
	LastMessageID *string
	LastMessageAt *time.Time
	CreatedAt     time.Time
	UpdatedAt     *time.Time
}

type ChatThreadID string

type ChatThreadsType string

const (
	ChatThreadsType_Direct ChatThreadsType = "direct"
	ChatThreadsType_Group  ChatThreadsType = "group"
)

var ChatThreadsTypeAllValues = []ChatThreadsType{
	ChatThreadsType_Direct,
	ChatThreadsType_Group,
}

func NewChatThreadID() (ChatThreadID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return ChatThreadID(id.String()), nil
}

type ChatMessage struct {
	ChatMessageID ChatMessageID
	ChatThreadID  ChatThreadID
	SenderUserID  UserID
	MessageType   ChatMessageMessageType
	Body          *string
	CreatedAt     time.Time
	UpdatedAt     *time.Time
	DeletedAt     *time.Time
}

type ChatMessageID string

type ChatMessageMessageType string

const (
	ChatMessageMessageType_Text       ChatMessageMessageType = "text"
	ChatMessageMessageType_Image      ChatMessageMessageType = "image"
	ChatMessageMessageType_Attachment ChatMessageMessageType = "attachment"
)

var ChatMessageMessageTypeAllValues = []ChatMessageMessageType{
	ChatMessageMessageType_Text,
	ChatMessageMessageType_Image,
	ChatMessageMessageType_Attachment,
}

func NewChatMessageID() (ChatMessageID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return ChatMessageID(id.String()), nil
}

type ChatThreadParticipant struct {
	ChatThreadID      ChatThreadID
	UserID            UserID
	ArchivedAt        *time.Time
	MutedAt           *time.Time
	LastReadMessageID *ChatMessageID
	LastReadAt        *time.Time
	JoinedAt          time.Time
	LeftAt            *time.Time
}

type UserID string

type ChatMessageAttachment struct {
	ChatMessageAttachmentID ChatMessageAttachmentID
	ChatMessageID           ChatMessageID
	FileName                string
	FileType                *string
	FileSize                *int64
	R2Bucket                string
	R2ObjectKey             string
	PublicURL               *string
	CreatedAt               time.Time
}

type ChatMessageAttachmentID string

func NewChatMessageAttachmentID() (ChatMessageAttachmentID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return ChatMessageAttachmentID(id.String()), nil
}

type ChatThreadListCond struct {
	ChatThreadID *ChatThreadID
	Title        *string
	CreatedAt    *time.Time

	SortKey ChatThreadSortKey
	Size    int64
	Offset  *int64
}

type ChatThreadCountCond ChatThreadListCond

type ChatThreadSortKey uint

const (
	ChatThreadSortKey_CreatedAt_ASC ChatThreadSortKey = iota
	ChatThreadSortKey_CreatedAt_DESC
)

type ChatMessageListCond struct {
	ChatMessageID *ChatMessageID
	ChatThreadID  *ChatThreadID
	SenderUserID  *UserID
	MessageType   *ChatMessageMessageType
	Body          *string

	SortKey ChatMessageSortKey
	Size    int64
	Offset  *int64
}

type ChatMessageCountCond ChatMessageListCond

type ChatMessageSortKey uint

const (
	ChatMessageSortKey_CreatedAt_ASC ChatMessageSortKey = iota
	ChatMessageSortKey_CreatedAt_DESC
)

type ChatThreadParticipantListCond struct {
	ChatThreadID *ChatThreadID
	UserID       *UserID
	JoinedAt     *time.Time
	LeftAt       *time.Time

	SortKey ChatThreadParticipantSortKey
	Size    int64
	Offset  *int64
}

type ChatThreadParticipantCountCond ChatThreadParticipantListCond

type ChatThreadParticipantSortKey uint

const (
	ChatThreadParticipantSortKey_JoinedAt_ASC ChatThreadParticipantSortKey = iota
	ChatThreadParticipantSortKey_JoinedAt_DESC
)

type ChatMessageAttachmentListCond struct {
	ChatMessageAttachmentID *ChatMessageAttachmentID
	ChatMessageID           *ChatMessageID

	SortKey ChatMessageAttachmentSortKey
	Size    int64
	Offset  *int64
}

type ChatMessageAttachmentCountCond ChatMessageAttachmentListCond

type ChatMessageAttachmentSortKey uint

const (
	ChatMessageAttachmentSortKey_CreatedAt_ASC ChatMessageAttachmentSortKey = iota
	ChatMessageAttachmentSortKey_CreatedAt_DESC
)
