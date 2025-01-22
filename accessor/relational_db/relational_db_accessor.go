package relational_db

import (
	"context"
	"database/sql"
	"fmt"

	"sonamusica-backend/accessor/relational_db/mysql"
)

// MySQLQueries is expected to be constructed using the NewMySQLQueries.
// This struct is useful for extending/wrapping SQLC's "Queries", which lacks of modifiability due to code generation.
//
// So far, the main goals are:
//  1. Wrap SQLC's "Queries" with WrapMySQLError()
//  2. Add shortcut to transaction using "ExecuteInTransaction()"
type MySQLQueries struct {
	mysql.Queries
	db *sql.DB
}

func NewMySQLQueries(db *sql.DB) *MySQLQueries {
	wrappedDBTX := &mysql.DBTXWrappedError{DB: db}
	queries := mysql.New(wrappedDBTX)
	return &MySQLQueries{*queries, db}
}

// ExecuteInTransaction is useful for:
//  1. Simplifying SQLC's boilerplate for database transaction (db.Begin(), WithTx(tx), Rollback(), Commit(), etc.)
//  2. Allow continuous database transaction across methods, by putting the sql.Tx inside Context.
//     Later, recursed ExecuteInTransaction() can reuse the sql.Tx if it exists inside the Context.
func (q MySQLQueries) ExecuteInTransaction(ctx context.Context, wrappedFunc func(context.Context, *mysql.Queries) error) error {
	var tx *sql.Tx
	var err error
	newCtx := ctx

	isReusingExistingTx := false
	if existingTx := GetSQLTx(ctx); existingTx != nil { // reuse existing pre-created SQL transaction (Tx) if exists
		tx = existingTx
		isReusingExistingTx = true
	} else {
		tx, err = q.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("db.BeginTx(): %w", err)
		}
		newCtx = NewContextWithSQLTx(ctx, tx)
		defer tx.Rollback()
	}

	qtx := q.WithTxWrappedError(tx)

	err = wrappedFunc(newCtx, qtx)
	if err != nil {
		return err
	}

	if !isReusingExistingTx {
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("tx.Commit(): %w", err)
		}
	}

	return nil
}

type sqlTxKey struct{}

// NewContextWithSQLTx copies a context, adds a Go's sql.Tx into it, and returns the new context.
//
// TODO: remove this and look for alternative? as we're utilizing this as optional parameter.
// Go' documentation officially doesn't recommend doing it.
func NewContextWithSQLTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, sqlTxKey{}, tx)
}

func GetSQLTx(ctx context.Context) *sql.Tx {
	sqlTx, ok := ctx.Value(sqlTxKey{}).(*sql.Tx)
	if !ok {
		return nil
	}
	return sqlTx
}
