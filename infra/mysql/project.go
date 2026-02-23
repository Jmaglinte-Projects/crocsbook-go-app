package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/project"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql/lib/db_crocs/model"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql/lib/db_crocs/table"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/projectsvc"
	jet "github.com/go-jet/jet/v2/mysql"
	"github.com/go-sql-driver/mysql"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{
		db: db,
	}
}

func (r *ProjectRepository) Find(ctx context.Context, id project.ProjectID) (*projectsvc.ViewProject, error) {
	stmt := table.Projects.SELECT(table.Projects.AllColumns).WHERE(
		table.Projects.ProjectID.EQ(jet.String(string(id))))

	dest := &ProjectModels{}
	err := stmt.Query(r.db, dest)
	if err != nil {
		return nil, err
	}

	if os.Getenv("MYSQL_LOGGING_ENABLED") == "true" {
		debugSql := stmt.DebugSql()
		fmt.Println("--------------------------------")
		fmt.Println(debugSql)
		fmt.Println("--------------------------------")
	}

	if len(*dest) == 0 {
		return nil, nil
	}

	out := dest.ViewProject()

	return out[0], nil
}

func (r *ProjectRepository) Store(ctx context.Context, entity *project.Project) error {
	model := model.Projects{
		ProjectID:      string(entity.ProjectID),
		ProjectUserID:  string(entity.ProjectUserID),
		Name:           entity.Name,
		Description:    entity.Description,
		ThumbnailKey:   entity.ThumbnailKey,
		Location:       entity.Location,
		Cost:           entity.Cost,
		StartDate:      entity.StartDate,
		CompletionDate: entity.CompletionDate,
		CreatedTime:    entity.CreatedTime,
		UpdatedTime:    entity.UpdatedTime,
	}

	insertStmt := table.Projects.INSERT(table.Projects.AllColumns).MODEL(model)

	updateStmt := table.Projects.UPDATE(table.Projects.AllColumns).MODEL(model)
	updateStmt = updateStmt.WHERE(table.Projects.ProjectID.EQ(jet.String(string(entity.ProjectID))))

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

func (r *ProjectRepository) Remove(ctx context.Context, ids ...project.ProjectID) error {
	idExpressions := make([]jet.Expression, 0, len(ids))
	for _, id := range ids {
		idExpressions = append(idExpressions, jet.String(string(id)))
	}

	stmt := table.Projects.DELETE().WHERE(table.Projects.ProjectID.IN(idExpressions...))
	_, err := stmt.Exec(r.db)
	if err != nil {
		return err
	}

	return nil
}

type ProjectService struct {
	db *sql.DB
}

func NewProjectService(db *sql.DB) *ProjectService {
	return &ProjectService{
		db: db,
	}
}

func (s *ProjectService) List(ctx context.Context, cond project.ListCond, option projectsvc.ListOption) ([]*projectsvc.ViewProject, error) {
	stmt := table.Projects.SELECT(table.Projects.AllColumns)
	pred := []jet.BoolExpression{}
	orderBy := []jet.OrderByClause{}

	if cond.ProjectID != nil {
		pred = append(pred, table.Projects.Name.EQ(jet.String(string(*cond.ProjectID))))
	}

	if len(cond.ProjectIDs) > 0 {
		idExpressions := make([]jet.Expression, 0, len(cond.ProjectIDs))

		for _, id := range cond.ProjectIDs {
			idExpressions = append(idExpressions, jet.String(string(id)))
		}

		pred = append(pred, table.Projects.ProjectID.IN(
			idExpressions...,
		))
	}

	switch option.SortKey {
	case projectsvc.ListOptionSortKey_CreatedAt_ASC:
		orderBy = append(orderBy, table.Projects.CreatedTime.ASC())
	case projectsvc.ListOptionSortKey_CreatedAt_DESC:
		orderBy = append(orderBy, table.Projects.CreatedTime.DESC())
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
	if os.Getenv("MYSQL_LOGGING_ENABLED") == "true" {
		fmt.Println("--------------------------------")
		fmt.Println(debugSql)
		fmt.Println("--------------------------------")
	}

	dest := &ProjectModels{}
	err := stmt.Query(s.db, dest)
	if err != nil {
		return nil, err
	}

	if len(*dest) == 0 {
		return nil, nil
	}

	out := dest.ViewProject()

	return out, nil
}

func (s *ProjectService) Count(ctx context.Context, cond project.CountCond, option projectsvc.CountOption) (*uint64, error) {
	stmt := table.Projects.SELECT(jet.COUNT(table.Projects.ProjectID).AS("count"))
	pred := []jet.BoolExpression{}

	if cond.ProjectID != nil {
		pred = append(pred, table.Projects.ProjectID.EQ(jet.String(string(*cond.ProjectID))))
	}

	if len(cond.ProjectIDs) > 0 {
		idExpressions := make([]jet.Expression, 0, len(cond.ProjectIDs))
		for _, id := range cond.ProjectIDs {
			idExpressions = append(idExpressions, jet.String(string(id)))
		}

		pred = append(pred, table.Projects.ProjectID.IN(
			idExpressions...,
		))
	}

	if len(pred) > 0 {
		stmt = stmt.WHERE(jet.AND(pred...))
	}

	debugSql := stmt.DebugSql()
	if os.Getenv("MYSQL_LOGGING_ENABLED") == "true" {
		fmt.Println("--------------------------------")
		fmt.Println(debugSql)
		fmt.Println("--------------------------------")
	}

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

type ProjectModels []struct {
	model.Projects
}

func (src ProjectModels) ViewProject() []*projectsvc.ViewProject {
	out := make([]*projectsvc.ViewProject, 0, len(src))
	for _, item := range src {
		entity := &project.Project{}
		entity.ProjectID = project.ProjectID(item.ProjectID)
		entity.ProjectUserID = project.UserID(item.ProjectUserID)
		entity.Name = item.Name
		entity.Description = item.Description
		entity.ThumbnailKey = item.ThumbnailKey
		entity.Location = item.Location
		entity.Cost = item.Cost
		entity.StartDate = item.StartDate
		entity.CompletionDate = item.CompletionDate
		entity.CreatedTime = item.CreatedTime
		entity.UpdatedTime = item.UpdatedTime

		vw := &projectsvc.ViewProject{
			Project: *entity,
		}
		out = append(out, vw)
	}
	return out
}
