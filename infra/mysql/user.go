package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/user"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql/lib/db_crocs/model"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql/lib/db_crocs/table"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/usersvc"
	jet "github.com/go-jet/jet/v2/mysql"
	"github.com/go-sql-driver/mysql"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) usersvc.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Find(ctx context.Context, id user.UserID) (*usersvc.ViewUser, error) {
	stmt := table.Users.SELECT(table.Users.AllColumns).WHERE(
		table.Users.UserID.EQ(jet.String(string(id))))

	dest := &UserModels{}
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

	out := dest.ViewUser()

	return out[0], nil
}

func (r *userRepository) Store(ctx context.Context, entity *user.User) error {
	model := model.Users{
		UserID:      string(entity.UserID),
		Email:       entity.Email,
		Gender:      string(entity.Gender),
		ProfileURL:  entity.ProfileURL,
		Nickname:    entity.Nickname,
		Username:    entity.Username,
		Password:    entity.Password,
		CreatedTime: entity.CreatedTime,
		UpdatedTime: entity.UpdatedTime,
	}

	insertStmt := table.Users.INSERT(table.Users.AllColumns).MODEL(model)

	updateStmt := table.Users.UPDATE(table.Users.AllColumns).MODEL(model)
	updateStmt = updateStmt.WHERE(table.Users.UserID.EQ(jet.String(string(entity.UserID))))

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

func (r *userRepository) Remove(ctx context.Context, ids ...user.UserID) error {
	idExpressions := make([]jet.Expression, 0, len(ids))
	for _, id := range ids {
		idExpressions = append(idExpressions, jet.String(string(id)))
	}

	stmt := table.Users.DELETE().WHERE(table.Users.UserID.IN(idExpressions...))
	_, err := stmt.Exec(r.db)
	if err != nil {
		return err
	}

	return nil
}

type userService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) usersvc.UserService {
	return &userService{
		db: db,
	}
}

func (s *userService) List(ctx context.Context, cond user.ListCond, option usersvc.ListOption) ([]*usersvc.ViewUser, error) {
	stmt := table.Users.SELECT(table.Users.AllColumns)
	pred := []jet.BoolExpression{}
	orderBy := []jet.OrderByClause{}

	if cond.UserID != nil {
		pred = append(pred, table.Users.UserID.EQ(jet.String(string(*cond.UserID))))
	}

	if len(cond.UserIDs) > 0 {
		idExpressions := make([]jet.Expression, 0, len(cond.UserIDs))

		for _, id := range cond.UserIDs {
			idExpressions = append(idExpressions, jet.String(string(id)))
		}

		pred = append(pred, table.Users.UserID.IN(
			idExpressions...,
		))
	}

	switch option.SortKey {
	case usersvc.ListOptionSortKey_CreatedAt_ASC:
		orderBy = append(orderBy, table.Users.CreatedTime.ASC())
	case usersvc.ListOptionSortKey_CreatedAt_DESC:
		orderBy = append(orderBy, table.Users.CreatedTime.DESC())
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

	dest := &UserModels{}
	err := stmt.Query(s.db, dest)
	if err != nil {
		return nil, err
	}

	if len(*dest) == 0 {
		return nil, nil
	}

	out := dest.ViewUser()

	return out, nil
}

func (s *userService) Count(ctx context.Context, cond user.CountCond, option usersvc.CountOption) (*uint64, error) {
	stmt := table.Users.SELECT(jet.COUNT(table.Users.UserID).AS("count"))
	pred := []jet.BoolExpression{}

	if cond.UserID != nil {
		pred = append(pred, table.Users.UserID.EQ(jet.String(string(*cond.UserID))))
	}

	if len(cond.UserIDs) > 0 {
		idExpressions := make([]jet.Expression, 0, len(cond.UserIDs))
		for _, id := range cond.UserIDs {
			idExpressions = append(idExpressions, jet.String(string(id)))
		}

		pred = append(pred, table.Users.UserID.IN(
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

type UserModels []struct {
	model.Users
}

func (src UserModels) ViewUser() []*usersvc.ViewUser {
	out := make([]*usersvc.ViewUser, 0, len(src))
	for _, item := range src {
		userEntity := &user.User{
			UserID:      user.UserID(item.UserID),
			Email:       item.Email,
			Gender:      user.Gender(item.Gender),
			ProfileURL:  item.ProfileURL,
			Nickname:    item.Nickname,
			Username:    item.Username,
			Password:    item.Password,
			CreatedTime: item.CreatedTime,
			UpdatedTime: item.UpdatedTime,
		}

		vw := &usersvc.ViewUser{
			User: *userEntity,
		}
		out = append(out, vw)
	}
	return out
}
