// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package sqlcgen

import (
	"context"
	"database/sql"
	"fmt"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

func Prepare(ctx context.Context, db DBTX) (*Queries, error) {
	q := Queries{db: db}
	var err error
	if q.getUserStmt, err = db.PrepareContext(ctx, getUser); err != nil {
		return nil, fmt.Errorf("error preparing query GetUser: %w", err)
	}
	if q.insertUserStmt, err = db.PrepareContext(ctx, insertUser); err != nil {
		return nil, fmt.Errorf("error preparing query InsertUser: %w", err)
	}
	if q.list10UsersStmt, err = db.PrepareContext(ctx, list10Users); err != nil {
		return nil, fmt.Errorf("error preparing query List10Users: %w", err)
	}
	return &q, nil
}

func (q *Queries) Close() error {
	var err error
	if q.getUserStmt != nil {
		if cerr := q.getUserStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing getUserStmt: %w", cerr)
		}
	}
	if q.insertUserStmt != nil {
		if cerr := q.insertUserStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing insertUserStmt: %w", cerr)
		}
	}
	if q.list10UsersStmt != nil {
		if cerr := q.list10UsersStmt.Close(); cerr != nil {
			err = fmt.Errorf("error closing list10UsersStmt: %w", cerr)
		}
	}
	return err
}

func (q *Queries) exec(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (sql.Result, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).ExecContext(ctx, args...)
	case stmt != nil:
		return stmt.ExecContext(ctx, args...)
	default:
		return q.db.ExecContext(ctx, query, args...)
	}
}

func (q *Queries) query(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) (*sql.Rows, error) {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryContext(ctx, args...)
	default:
		return q.db.QueryContext(ctx, query, args...)
	}
}

func (q *Queries) queryRow(ctx context.Context, stmt *sql.Stmt, query string, args ...interface{}) *sql.Row {
	switch {
	case stmt != nil && q.tx != nil:
		return q.tx.StmtContext(ctx, stmt).QueryRowContext(ctx, args...)
	case stmt != nil:
		return stmt.QueryRowContext(ctx, args...)
	default:
		return q.db.QueryRowContext(ctx, query, args...)
	}
}

type Queries struct {
	db              DBTX
	tx              *sql.Tx
	getUserStmt     *sql.Stmt
	insertUserStmt  *sql.Stmt
	list10UsersStmt *sql.Stmt
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db:              tx,
		tx:              tx,
		getUserStmt:     q.getUserStmt,
		insertUserStmt:  q.insertUserStmt,
		list10UsersStmt: q.list10UsersStmt,
	}
}
