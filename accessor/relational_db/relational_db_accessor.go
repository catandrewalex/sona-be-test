package relational_db

import (
	"context"
	"database/sql"

	"sonamusica-backend/accessor/relational_db/mysql"
)

// MySQLQueries is expected to be constructed using the NewMySQLQueries.
// This struct is useful for wrapping SQLC's "Queries", which lacks of modifiability due to code generation.
//
// So far, the main goals are:
//  1. Wrap SQLC's "Queries" with WrapMySQLError()
//  2. Add shortcut to *sql.DB.Begin() without directly accessing the *sql.DB
type MySQLQueries struct {
	mysql.Queries
	db *sql.DB
}

func NewMySQLQueries(db *sql.DB) *MySQLQueries {
	wrappedDB := &mysql.DBTXWrappedError{DB: db}
	queries := mysql.New(wrappedDB)
	return &MySQLQueries{*queries, db}
}

// Begin() is a wrapper for MySQLQueries.db.Begin(). So that the caller doesn't need to directly access the *sql.DB.
func (q MySQLQueries) Begin() (*sql.Tx, error) {
	sqlTx, err := q.db.Begin()
	return sqlTx, err
}

// BeginTx() is a wrapper for MySQLQueries.db.BeginTx(). So that the caller doesn't need to directly access the *sql.DB.
func (q MySQLQueries) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	sqlTx, err := q.db.BeginTx(ctx, opts)
	return sqlTx, err
}
