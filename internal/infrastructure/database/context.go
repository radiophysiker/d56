package database

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// contextKey используется для типобезопасного хранения транзакции в контексте
type contextKey string

const transactionKey contextKey = "database_transaction"

// DBExecutor объединяет интерфейсы для работы с DB и транзакциями
type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// WithTransaction добавляет транзакцию в контекст
func WithTransaction(ctx context.Context, tx Transaction) context.Context {
	return context.WithValue(ctx, transactionKey, tx)
}

// GetTransaction извлекает транзакцию из контекста
func GetTransaction(ctx context.Context) (Transaction, bool) {
	tx, ok := ctx.Value(transactionKey).(Transaction)
	return tx, ok
}

// GetTxOrDB возвращает транзакцию из контекста или основной DB
// Это позволяет репозиториям работать как с транзакциями, так и без них
func GetTxOrDB(ctx context.Context, db *sqlx.DB) DBExecutor {
	if tx, ok := GetTransaction(ctx); ok {
		return tx.GetTx()
	}
	return db
}
