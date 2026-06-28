package chatsvc

import (
	"context"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/chat"
)

type ChatThreadRepository interface {
	Find(ctx context.Context, id chat.ChatThreadID) (*chat.ChatThread, error)
	Store(ctx context.Context, entity *chat.ChatThread) error
	Remove(ctx context.Context, ids ...chat.ChatThreadID) error
}

type ChatThreadService interface {
	List(ctx context.Context, cond chat.ChatThreadListCond) ([]*chat.ChatThread, error)
	Count(ctx context.Context, cond chat.ChatThreadCountCond) (*uint64, error)
}

type ChatMessageRepository interface {
	Find(ctx context.Context, id chat.ChatMessageID) (*chat.ChatMessage, error)
	Store(ctx context.Context, entity *chat.ChatMessage) error
	Remove(ctx context.Context, ids ...chat.ChatMessageID) error
}

type ChatMessageService interface {
	List(ctx context.Context, cond chat.ChatMessageListCond) ([]*chat.ChatMessage, error)
	Count(ctx context.Context, cond chat.ChatMessageCountCond) (*uint64, error)
}

type ChatThreadParticipantRepository interface {
	Find(ctx context.Context, chatThreadID chat.ChatThreadID, userID chat.UserID) (*chat.ChatThreadParticipant, error)
	Store(ctx context.Context, entity *chat.ChatThreadParticipant) error
	Remove(ctx context.Context, conds ...RemoveParticipantCond) error
}

type ChatThreadParticipantService interface {
	List(ctx context.Context, cond chat.ChatThreadParticipantListCond) ([]*chat.ChatThreadParticipant, error)
	Count(ctx context.Context, cond chat.ChatThreadParticipantCountCond) (*uint64, error)
}

type ChatMessageAttachmentRepository interface {
	Find(ctx context.Context, id chat.ChatMessageAttachmentID) (*chat.ChatMessageAttachment, error)
	Store(ctx context.Context, entity *chat.ChatMessageAttachment) error
	Remove(ctx context.Context, ids ...chat.ChatMessageAttachmentID) error
}

type ChatMessageAttachmentService interface {
	List(ctx context.Context, cond chat.ChatMessageAttachmentListCond) ([]*chat.ChatMessageAttachment, error)
	Count(ctx context.Context, cond chat.ChatMessageAttachmentCountCond) (*uint64, error)
}

type RemoveParticipantCond struct {
	ChatThreadID chat.ChatThreadID
	UserID       chat.UserID
}
