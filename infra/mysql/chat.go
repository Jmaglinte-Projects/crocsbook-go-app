package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/chat"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql/lib/db_crocs/model"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql/lib/db_crocs/table"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/chatsvc"
	jet "github.com/go-jet/jet/v2/mysql"
	"github.com/go-sql-driver/mysql"
)

type chatThreadRepository struct {
	db *sql.DB
}

func NewChatThreadRepository(db *sql.DB) chatsvc.ChatThreadRepository {
	return &chatThreadRepository{db: db}
}

func (r *chatThreadRepository) Find(ctx context.Context, id chat.ChatThreadID) (*chat.ChatThread, error) {
	stmt := table.ChatThreads.SELECT(table.ChatThreads.AllColumns).WHERE(
		table.ChatThreads.ChatThreadID.EQ(jet.String(string(id))))

	dest := &ChatThreadModels{}
	if err := stmt.Query(r.db, dest); err != nil {
		return nil, err
	}

	if len(*dest) == 0 {
		return nil, nil
	}

	return dest.Unmarshal()[0], nil
}

func (r *chatThreadRepository) Store(ctx context.Context, entity *chat.ChatThread) error {
	threadType := model.ChatThreadsType(entity.Type)
	m := model.ChatThreads{
		ChatThreadID:  string(entity.ChatThreadID),
		Type:          threadType,
		Title:         entity.Title,
		LastMessageID: entity.LastMessageID,
		LastMessageAt: entity.LastMessageAt,
		CreatedAt:     entity.CreatedAt,
		UpdatedAt:     entity.UpdatedAt,
	}

	insertStmt := table.ChatThreads.INSERT(table.ChatThreads.AllColumns).MODEL(m)
	updateStmt := table.ChatThreads.UPDATE(table.ChatThreads.AllColumns).MODEL(m)
	updateStmt = updateStmt.WHERE(table.ChatThreads.ChatThreadID.EQ(jet.String(string(entity.ChatThreadID))))

	_, err := insertStmt.Exec(r.db)
	if err != nil {
		if mysqlerr, ok := err.(*mysql.MySQLError); ok {
			switch mysqlerr.Number {
			case 1062:
				result, err := updateStmt.Exec(r.db)
				if err != nil {
					return err
				}
				rowsAffected, err := result.RowsAffected()
				if err != nil {
					return err
				}
				if rowsAffected == 0 {
					return fmt.Errorf("entity version conflicted")
				}
			default:
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func (r *chatThreadRepository) Remove(ctx context.Context, ids ...chat.ChatThreadID) error {
	idExpressions := make([]jet.Expression, 0, len(ids))
	for _, id := range ids {
		idExpressions = append(idExpressions, jet.String(string(id)))
	}

	stmt := table.ChatThreads.DELETE().WHERE(table.ChatThreads.ChatThreadID.IN(idExpressions...))
	_, err := stmt.Exec(r.db)
	return err
}

type chatThreadService struct {
	db *sql.DB
}

func NewChatThreadService(db *sql.DB) chatsvc.ChatThreadService {
	return &chatThreadService{db: db}
}

func (s *chatThreadService) List(ctx context.Context, cond chat.ChatThreadListCond) ([]*chat.ChatThread, error) {
	stmt := table.ChatThreads.SELECT(table.ChatThreads.AllColumns)
	pred := []jet.BoolExpression{}
	orderBy := []jet.OrderByClause{}

	if cond.ChatThreadID != nil {
		pred = append(pred, table.ChatThreads.ChatThreadID.EQ(jet.String(string(*cond.ChatThreadID))))
	}

	if cond.Title != nil {
		pred = append(pred, table.ChatThreads.Title.EQ(jet.String(*cond.Title)))
	}

	if cond.CreatedAt != nil {
		pred = append(pred, table.ChatThreads.CreatedAt.EQ(jet.TimestampT(*cond.CreatedAt)))
	}

	switch cond.SortKey {
	case chat.ChatThreadSortKey_CreatedAt_ASC:
		orderBy = append(orderBy, table.ChatThreads.CreatedAt.ASC())
	case chat.ChatThreadSortKey_CreatedAt_DESC:
		orderBy = append(orderBy, table.ChatThreads.CreatedAt.DESC())
	default:
		orderBy = append(orderBy, table.ChatThreads.CreatedAt.DESC())
	}

	if len(pred) > 0 {
		stmt = stmt.WHERE(jet.AND(pred...))
	}

	stmt = stmt.ORDER_BY(orderBy...)

	if cond.Offset != nil {
		stmt = stmt.OFFSET(*cond.Offset)
	}

	if cond.Size > 0 {
		stmt = stmt.LIMIT(cond.Size)
	}

	dest := &ChatThreadModels{}
	if err := stmt.Query(s.db, dest); err != nil {
		return nil, err
	}

	if len(*dest) == 0 {
		return nil, nil
	}

	return dest.Unmarshal(), nil
}

func (s *chatThreadService) Count(ctx context.Context, cond chat.ChatThreadCountCond) (*uint64, error) {
	stmt := table.ChatThreads.SELECT(jet.COUNT(table.ChatThreads.ChatThreadID).AS("count"))
	pred := []jet.BoolExpression{}

	if cond.ChatThreadID != nil {
		pred = append(pred, table.ChatThreads.ChatThreadID.EQ(jet.String(string(*cond.ChatThreadID))))
	}

	if cond.Title != nil {
		pred = append(pred, table.ChatThreads.Title.EQ(jet.String(*cond.Title)))
	}

	if cond.CreatedAt != nil {
		pred = append(pred, table.ChatThreads.CreatedAt.EQ(jet.TimestampT(*cond.CreatedAt)))
	}

	if len(pred) > 0 {
		stmt = stmt.WHERE(jet.AND(pred...))
	}

	var dest []struct {
		Count uint64
	}

	if err := stmt.QueryContext(ctx, s.db, &dest); err != nil {
		return nil, err
	}

	return &dest[0].Count, nil
}

type chatMessageRepository struct {
	db *sql.DB
}

func NewChatMessageRepository(db *sql.DB) chatsvc.ChatMessageRepository {
	return &chatMessageRepository{db: db}
}

func (r *chatMessageRepository) Find(ctx context.Context, id chat.ChatMessageID) (*chat.ChatMessage, error) {
	stmt := table.ChatMessages.SELECT(table.ChatMessages.AllColumns).WHERE(
		table.ChatMessages.ChatMessageID.EQ(jet.String(string(id))))

	dest := &ChatMessageModels{}
	if err := stmt.Query(r.db, dest); err != nil {
		return nil, err
	}

	if len(*dest) == 0 {
		return nil, nil
	}

	return dest.Unmarshal()[0], nil
}

func (r *chatMessageRepository) Store(ctx context.Context, entity *chat.ChatMessage) error {
	messageType := model.ChatMessagesMessageType(entity.MessageType)
	m := model.ChatMessages{
		ChatMessageID: string(entity.ChatMessageID),
		ChatThreadID:  string(entity.ChatThreadID),
		SenderUserID:  string(entity.SenderUserID),
		MessageType:   messageType,
		Body:          entity.Body,
		CreatedAt:     entity.CreatedAt,
		UpdatedAt:     entity.UpdatedAt,
		DeletedAt:     entity.DeletedAt,
	}

	insertStmt := table.ChatMessages.INSERT(table.ChatMessages.AllColumns).MODEL(m)
	updateStmt := table.ChatMessages.UPDATE(table.ChatMessages.AllColumns).MODEL(m)
	updateStmt = updateStmt.WHERE(table.ChatMessages.ChatMessageID.EQ(jet.String(string(entity.ChatMessageID))))

	_, err := insertStmt.Exec(r.db)
	if err != nil {
		if mysqlerr, ok := err.(*mysql.MySQLError); ok {
			switch mysqlerr.Number {
			case 1062:
				result, err := updateStmt.Exec(r.db)
				if err != nil {
					return err
				}
				rowsAffected, err := result.RowsAffected()
				if err != nil {
					return err
				}
				if rowsAffected == 0 {
					return fmt.Errorf("entity version conflicted")
				}
			default:
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func (r *chatMessageRepository) Remove(ctx context.Context, ids ...chat.ChatMessageID) error {
	idExpressions := make([]jet.Expression, 0, len(ids))
	for _, id := range ids {
		idExpressions = append(idExpressions, jet.String(string(id)))
	}

	stmt := table.ChatMessages.DELETE().WHERE(table.ChatMessages.ChatMessageID.IN(idExpressions...))
	_, err := stmt.Exec(r.db)
	return err
}

type chatMessageService struct {
	db *sql.DB
}

func NewChatMessageService(db *sql.DB) chatsvc.ChatMessageService {
	return &chatMessageService{db: db}
}

func (s *chatMessageService) List(ctx context.Context, cond chat.ChatMessageListCond) ([]*chat.ChatMessage, error) {
	stmt := table.ChatMessages.SELECT(table.ChatMessages.AllColumns)
	pred := []jet.BoolExpression{}
	orderBy := []jet.OrderByClause{}

	if cond.ChatMessageID != nil {
		pred = append(pred, table.ChatMessages.ChatMessageID.EQ(jet.String(string(*cond.ChatMessageID))))
	}

	if cond.ChatThreadID != nil {
		pred = append(pred, table.ChatMessages.ChatThreadID.EQ(jet.String(string(*cond.ChatThreadID))))
	}

	if cond.SenderUserID != nil {
		pred = append(pred, table.ChatMessages.SenderUserID.EQ(jet.String(string(*cond.SenderUserID))))
	}

	if cond.MessageType != nil {
		pred = append(pred, table.ChatMessages.MessageType.EQ(jet.String(string(*cond.MessageType))))
	}

	if cond.Body != nil {
		pred = append(pred, table.ChatMessages.Body.EQ(jet.String(*cond.Body)))
	}

	switch cond.SortKey {
	case chat.ChatMessageSortKey_CreatedAt_ASC:
		orderBy = append(orderBy, table.ChatMessages.CreatedAt.ASC())
	case chat.ChatMessageSortKey_CreatedAt_DESC:
		orderBy = append(orderBy, table.ChatMessages.CreatedAt.DESC())
	default:
		orderBy = append(orderBy, table.ChatMessages.CreatedAt.DESC())
	}

	if len(pred) > 0 {
		stmt = stmt.WHERE(jet.AND(pred...))
	}

	stmt = stmt.ORDER_BY(orderBy...)

	if cond.Offset != nil {
		stmt = stmt.OFFSET(*cond.Offset)
	}

	if cond.Size > 0 {
		stmt = stmt.LIMIT(cond.Size)
	}

	dest := &ChatMessageModels{}
	if err := stmt.Query(s.db, dest); err != nil {
		return nil, err
	}

	if len(*dest) == 0 {
		return nil, nil
	}

	return dest.Unmarshal(), nil
}

func (s *chatMessageService) Count(ctx context.Context, cond chat.ChatMessageCountCond) (*uint64, error) {
	stmt := table.ChatMessages.SELECT(jet.COUNT(table.ChatMessages.ChatMessageID).AS("count"))
	pred := []jet.BoolExpression{}

	if cond.ChatMessageID != nil {
		pred = append(pred, table.ChatMessages.ChatMessageID.EQ(jet.String(string(*cond.ChatMessageID))))
	}

	if cond.ChatThreadID != nil {
		pred = append(pred, table.ChatMessages.ChatThreadID.EQ(jet.String(string(*cond.ChatThreadID))))
	}

	if cond.SenderUserID != nil {
		pred = append(pred, table.ChatMessages.SenderUserID.EQ(jet.String(string(*cond.SenderUserID))))
	}

	if cond.MessageType != nil {
		pred = append(pred, table.ChatMessages.MessageType.EQ(jet.String(string(*cond.MessageType))))
	}

	if cond.Body != nil {
		pred = append(pred, table.ChatMessages.Body.EQ(jet.String(*cond.Body)))
	}

	if len(pred) > 0 {
		stmt = stmt.WHERE(jet.AND(pred...))
	}

	var dest []struct {
		Count uint64
	}

	if err := stmt.QueryContext(ctx, s.db, &dest); err != nil {
		return nil, err
	}

	return &dest[0].Count, nil
}

type chatThreadParticipantRepository struct {
	db *sql.DB
}

func NewChatThreadParticipantRepository(db *sql.DB) chatsvc.ChatThreadParticipantRepository {
	return &chatThreadParticipantRepository{db: db}
}

func (r *chatThreadParticipantRepository) Find(ctx context.Context, chatThreadID chat.ChatThreadID, userID chat.UserID) (*chat.ChatThreadParticipant, error) {
	stmt := table.ChatThreadParticipants.SELECT(table.ChatThreadParticipants.AllColumns).WHERE(
		jet.AND(
			table.ChatThreadParticipants.ChatThreadID.EQ(jet.String(string(chatThreadID))),
			table.ChatThreadParticipants.UserID.EQ(jet.String(string(userID))),
		))

	dest := &ChatThreadParticipantModels{}
	if err := stmt.Query(r.db, dest); err != nil {
		return nil, err
	}

	if len(*dest) == 0 {
		return nil, nil
	}

	return dest.Unmarshal()[0], nil
}

func (r *chatThreadParticipantRepository) Store(ctx context.Context, entity *chat.ChatThreadParticipant) error {
	var lastReadMessageID *string
	if entity.LastReadMessageID != nil {
		id := string(*entity.LastReadMessageID)
		lastReadMessageID = &id
	}

	m := model.ChatThreadParticipants{
		ChatThreadID:      string(entity.ChatThreadID),
		UserID:            string(entity.UserID),
		ArchivedAt:        entity.ArchivedAt,
		MutedAt:           entity.MutedAt,
		LastReadMessageID: lastReadMessageID,
		LastReadAt:        entity.LastReadAt,
		JoinedAt:          entity.JoinedAt,
		LeftAt:            entity.LeftAt,
	}

	insertStmt := table.ChatThreadParticipants.INSERT(table.ChatThreadParticipants.AllColumns).MODEL(m)
	updateStmt := table.ChatThreadParticipants.UPDATE(table.ChatThreadParticipants.AllColumns).MODEL(m)
	updateStmt = updateStmt.WHERE(jet.AND(
		table.ChatThreadParticipants.ChatThreadID.EQ(jet.String(string(entity.ChatThreadID))),
		table.ChatThreadParticipants.UserID.EQ(jet.String(string(entity.UserID))),
	))

	_, err := insertStmt.Exec(r.db)
	if err != nil {
		if mysqlerr, ok := err.(*mysql.MySQLError); ok {
			switch mysqlerr.Number {
			case 1062:
				result, err := updateStmt.Exec(r.db)
				if err != nil {
					return err
				}
				rowsAffected, err := result.RowsAffected()
				if err != nil {
					return err
				}
				if rowsAffected == 0 {
					return fmt.Errorf("entity version conflicted")
				}
			default:
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func (r *chatThreadParticipantRepository) Remove(ctx context.Context, conds ...chatsvc.RemoveParticipantCond) error {
	if len(conds) == 0 {
		return nil
	}

	pred := make([]jet.BoolExpression, 0, len(conds))
	for _, c := range conds {
		pred = append(pred, jet.AND(
			table.ChatThreadParticipants.ChatThreadID.EQ(jet.String(string(c.ChatThreadID))),
			table.ChatThreadParticipants.UserID.EQ(jet.String(string(c.UserID))),
		))
	}

	stmt := table.ChatThreadParticipants.DELETE().WHERE(jet.OR(pred...))
	_, err := stmt.Exec(r.db)
	return err
}

type chatThreadParticipantService struct {
	db *sql.DB
}

func NewChatThreadParticipantService(db *sql.DB) chatsvc.ChatThreadParticipantService {
	return &chatThreadParticipantService{db: db}
}

func (s *chatThreadParticipantService) List(ctx context.Context, cond chat.ChatThreadParticipantListCond) ([]*chat.ChatThreadParticipant, error) {
	stmt := table.ChatThreadParticipants.SELECT(table.ChatThreadParticipants.AllColumns)
	pred := []jet.BoolExpression{}
	orderBy := []jet.OrderByClause{}

	if cond.ChatThreadID != nil {
		pred = append(pred, table.ChatThreadParticipants.ChatThreadID.EQ(jet.String(string(*cond.ChatThreadID))))
	}

	if cond.UserID != nil {
		pred = append(pred, table.ChatThreadParticipants.UserID.EQ(jet.String(string(*cond.UserID))))
	}

	if cond.JoinedAt != nil {
		pred = append(pred, table.ChatThreadParticipants.JoinedAt.EQ(jet.TimestampT(*cond.JoinedAt)))
	}

	if cond.LeftAt != nil {
		pred = append(pred, table.ChatThreadParticipants.LeftAt.EQ(jet.TimestampT(*cond.LeftAt)))
	}

	switch cond.SortKey {
	case chat.ChatThreadParticipantSortKey_JoinedAt_ASC:
		orderBy = append(orderBy, table.ChatThreadParticipants.JoinedAt.ASC())
	case chat.ChatThreadParticipantSortKey_JoinedAt_DESC:
		orderBy = append(orderBy, table.ChatThreadParticipants.JoinedAt.DESC())
	default:
		orderBy = append(orderBy, table.ChatThreadParticipants.JoinedAt.DESC())
	}

	if len(pred) > 0 {
		stmt = stmt.WHERE(jet.AND(pred...))
	}

	stmt = stmt.ORDER_BY(orderBy...)

	if cond.Offset != nil {
		stmt = stmt.OFFSET(*cond.Offset)
	}

	if cond.Size > 0 {
		stmt = stmt.LIMIT(cond.Size)
	}

	dest := &ChatThreadParticipantModels{}
	if err := stmt.Query(s.db, dest); err != nil {
		return nil, err
	}

	if len(*dest) == 0 {
		return nil, nil
	}

	return dest.Unmarshal(), nil
}

func (s *chatThreadParticipantService) Count(ctx context.Context, cond chat.ChatThreadParticipantCountCond) (*uint64, error) {
	stmt := table.ChatThreadParticipants.SELECT(
		jet.COUNT(table.ChatThreadParticipants.ChatThreadID).AS("count"),
	)
	pred := []jet.BoolExpression{}

	if cond.ChatThreadID != nil {
		pred = append(pred, table.ChatThreadParticipants.ChatThreadID.EQ(jet.String(string(*cond.ChatThreadID))))
	}

	if cond.UserID != nil {
		pred = append(pred, table.ChatThreadParticipants.UserID.EQ(jet.String(string(*cond.UserID))))
	}

	if cond.JoinedAt != nil {
		pred = append(pred, table.ChatThreadParticipants.JoinedAt.EQ(jet.TimestampT(*cond.JoinedAt)))
	}

	if cond.LeftAt != nil {
		pred = append(pred, table.ChatThreadParticipants.LeftAt.EQ(jet.TimestampT(*cond.LeftAt)))
	}

	if len(pred) > 0 {
		stmt = stmt.WHERE(jet.AND(pred...))
	}

	var dest []struct {
		Count uint64
	}

	if err := stmt.QueryContext(ctx, s.db, &dest); err != nil {
		return nil, err
	}

	return &dest[0].Count, nil
}

type chatMessageAttachmentRepository struct {
	db *sql.DB
}

func NewChatMessageAttachmentRepository(db *sql.DB) chatsvc.ChatMessageAttachmentRepository {
	return &chatMessageAttachmentRepository{db: db}
}

func (r *chatMessageAttachmentRepository) Find(ctx context.Context, id chat.ChatMessageAttachmentID) (*chat.ChatMessageAttachment, error) {
	stmt := table.ChatMessageAttachments.SELECT(table.ChatMessageAttachments.AllColumns).WHERE(
		table.ChatMessageAttachments.ChatMessageAttachmentID.EQ(jet.String(string(id))))

	dest := &ChatMessageAttachmentModels{}
	if err := stmt.Query(r.db, dest); err != nil {
		return nil, err
	}

	if len(*dest) == 0 {
		return nil, nil
	}

	return dest.Unmarshal()[0], nil
}

func (r *chatMessageAttachmentRepository) Store(ctx context.Context, entity *chat.ChatMessageAttachment) error {
	m := model.ChatMessageAttachments{
		ChatMessageAttachmentID: string(entity.ChatMessageAttachmentID),
		ChatMessageID:           string(entity.ChatMessageID),
		FileName:                entity.FileName,
		FileType:                entity.FileType,
		FileSize:                entity.FileSize,
		R2Bucket:                entity.R2Bucket,
		R2ObjectKey:             entity.R2ObjectKey,
		PublicURL:               entity.PublicURL,
		CreatedAt:               entity.CreatedAt,
	}

	insertStmt := table.ChatMessageAttachments.INSERT(table.ChatMessageAttachments.AllColumns).MODEL(m)
	updateStmt := table.ChatMessageAttachments.UPDATE(table.ChatMessageAttachments.AllColumns).MODEL(m)
	updateStmt = updateStmt.WHERE(
		table.ChatMessageAttachments.ChatMessageAttachmentID.EQ(jet.String(string(entity.ChatMessageAttachmentID))),
	)

	_, err := insertStmt.Exec(r.db)
	if err != nil {
		if mysqlerr, ok := err.(*mysql.MySQLError); ok {
			switch mysqlerr.Number {
			case 1062:
				result, err := updateStmt.Exec(r.db)
				if err != nil {
					return err
				}
				rowsAffected, err := result.RowsAffected()
				if err != nil {
					return err
				}
				if rowsAffected == 0 {
					return fmt.Errorf("entity version conflicted")
				}
			default:
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func (r *chatMessageAttachmentRepository) Remove(ctx context.Context, ids ...chat.ChatMessageAttachmentID) error {
	idExpressions := make([]jet.Expression, 0, len(ids))
	for _, id := range ids {
		idExpressions = append(idExpressions, jet.String(string(id)))
	}

	stmt := table.ChatMessageAttachments.DELETE().WHERE(
		table.ChatMessageAttachments.ChatMessageAttachmentID.IN(idExpressions...),
	)
	_, err := stmt.Exec(r.db)
	return err
}

type chatMessageAttachmentService struct {
	db *sql.DB
}

func NewChatMessageAttachmentService(db *sql.DB) chatsvc.ChatMessageAttachmentService {
	return &chatMessageAttachmentService{db: db}
}

func (s *chatMessageAttachmentService) List(ctx context.Context, cond chat.ChatMessageAttachmentListCond) ([]*chat.ChatMessageAttachment, error) {
	stmt := table.ChatMessageAttachments.SELECT(table.ChatMessageAttachments.AllColumns)
	pred := []jet.BoolExpression{}
	orderBy := []jet.OrderByClause{}

	if cond.ChatMessageAttachmentID != nil {
		pred = append(pred, table.ChatMessageAttachments.ChatMessageAttachmentID.EQ(jet.String(string(*cond.ChatMessageAttachmentID))))
	}

	if cond.ChatMessageID != nil {
		pred = append(pred, table.ChatMessageAttachments.ChatMessageID.EQ(jet.String(string(*cond.ChatMessageID))))
	}

	switch cond.SortKey {
	case chat.ChatMessageAttachmentSortKey_CreatedAt_ASC:
		orderBy = append(orderBy, table.ChatMessageAttachments.CreatedAt.ASC())
	case chat.ChatMessageAttachmentSortKey_CreatedAt_DESC:
		orderBy = append(orderBy, table.ChatMessageAttachments.CreatedAt.DESC())
	default:
		orderBy = append(orderBy, table.ChatMessageAttachments.CreatedAt.DESC())
	}

	if len(pred) > 0 {
		stmt = stmt.WHERE(jet.AND(pred...))
	}

	stmt = stmt.ORDER_BY(orderBy...)

	if cond.Offset != nil {
		stmt = stmt.OFFSET(*cond.Offset)
	}

	if cond.Size > 0 {
		stmt = stmt.LIMIT(cond.Size)
	}

	dest := &ChatMessageAttachmentModels{}
	if err := stmt.Query(s.db, dest); err != nil {
		return nil, err
	}

	if len(*dest) == 0 {
		return nil, nil
	}

	return dest.Unmarshal(), nil
}

func (s *chatMessageAttachmentService) Count(ctx context.Context, cond chat.ChatMessageAttachmentCountCond) (*uint64, error) {
	stmt := table.ChatMessageAttachments.SELECT(
		jet.COUNT(table.ChatMessageAttachments.ChatMessageAttachmentID).AS("count"),
	)
	pred := []jet.BoolExpression{}

	if cond.ChatMessageAttachmentID != nil {
		pred = append(pred, table.ChatMessageAttachments.ChatMessageAttachmentID.EQ(jet.String(string(*cond.ChatMessageAttachmentID))))
	}

	if cond.ChatMessageID != nil {
		pred = append(pred, table.ChatMessageAttachments.ChatMessageID.EQ(jet.String(string(*cond.ChatMessageID))))
	}

	if len(pred) > 0 {
		stmt = stmt.WHERE(jet.AND(pred...))
	}

	var dest []struct {
		Count uint64
	}

	if err := stmt.QueryContext(ctx, s.db, &dest); err != nil {
		return nil, err
	}

	return &dest[0].Count, nil
}

type ChatThreadModels []struct {
	model.ChatThreads
}

func (src ChatThreadModels) Unmarshal() []*chat.ChatThread {
	out := make([]*chat.ChatThread, 0, len(src))
	for _, item := range src {
		out = append(out, &chat.ChatThread{
			ChatThreadID:  chat.ChatThreadID(item.ChatThreadID),
			Type:          chat.ChatThreadsType(item.Type),
			Title:         item.Title,
			LastMessageID: item.LastMessageID,
			LastMessageAt: item.LastMessageAt,
			CreatedAt:     item.CreatedAt,
			UpdatedAt:     item.UpdatedAt,
		})
	}
	return out
}

type ChatMessageModels []struct {
	model.ChatMessages
}

func (src ChatMessageModels) Unmarshal() []*chat.ChatMessage {
	out := make([]*chat.ChatMessage, 0, len(src))
	for _, item := range src {
		out = append(out, &chat.ChatMessage{
			ChatMessageID: chat.ChatMessageID(item.ChatMessageID),
			ChatThreadID:  chat.ChatThreadID(item.ChatThreadID),
			SenderUserID:  chat.UserID(item.SenderUserID),
			MessageType:   chat.ChatMessageMessageType(item.MessageType),
			Body:          item.Body,
			CreatedAt:     item.CreatedAt,
			UpdatedAt:     item.UpdatedAt,
			DeletedAt:     item.DeletedAt,
		})
	}
	return out
}

type ChatThreadParticipantModels []struct {
	model.ChatThreadParticipants
}

func (src ChatThreadParticipantModels) Unmarshal() []*chat.ChatThreadParticipant {
	out := make([]*chat.ChatThreadParticipant, 0, len(src))
	for _, item := range src {
		var lastReadMessageID *chat.ChatMessageID
		if item.LastReadMessageID != nil {
			id := chat.ChatMessageID(*item.LastReadMessageID)
			lastReadMessageID = &id
		}

		out = append(out, &chat.ChatThreadParticipant{
			ChatThreadID:      chat.ChatThreadID(item.ChatThreadID),
			UserID:            chat.UserID(item.UserID),
			ArchivedAt:        item.ArchivedAt,
			MutedAt:           item.MutedAt,
			LastReadMessageID: lastReadMessageID,
			LastReadAt:        item.LastReadAt,
			JoinedAt:          item.JoinedAt,
			LeftAt:            item.LeftAt,
		})
	}
	return out
}

type ChatMessageAttachmentModels []struct {
	model.ChatMessageAttachments
}

func (src ChatMessageAttachmentModels) Unmarshal() []*chat.ChatMessageAttachment {
	out := make([]*chat.ChatMessageAttachment, 0, len(src))
	for _, item := range src {
		out = append(out, &chat.ChatMessageAttachment{
			ChatMessageAttachmentID: chat.ChatMessageAttachmentID(item.ChatMessageAttachmentID),
			ChatMessageID:           chat.ChatMessageID(item.ChatMessageID),
			FileName:                item.FileName,
			FileType:                item.FileType,
			FileSize:                item.FileSize,
			R2Bucket:                item.R2Bucket,
			R2ObjectKey:             item.R2ObjectKey,
			PublicURL:               item.PublicURL,
			CreatedAt:               item.CreatedAt,
		})
	}
	return out
}
