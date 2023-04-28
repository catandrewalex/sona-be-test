package relational_db

import (
	"context"
	"database/sql"

	"sonamusica-backend/accessor/relational_db/mysql"
	"sonamusica-backend/errs"
)

// MySQLQueries is expected to be constructed using the NewMySQLQueries.
// This struct is useful for wrapping SQLC's "Queries", which lacks of modification due to code generation.
//
//	So far, the main goal is to wrap SQLC's "Queries" with WrapMySQLError.
type MySQLQueries struct {
	mysql.Queries
	DB *sql.DB
}

func NewMySQLQueries(db *sql.DB) *MySQLQueries {
	wrappedDB := &dbtxWrappedError{db}
	queries := mysql.New(wrappedDB)
	return &MySQLQueries{*queries, db}
}

type dbtxWrappedError struct {
	db mysql.DBTX
}

func (w dbtxWrappedError) ExecContext(ctx context.Context, sqlQuery string, params ...interface{}) (sql.Result, error) {
	result, err := w.db.ExecContext(ctx, sqlQuery, params...)
	return result, errs.WrapMySQLError(err)
}
func (w dbtxWrappedError) PrepareContext(ctx context.Context, sqlQuery string) (*sql.Stmt, error) {
	result, err := w.db.PrepareContext(ctx, sqlQuery)
	return result, errs.WrapMySQLError(err)

}
func (w dbtxWrappedError) QueryContext(ctx context.Context, sqlQuery string, params ...interface{}) (*sql.Rows, error) {
	result, err := w.db.QueryContext(ctx, sqlQuery, params...)
	return result, errs.WrapMySQLError(err)

}
func (w dbtxWrappedError) QueryRowContext(ctx context.Context, sqlQuery string, params ...interface{}) *sql.Row {
	return w.db.QueryRowContext(ctx, sqlQuery, params...)
}
