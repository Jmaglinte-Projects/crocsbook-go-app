package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/reaction"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql/lib/db_crocs/model"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql/lib/db_crocs/table"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/reactionsvc"
	jet "github.com/go-jet/jet/v2/mysql"
	"github.com/go-sql-driver/mysql"
)

type ReactionRepository struct {
	db *sql.DB
}

func NewReactionRepository(db *sql.DB) *ReactionRepository {
	return &ReactionRepository{
		db: db,
	}
}

func (r *ReactionRepository) Find(ctx context.Context, id reaction.ReactionID) (*reactionsvc.ViewReaction, error) {
	stmt := table.Reactions.SELECT(table.Reactions.AllColumns).WHERE(
		table.Reactions.ReactionID.EQ(jet.String(string(id))))

	dest := &ReactionModels{}
	err := stmt.Query(r.db, dest)
	if err != nil {
		return nil, err
	}

	debugSql := stmt.DebugSql()
	fmt.Println("--------------------------------")
	fmt.Println(debugSql)
	fmt.Println("--------------------------------")

	if len(*dest) == 0 {
		return nil, nil
	}

	out := dest.ViewReaction()

	return out[0], nil
}

func (r *ReactionRepository) Store(ctx context.Context, entity *reaction.Reaction) error {

	m := model.Reactions{}
	m.ReactionID = string(entity.ReactionID)
	m.ReactionProjectID = string(entity.ReactionProjectID)

	t := model.ReactionsType(*entity.Type)
	m.Type = &t
	m.CreatedTime = entity.CreatedTime

	insertStmt := table.Reactions.INSERT(table.Reactions.AllColumns).MODEL(m)

	updateStmt := table.Reactions.UPDATE(table.Reactions.AllColumns).MODEL(m)
	updateStmt = updateStmt.WHERE(table.Reactions.ReactionID.EQ(jet.String(string(entity.ReactionID))))

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

func (r *ReactionRepository) Remove(ctx context.Context, ids ...reaction.ReactionID) error {
	idExpressions := make([]jet.Expression, 0, len(ids))
	for _, id := range ids {
		idExpressions = append(idExpressions, jet.String(string(id)))
	}

	stmt := table.Reactions.DELETE().WHERE(table.Reactions.ReactionID.IN(idExpressions...))
	_, err := stmt.Exec(r.db)
	if err != nil {
		return err
	}

	return nil
}

type ReactionService struct {
	db *sql.DB
}

func NewReactionService(db *sql.DB) *ReactionService {
	return &ReactionService{
		db: db,
	}
}

func (s *ReactionService) List(ctx context.Context, cond reaction.ListCond, option reactionsvc.ListOption) ([]*reactionsvc.ViewReaction, error) {
	stmt := table.Reactions.SELECT(table.Reactions.AllColumns)
	pred := []jet.BoolExpression{}
	orderBy := []jet.OrderByClause{}

	if cond.ReactionID != nil {
		pred = append(pred, table.Reactions.ReactionID.EQ(jet.String(string(*cond.ReactionID))))
	}

	if len(cond.ReactionIDs) > 0 {
		idExpressions := make([]jet.Expression, 0, len(cond.ReactionIDs))

		for _, id := range cond.ReactionIDs {
			idExpressions = append(idExpressions, jet.String(string(id)))
		}

		pred = append(pred, table.Reactions.ReactionID.IN(
			idExpressions...,
		))
	}

	switch option.SortKey {
	case reactionsvc.ListOptionSortKey_CreatedAt_ASC:
		orderBy = append(orderBy, table.Reactions.CreatedTime.ASC())
	case reactionsvc.ListOptionSortKey_CreatedAt_DESC:
		orderBy = append(orderBy, table.Reactions.CreatedTime.DESC())
	}

	if len(pred) > 0 {
		stmt = stmt.WHERE(jet.AND(pred...))
	}

	stmt = stmt.ORDER_BY(orderBy...)

	if option.Offset != nil {
		stmt = stmt.OFFSET(*option.Offset)
	}

	if option.Size > 0 {
		stmt = stmt.LIMIT(option.Size)
	}

	debugSql := stmt.DebugSql()
	fmt.Println("--------------------------------")
	fmt.Println(debugSql)
	fmt.Println("--------------------------------")

	dest := &ReactionModels{}
	err := stmt.Query(s.db, dest)
	if err != nil {
		return nil, err
	}

	if len(*dest) == 0 {
		return nil, nil
	}

	out := dest.ViewReaction()

	return out, nil
}

func (s *ReactionService) Count(ctx context.Context, cond reaction.CountCond, option reactionsvc.CountOption) (*uint64, error) {
	stmt := table.Reactions.SELECT(jet.COUNT(table.Reactions.ReactionID).AS("count"))
	pred := []jet.BoolExpression{}

	if cond.ReactionID != nil {
		pred = append(pred, table.Reactions.ReactionID.EQ(jet.String(string(*cond.ReactionID))))
	}

	if len(cond.ReactionIDs) > 0 {
		idExpressions := make([]jet.Expression, 0, len(cond.ReactionIDs))
		for _, id := range cond.ReactionIDs {
			idExpressions = append(idExpressions, jet.String(string(id)))
		}

		pred = append(pred, table.Reactions.ReactionID.IN(
			idExpressions...,
		))
	}

	if len(pred) > 0 {
		stmt = stmt.WHERE(jet.AND(pred...))
	}

	debugSql := stmt.DebugSql()
	fmt.Println("--------------------------------")
	fmt.Println(debugSql)
	fmt.Println("--------------------------------")

	var dest []struct {
		// TIP if there are weird error this was changed from uint32 to uint64
		Count uint64
	}

	err := stmt.QueryContext(ctx, s.db, &dest)
	if err != nil {
		return nil, err
	}

	return &dest[0].Count, nil
}

type ReactionModels []struct {
	model.Reactions
}

func (src ReactionModels) ViewReaction() []*reactionsvc.ViewReaction {
	out := make([]*reactionsvc.ViewReaction, 0, len(src))
	for _, item := range src {
		reactionEntity := &reaction.Reaction{}
		reactionEntity.ReactionID = reaction.ReactionID(item.ReactionID)
		reactionEntity.ReactionProjectID = reaction.ProjectID(item.ReactionProjectID)

		t := reaction.Type(*item.Type)
		reactionEntity.Type = &t
		reactionEntity.CreatedTime = item.CreatedTime

		vw := &reactionsvc.ViewReaction{
			Reaction: *reactionEntity,
		}
		out = append(out, vw)
	}
	return out
}
