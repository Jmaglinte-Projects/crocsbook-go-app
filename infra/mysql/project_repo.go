package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/project"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql/lib/db_crocs/model"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql/lib/db_crocs/table"
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

func (r *ProjectRepository) Find(ctx context.Context, id project.ProjectID) (*project.Project, error) {
	stmt := table.Projects.SELECT(table.Projects.AllColumns).WHERE(
		table.Projects.ProjectID.EQ(jet.String(string(id))))

	dest := &ProjectModels{}
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

	out := dest.ToDomain()

	return out[0], nil
}

func (r *ProjectRepository) Store(ctx context.Context, entity *project.Project) error {
	model := model.Projects{
		ProjectID:      string(entity.ProjectID),
		ProjectUserID:  string(entity.ProjectUserID),
		Name:           entity.Name,
		Description:    entity.Description,
		Thumbnail:      entity.Thumbnail,
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

func (r *ProjectRepository) Remove(ctx context.Context, ids ...string) error {
	idExpressions := make([]jet.Expression, 0, len(ids))
	for _, id := range ids {
		idExpressions = append(idExpressions, jet.String(id))
	}

	stmt := table.Projects.DELETE().WHERE(table.Projects.ProjectID.IN(idExpressions...))
	_, err := stmt.Exec(r.db)
	if err != nil {
		return err
	}

	return nil
}

type ProjectModels []struct {
	model.Projects
}

func (src ProjectModels) ToDomain() []*project.Project {
	out := make([]*project.Project, 0, len(src))
	for _, item := range src {
		entity := &project.Project{
			ProjectID:      project.ProjectID(item.ProjectID),
			ProjectUserID:  project.UserID(item.ProjectUserID),
			Name:           item.Name,
			Description:    item.Description,
			Thumbnail:      item.Thumbnail,
			Location:       item.Location,
			Cost:           item.Cost,
			StartDate:      item.StartDate,
			CompletionDate: item.CompletionDate,
			CreatedTime:    item.CreatedTime,
			UpdatedTime:    item.UpdatedTime,
		}
		out = append(out, entity)
	}
	return out
}
