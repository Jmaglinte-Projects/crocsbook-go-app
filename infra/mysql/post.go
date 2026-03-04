package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

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
	postReactionSubQuery := jet.SELECT(table.PostReactions.PostID, jet.COUNT(table.PostReactions.PostID).AS("post_reaction_count")).FROM(table.PostReactions).GROUP_BY(table.PostReactions.PostID).AsTable("pr")

	pr := table.PostReactions
	hasReactedExists := jet.EXISTS(
		jet.SELECT(jet.Int(1)).
			FROM(pr).
			WHERE(
				pr.PostID.EQ(table.Posts.PostID).
					AND(pr.UserID.EQ(table.Projects.ProjectUserID)),
			),
	).AS("has_reacted")

	stmt := table.Posts.
		LEFT_JOIN(postReactionSubQuery, table.PostReactions.PostID.From(postReactionSubQuery).EQ(table.Posts.PostID)).
		LEFT_JOIN(table.Projects, table.Projects.ProjectID.EQ(table.Posts.PostProjectID)).
		SELECT(table.Posts.AllColumns, jet.IntegerColumn("post_reaction_count").From(postReactionSubQuery), hasReactedExists).WHERE(
		table.Posts.PostID.EQ(jet.String(string(id))))

	debugSql := stmt.DebugSql()
	fmt.Println("--------------------------------")
	fmt.Println(debugSql)

	/*
		SELECT posts.post_id AS "posts.post_id",
			posts.post_project_id AS "posts.post_project_id",
			posts.content AS "posts.content",
			posts.visibility AS "posts.visibility",
			posts.created_time AS "posts.created_time",
			posts.updated_time AS "posts.updated_time",
			pr.post_reaction_count AS "post_reaction_count",
			(EXISTS (
					SELECT 1
					FROM db_crocs.post_reactions
					WHERE (post_reactions.post_id = posts.post_id) AND (post_reactions.user_id = projects.project_user_id)
			)) AS "has_reacted"
		FROM db_crocs.posts
				LEFT JOIN (
							SELECT post_reactions.post_id AS "post_reactions.post_id",
									COUNT(post_reactions.post_id) AS "post_reaction_count"
							FROM db_crocs.post_reactions
							GROUP BY post_reactions.post_id
				) AS pr ON (pr.`post_reactions.post_id` = posts.post_id)
				LEFT JOIN db_crocs.projects ON (projects.project_id = posts.post_project_id)
		WHERE posts.post_id = '15aac350-0f7f-11f1-826f-8abfd21201dc';
	*/
	fmt.Println("--------------------------------")

	dest := &PostModels{}
	err := stmt.Query(r.db, dest)
	if err != nil {
		return nil, err
	}

	// debugSql := stmt.DebugSql()
	// fmt.Println("--------------------------------")
	// fmt.Println(debugSql)

	/*
		SELECT posts.post_id AS "posts.post_id",
				posts.post_project_id AS "posts.post_project_id",
				posts.content AS "posts.content",
				posts.visibility AS "posts.visibility",
				posts.created_time AS "posts.created_time",
				posts.updated_time AS "posts.updated_time",
				pr.post_reaction_count AS "post_reaction_count"
		FROM db_crocs.posts
				LEFT JOIN (
							SELECT post_reactions.post_id AS "post_reactions.post_id",
									COUNT(post_reactions.post_id) AS "post_reaction_count"
							FROM db_crocs.post_reactions
							GROUP BY post_reactions.post_id
				) AS pr ON (pr.`post_reactions.post_id` = posts.post_id)
		WHERE posts.post_id = '15aac350-0f7f-11f1-826f-8abfd21201dc';
	*/
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

// type postReactionRepository struct {
// 	db *sql.DB
// }

// func NewPostReactionRepository(db *sql.DB) postsvc.PostReactionRepository {
// 	return &postReactionRepository{
// 		db: db,
// 	}
// }

// func (r *postReactionRepository) Find(ctx context.Context, id post.PostReactionID) (*post.PostReactions, error) {
// 	stmt := table.PostReactions.SELECT(table.PostReactions.AllColumns).WHERE(
// 		table.PostReactions.PostReactionID.EQ(jet.String(string(id))))

// 	dest := &PostReactionsModels{}
// 	err := stmt.Query(r.db, dest)
// 	if err != nil {
// 		return nil, err
// 	}

// 	debugSql := stmt.DebugSql()
// 	fmt.Println("--------------------------------")
// 	fmt.Println(debugSql)
// 	fmt.Println("--------------------------------")

// 	if len(*dest) == 0 {
// 		return nil, nil
// 	}

// 	out := dest.Unmarshal()

// 	return out[0], nil
// }

// func (r *postReactionRepository) Store(ctx context.Context, entity *post.PostReactions) error {
// 	return nil
// }

// func (r *postReactionRepository) Remove(ctx context.Context, ids ...post.PostReactionID) error {
// 	return nil
// }

type postService struct {
	db *sql.DB
}

func NewPostService(db *sql.DB) postsvc.PostService {
	return &postService{
		db: db,
	}
}

func (s *postService) List(ctx context.Context, cond post.ListCond) ([]*postsvc.ViewPost, error) {
	postReactionSubQuery := jet.SELECT(table.PostReactions.PostID, jet.COUNT(table.PostReactions.PostID).AS("post_reaction_count")).FROM(table.PostReactions).GROUP_BY(table.PostReactions.PostID).AsTable("pr")

	pr := table.PostReactions
	hasReactedExists := jet.EXISTS(
		jet.SELECT(jet.Int(1)).
			FROM(pr).
			WHERE(
				pr.PostID.EQ(table.Posts.PostID).
					AND(pr.UserID.EQ(table.Projects.ProjectUserID)),
			),
	).AS("has_reacted")

	stmt := table.Posts.
		LEFT_JOIN(postReactionSubQuery, table.PostReactions.PostID.From(postReactionSubQuery).EQ(table.Posts.PostID)).
		LEFT_JOIN(table.Projects, table.Projects.ProjectID.EQ(table.Posts.PostProjectID)).
		SELECT(table.Posts.AllColumns, jet.IntegerColumn("post_reaction_count").From(postReactionSubQuery), hasReactedExists)

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

	if cond.PostProjectID != nil {
		pred = append(pred, table.Posts.PostProjectID.EQ(jet.String(string(*cond.PostProjectID))))
	}

	if len(cond.PostProjectIDs) > 0 {
		idExpressions := make([]jet.Expression, 0, len(cond.PostProjectIDs))
		for _, id := range cond.PostProjectIDs {
			idExpressions = append(idExpressions, jet.String(string(id)))
		}

		pred = append(pred, table.Posts.PostProjectID.IN(idExpressions...))
	}

	if cond.CreatedTime != nil {
		/* Example query:
		AND (posts.created_time >= TIMESTAMP('2026-02-01 00:00:00'))
		AND (posts.created_time < TIMESTAMP('2026-03-01 00:00:00'))
		*/
		start := *cond.CreatedTime
		firstOfMonth := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
		firstOfNextMonth := firstOfMonth.AddDate(0, 1, 0)
		pred = append(pred,
			table.Posts.CreatedTime.GT_EQ(jet.Timestamp(firstOfMonth.Year(), firstOfMonth.Month(), firstOfMonth.Day(), 0, 0, 0, 0)),
			table.Posts.CreatedTime.LT(jet.Timestamp(firstOfNextMonth.Year(), firstOfNextMonth.Month(), firstOfNextMonth.Day(), 0, 0, 0, 0)),
		)
	}

	switch cond.SortKey {
	case post.PostSortKey_CreatedTime_ASC:
		orderBy = append(orderBy, table.Posts.CreatedTime.ASC())
	case post.PostSortKey_CreatedTime_DESC:
		orderBy = append(orderBy, table.Posts.CreatedTime.DESC())
	default:
		orderBy = append(orderBy, table.Posts.CreatedTime.DESC())
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

	debugSql := stmt.DebugSql()
	fmt.Println("--------------------------------")
	fmt.Println(debugSql)
	fmt.Println("--------------------------------")

	dest := &PostModels{}
	err := stmt.Query(s.db, dest)
	if err != nil {
		fmt.Println("error:", err)
		return nil, err
	}

	if len(*dest) == 0 {
		return nil, nil
	}

	out := dest.ViewPost()

	return out, nil
}

func (s *postService) Count(ctx context.Context, cond post.CountCond) (*uint64, error) {
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

	if cond.PostProjectID != nil {
		pred = append(pred, table.Posts.PostProjectID.EQ(jet.String(string(*cond.PostProjectID))))
	}

	if len(cond.PostProjectIDs) > 0 {
		idExpressions := make([]jet.Expression, 0, len(cond.PostProjectIDs))
		for _, id := range cond.PostProjectIDs {
			idExpressions = append(idExpressions, jet.String(string(id)))
		}

		pred = append(pred, table.Posts.PostProjectID.IN(idExpressions...))
	}

	if cond.CreatedTime != nil {
		start := *cond.CreatedTime
		end := start.AddDate(0, 1, 0) // first day of next month
		pred = append(pred,
			table.Posts.CreatedTime.GT_EQ(jet.Timestamp(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0)),
			table.Posts.CreatedTime.LT(jet.Timestamp(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0)),
		)
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

func (s *postService) ListPostStatsByProjectIds(ctx context.Context, cond post.ListPostStatsByProjectIdsCond, projectIds ...post.ProjectID) ([]*postsvc.ListPostStatsByProjectIds, error) {
	stmt := table.Posts.SELECT(jet.COUNT(table.Posts.PostID).AS("count"), table.Posts.PostProjectID.AS("post_project_id"), jet.MAX(table.Posts.CreatedTime).AS("last_post_time")).GROUP_BY(table.Posts.PostProjectID)
	pred := []jet.BoolExpression{}
	orderBy := []jet.OrderByClause{
		jet.MAX(table.Posts.CreatedTime).DESC(),
	}

	idExpressions := make([]jet.Expression, 0, len(projectIds))
	for _, id := range projectIds {
		idExpressions = append(idExpressions, jet.String(string(id)))
	}

	if len(projectIds) > 0 {
		pred = append(pred, table.Posts.PostProjectID.IN(idExpressions...))
	}

	if cond.CreatedTime != nil {
		start := *cond.CreatedTime
		firstOfMonth := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
		firstOfNextMonth := firstOfMonth.AddDate(0, 1, 0)
		pred = append(pred,
			table.Posts.CreatedTime.GT_EQ(jet.Timestamp(firstOfMonth.Year(), firstOfMonth.Month(), firstOfMonth.Day(), 0, 0, 0, 0)),
			table.Posts.CreatedTime.LT(jet.Timestamp(firstOfNextMonth.Year(), firstOfNextMonth.Month(), firstOfNextMonth.Day(), 0, 0, 0, 0)),
		)
	}

	// switch cond.SortKey {
	// case post.PostSortKey_CreatedTime_ASC:
	// 	orderBy = append(orderBy, table.Posts.CreatedTime.ASC())
	// case post.PostSortKey_CreatedTime_DESC:
	// 	orderBy = append(orderBy, table.Posts.CreatedTime.DESC())
	// default:
	// 	orderBy = append(orderBy, table.Posts.CreatedTime.DESC())
	// }

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

	debugSql := stmt.DebugSql()
	fmt.Println("--------------------------------")
	fmt.Println(debugSql)
	fmt.Println("--------------------------------")

	dest := &CountTotalPostsByProjectIdModel{}
	err := stmt.Query(s.db, dest)
	if err != nil {
		return nil, err
	}

	out := dest.Format()

	jsonText, _ := json.MarshalIndent(dest, "", "\t")
	fmt.Println(string(jsonText))

	return out, nil
}

type PostModels []struct {
	model.Posts

	PostReactionCount uint64
	HasReacted        bool
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
		postEntity.PostReactionCount = item.PostReactionCount
		postEntity.HasReacted = item.HasReacted

		vw := &postsvc.ViewPost{
			Post: *postEntity,
		}
		out = append(out, vw)
	}
	return out
}

type CountTotalPostsByProjectIdModel []struct {
	PostProjectID string
	Count         uint64
	LastPostTime  time.Time
}

func (src CountTotalPostsByProjectIdModel) Format() []*postsvc.ListPostStatsByProjectIds {
	out := make([]*postsvc.ListPostStatsByProjectIds, 0, len(src))
	for _, item := range src {
		out = append(out, &postsvc.ListPostStatsByProjectIds{
			ProjectID:    post.ProjectID(item.PostProjectID),
			Count:        item.Count,
			LastPostTime: item.LastPostTime,
		})
	}
	return out
}

// type PostReactionsModels []struct {
// 	model.PostReactions
// }

// func (src PostReactionsModels) Unmarshal() []*post.PostReactions {
// 	out := make([]*post.PostReactions, 0, len(src))
// 	for _, item := range src {
// 		postReactionEntity := &post.PostReactions{}
// 		postReactionEntity.PostReactionID = post.PostReactionID(item.PostReactionID)
// 		postReactionEntity.PostID = post.PostID(item.PostID)
// 		postReactionEntity.UserID = post.UserID(item.UserID)
// 		postReactionEntity.ReactionType = (*post.ReactionType)(item.ReactionType)
// 		postReactionEntity.CreatedTime = item.CreatedTime
// 		out = append(out, postReactionEntity)
// 	}
// 	return out
// }
