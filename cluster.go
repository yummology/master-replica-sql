package sqlcluster

import (
	"context"
	"database/sql"
	"time"
)

type cluster struct {
	master        SQLDatabase
	readReplicas  *replicaPool
	replicasCount int64
	nextReplica   int64
}

func (c *cluster) Ping() error {
	return c.PingContext(context.Background())
}

func (c *cluster) PingContext(ctx context.Context) error {
	if err := c.master.PingContext(ctx); err != nil {
		return err
	}
	if err := c.readReplicas.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

// Query executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
func (c *cluster) Query(query string, args ...interface{}) (rows *sql.Rows, err error) {
	c.readReplicas.RunOnNextReplica(func(_ int, replica SQLDatabase) error {
		rows, err = replica.Query(query, args...)
		return err
	})
	return rows, err
}

// QueryContext executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
func (c *cluster) QueryContext(ctx context.Context, query string, args ...interface{}) (rows *sql.Rows, err error) {
	c.readReplicas.RunOnNextReplica(func(_ int, replica SQLDatabase) error {
		rows, err = replica.QueryContext(ctx, query, args...)
		return err
	})
	return rows, err
}

// QueryRow executes a query that is expected to return at most one row.
// QueryRow always returns a non-nil value. Errors are deferred until
// Row's Scan method is called.
// If the query selects no rows, the *Row's Scan will return ErrNoRows.
// Otherwise, the *Row's Scan scans the first selected row and discards
// the rest.
func (c *cluster) QueryRow(query string, args ...interface{}) *sql.Row {
	return c.QueryRowContext(nil, query, args...)
}

// QueryRowContext executes a query that is expected to return at most one row.
// QueryRowContext always returns a non-nil value. Errors are deferred until
// Row's Scan method is called.
// If the query selects no rows, the *Row's Scan will return ErrNoRows.
// Otherwise, the *Row's Scan scans the first selected row and discards
// the rest.
func (c *cluster) QueryRowContext(ctx context.Context, query string, args ...interface{}) (rows *sql.Row) {
	c.readReplicas.RunOnNextReplica(func(_ int, replica SQLDatabase) error {
		rows = replica.QueryRowContext(ctx, query, args...)
		return ctx.Err()
	})
	return rows
}

func (c *cluster) Begin() (*sql.Tx, error) {
	return c.master.Begin()
}

func (c *cluster) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return c.master.BeginTx(ctx, opts)
}

func (c *cluster) Close() error {
	c.master.Close()
	c.readReplicas.Walk(func(replica SQLDatabase) {
		replica.Close()
	})
	return nil
}

func (c *cluster) Exec(query string, args ...interface{}) (sql.Result, error) {
	return c.master.Exec(query, args...)
}

func (c *cluster) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return c.master.ExecContext(ctx, query, args...)
}

func (c *cluster) Prepare(query string) (*sql.Stmt, error) {
	return c.master.Prepare(query)
}

func (c *cluster) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return c.master.PrepareContext(ctx, query)
}

func (c *cluster) SetConnMaxLifetime(d time.Duration) {
	c.master.SetConnMaxLifetime(d)
	c.readReplicas.Walk(func(replica SQLDatabase) {
		replica.SetConnMaxLifetime(d)
	})
}

func (c *cluster) SetMaxIdleConns(n int) {
	c.master.SetMaxIdleConns(n)
	c.readReplicas.Walk(func(replica SQLDatabase) {
		replica.SetMaxIdleConns(n)
	})
}

func (c *cluster) SetMaxOpenConns(n int) {
	c.master.SetMaxOpenConns(n)
	c.readReplicas.Walk(func(replica SQLDatabase) {
		replica.SetMaxOpenConns(n)
	})
}
