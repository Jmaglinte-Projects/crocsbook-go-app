package mysql

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	DSN = "root:@dm1n1234@tcp(localhost:3306)/db_crocs?loc=Local&parseTime=true"
)

func SetupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("mysql", DSN)
	assert.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	return db
}
