package sqlcluster

import (
	`context`
	`database/sql`
	`time`
)

type SQLDatabase interface {
	Ping() error
	PingContext(ctx context.Context) error
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	SetConnMaxLifetime(d time.Duration)
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
}

type Replicas []SQLDatabase

type Config struct {
	Master       SQLDatabase
	ReadReplicas Replicas
}

func New(config Config) (SQLDatabase, error) {
	pool, err := newReplicaPool(config.ReadReplicas...)
	if err != nil {
		return nil, err
	}
	return &cluster{
		master:       config.Master,
		readReplicas: pool,
	}, nil
}
