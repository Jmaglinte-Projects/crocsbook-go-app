package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/project"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql/lib/db_crocs/table"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/projectsvc"
	jet "github.com/go-jet/jet/v2/mysql"
)

type ProjectService struct {
	db *sql.DB
}

func NewProjectService(db *sql.DB) *ProjectService {
	return &ProjectService{
		db: db,
	}
}

func (s *ProjectService) List(ctx context.Context, cond project.ListCond, option projectsvc.ListOption) ([]*project.Project, error) {
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
	case projectsvc.ListOptionSortKey_ASC:
		orderBy = append(orderBy, table.Projects.CreatedTime.ASC())
	case projectsvc.ListOptionSortKey_DESC:
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
	fmt.Println("--------------------------------")
	fmt.Println(debugSql)
	fmt.Println("--------------------------------")

	dest := &ProjectModels{}
	err := stmt.Query(s.db, dest)
	if err != nil {
		return nil, err
	}

	if len(*dest) == 0 {
		return nil, nil
	}

	out := dest.ToDomain()

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
