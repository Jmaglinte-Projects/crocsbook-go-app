package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/media"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql/lib/db_crocs/model"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql/lib/db_crocs/table"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/mediasvc"
	jet "github.com/go-jet/jet/v2/mysql"
	"github.com/go-sql-driver/mysql"
)

type mediaRepository struct {
	db          *sql.DB
	mediaR2Repo mediasvc.MediaR2Repository
}

func NewMediaRepository(db *sql.DB, mediaR2Repo mediasvc.MediaR2Repository) mediasvc.MediaRepository {
	return &mediaRepository{
		db:          db,
		mediaR2Repo: mediaR2Repo,
	}
}

func (r *mediaRepository) Find(ctx context.Context, id media.MediaID) (*mediasvc.ViewMedia, error) {
	stmt := table.Medias.SELECT(table.Medias.AllColumns).WHERE(
		table.Medias.MediaID.EQ(jet.String(string(id))))

	dest := &MediaModels{}
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

	out := dest.ViewMedia()
	item := out[0]

	// get the URL from the cloudflare R2 repository (skip if no R2 repo or no key)
	if item.ObjectKey != "" && r.mediaR2Repo != nil {
		if url, err := r.mediaR2Repo.Find(ctx, item.ObjectKey); err == nil {
			item.PresignedURL = url
		}
	}

	return item, nil
}

func (r *mediaRepository) Store(ctx context.Context, entity *media.Media) error {
	if r.mediaR2Repo != nil {
		if err := r.mediaR2Repo.Store(ctx, entity); err != nil {
			return err
		}
	}

	m := model.Medias{}
	m.MediaID = string(entity.MediaID)
	m.MediaPostID = string(entity.MediaPostID)
	m.ObjectKey = &entity.ObjectKey

	t := string(entity.Type)
	m.Type = &t
	m.CreatedTime = entity.CreatedTime

	insertStmt := table.Medias.INSERT(table.Medias.AllColumns).MODEL(m)

	updateStmt := table.Medias.UPDATE(table.Medias.AllColumns).MODEL(m)
	updateStmt = updateStmt.WHERE(table.Medias.MediaID.EQ(jet.String(string(entity.MediaID))))

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

func (r *mediaRepository) Remove(ctx context.Context, ids ...media.MediaID) error {
	idExpressions := make([]jet.Expression, 0, len(ids))
	for _, id := range ids {
		idExpressions = append(idExpressions, jet.String(string(id)))
	}

	stmt := table.Medias.DELETE().WHERE(table.Medias.MediaID.IN(idExpressions...))
	_, err := stmt.Exec(r.db)
	if err != nil {
		return err
	}

	return nil
}

type mediaService struct {
	db          *sql.DB
	mediaR2Repo mediasvc.MediaR2Repository
}

func NewMediaService(db *sql.DB, mediaR2Repo mediasvc.MediaR2Repository) mediasvc.MediaService {
	return &mediaService{
		db:          db,
		mediaR2Repo: mediaR2Repo,
	}
}

func (s *mediaService) List(ctx context.Context, cond media.ListCond, option mediasvc.ListOption) ([]*mediasvc.ViewMedia, error) {
	stmt := table.Medias.SELECT(table.Medias.AllColumns)
	pred := []jet.BoolExpression{}
	orderBy := []jet.OrderByClause{}

	if cond.MediaID != nil {
		pred = append(pred, table.Medias.MediaID.EQ(jet.String(string(*cond.MediaID))))
	}

	if len(cond.MediaIDs) > 0 {
		idExpressions := make([]jet.Expression, 0, len(cond.MediaIDs))

		for _, id := range cond.MediaIDs {
			idExpressions = append(idExpressions, jet.String(string(id)))
		}

		pred = append(pred, table.Medias.MediaID.IN(
			idExpressions...,
		))
	}

	if cond.MediaPostID != nil {
		pred = append(pred, table.Medias.MediaPostID.EQ(jet.String(string(*cond.MediaPostID))))
	}

	switch option.SortKey {
	case mediasvc.ListOptionSortKey_CreatedAt_ASC:
		orderBy = append(orderBy, table.Medias.CreatedTime.ASC())
	case mediasvc.ListOptionSortKey_CreatedAt_DESC:
		orderBy = append(orderBy, table.Medias.CreatedTime.DESC())
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

	dest := &MediaModels{}
	err := stmt.Query(s.db, dest)
	if err != nil {
		return nil, err
	}

	if len(*dest) == 0 {
		return nil, nil
	}

	out := dest.ViewMedia()

	if s.mediaR2Repo != nil {
		for _, item := range out {
			if item.ObjectKey != "" {
				// if url, err := s.mediaR2Repo.Find(ctx, item.ObjectKey);
				url := fmt.Sprintf("%s/%s", os.Getenv("R2_PUBLIC_BASE_URL"), item.ObjectKey)
				item.PresignedURL = url
			}
		}
	}

	return out, nil
}

func (s *mediaService) Count(ctx context.Context, cond media.CountCond, option mediasvc.CountOption) (*uint64, error) {
	stmt := table.Medias.SELECT(jet.COUNT(table.Medias.MediaID).AS("count"))
	pred := []jet.BoolExpression{}

	if cond.MediaID != nil {
		pred = append(pred, table.Medias.MediaID.EQ(jet.String(string(*cond.MediaID))))
	}

	if len(cond.MediaIDs) > 0 {
		idExpressions := make([]jet.Expression, 0, len(cond.MediaIDs))
		for _, id := range cond.MediaIDs {
			idExpressions = append(idExpressions, jet.String(string(id)))
		}

		pred = append(pred, table.Medias.MediaID.IN(
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

type MediaModels []struct {
	model.Medias
}

func (src MediaModels) ViewMedia() []*mediasvc.ViewMedia {
	out := make([]*mediasvc.ViewMedia, 0, len(src))
	for _, item := range src {
		mediaEntity := &media.Media{}
		mediaEntity.MediaID = media.MediaID(item.MediaID)
		mediaEntity.MediaPostID = media.PostID(item.MediaPostID)

		objectKey := item.ObjectKey
		if objectKey != nil {
			mediaEntity.ObjectKey = *objectKey
		}

		t := media.Type(*item.Type)
		mediaEntity.Type = t
		mediaEntity.CreatedTime = item.CreatedTime

		vw := &mediasvc.ViewMedia{
			Media: *mediaEntity,
		}
		out = append(out, vw)
	}
	return out
}
