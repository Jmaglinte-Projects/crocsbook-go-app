package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/post"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql/lib/db_crocs/model"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql/lib/db_crocs/table"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/postsvc"
	jet "github.com/go-jet/jet/v2/mysql"
	"github.com/go-sql-driver/mysql"
)

type postRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) postsvc.PostRepository {
	return &postRepository{
		db: db,
	}
}

func (r *postRepository) Find(ctx context.Context, id post.PostID) (*postsvc.ViewPost, error) {
	stmt := table.Posts.SELECT(table.Posts.AllColumns).WHERE(
		table.Posts.PostID.EQ(jet.String(string(id))))

	dest := &PostModels{}
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

	out := dest.ViewPost()

	return out[0], nil
}

func (r *postRepository) Store(ctx context.Context, entity *post.Post) error {

	m := model.Posts{}
	m.PostID = string(entity.PostID)
	m.PostProjectID = string(entity.PostProjectID)
	m.Content = entity.Content

	visibility := model.PostsVisibility(*entity.Visibility)
	m.Visibility = &visibility
	m.CreatedTime = entity.CreatedTime
	m.UpdatedTime = entity.UpdatedTime

	insertStmt := table.Posts.INSERT(table.Posts.AllColumns).MODEL(m)

	updateStmt := table.Posts.UPDATE(table.Posts.AllColumns).MODEL(m)
	updateStmt = updateStmt.WHERE(table.Posts.PostID.EQ(jet.String(string(entity.PostID))))

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

func (r *postRepository) Remove(ctx context.Context, ids ...post.PostID) error {
	idExpressions := make([]jet.Expression, 0, len(ids))
	for _, id := range ids {
		idExpressions = append(idExpressions, jet.String(string(id)))
	}

	stmt := table.Posts.DELETE().WHERE(table.Posts.PostID.IN(idExpressions...))
	_, err := stmt.Exec(r.db)
	if err != nil {
		return err
	}

	return nil
}

type postService struct {
	db *sql.DB
}

func NewPostService(db *sql.DB) postsvc.PostService {
	return &postService{
		db: db,
	}
}

func (s *postService) List(ctx context.Context, cond post.ListCond, option postsvc.ListOption) ([]*postsvc.ViewPost, error) {
	stmt := table.Posts.SELECT(table.Posts.AllColumns)
	pred := []jet.BoolExpression{}
	orderBy := []jet.OrderByClause{}

	if cond.PostID != nil {
		pred = append(pred, table.Posts.PostID.EQ(jet.String(string(*cond.PostID))))
	}

	if len(cond.PostIDs) > 0 {
		idExpressions := make([]jet.Expression, 0, len(cond.PostIDs))

		for _, id := range cond.PostIDs {
			idExpressions = append(idExpressions, jet.String(string(id)))
		}

		pred = append(pred, table.Posts.PostID.IN(
			idExpressions...,
		))
	}

	switch option.SortKey {
	case postsvc.ListOptionSortKey_CreatedAt_ASC:
		orderBy = append(orderBy, table.Posts.CreatedTime.ASC())
	case postsvc.ListOptionSortKey_CreatedAt_DESC:
		orderBy = append(orderBy, table.Posts.CreatedTime.DESC())
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

	dest := &PostModels{}
	err := stmt.Query(s.db, dest)
	if err != nil {
		return nil, err
	}

	if len(*dest) == 0 {
		return nil, nil
	}

	out := dest.ViewPost()

	return out, nil
}

func (s *postService) Count(ctx context.Context, cond post.CountCond, option postsvc.CountOption) (*uint64, error) {
	stmt := table.Posts.SELECT(jet.COUNT(table.Posts.PostID).AS("count"))
	pred := []jet.BoolExpression{}

	if cond.PostID != nil {
		pred = append(pred, table.Posts.PostID.EQ(jet.String(string(*cond.PostID))))
	}

	if len(cond.PostIDs) > 0 {
		idExpressions := make([]jet.Expression, 0, len(cond.PostIDs))
		for _, id := range cond.PostIDs {
			idExpressions = append(idExpressions, jet.String(string(id)))
		}

		pred = append(pred, table.Posts.PostID.IN(
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

type PostModels []struct {
	model.Posts
}

func (src PostModels) ViewPost() []*postsvc.ViewPost {
	out := make([]*postsvc.ViewPost, 0, len(src))
	for _, item := range src {
		postEntity := &post.Post{}
		postEntity.PostID = post.PostID(item.PostID)
		postEntity.PostProjectID = post.ProjectID(item.PostProjectID)
		postEntity.Content = item.Content
		postEntity.Visibility = (*post.Visibility)(item.Visibility)
		postEntity.CreatedTime = item.CreatedTime
		postEntity.UpdatedTime = item.UpdatedTime

		vw := &postsvc.ViewPost{
			Post: *postEntity,
		}
		out = append(out, vw)
	}
	return out
}
