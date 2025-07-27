package database

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// Transaction представляет интерфейс для работы с транзакциями
type Transaction interface {
	// Commit завершает транзакцию
	Commit() error
	// Rollback откатывает транзакцию
	Rollback() error
	// GetTx возвращает базовую транзакцию для использования в репозиториях
	GetTx() *sqlx.Tx
}

// TransactionManager управляет транзакциями
type TransactionManager interface {
	// BeginTransaction начинает новую транзакцию
	BeginTransaction(ctx context.Context) (Transaction, error)
	// WithTransaction выполняет функцию в рамках транзакции
	WithTransaction(ctx context.Context, fn func(ctx context.Context, tx Transaction) error) error
	// WithTransactionContext выполняет функцию в рамках транзакции, передавая её через context
	WithTransactionContext(ctx context.Context, fn func(ctx context.Context) error) error
}

// PostgreSQLTransaction реализация Transaction для PostgreSQL
type PostgreSQLTransaction struct {
	tx *sqlx.Tx
}

func (t *PostgreSQLTransaction) Commit() error {
	return t.tx.Commit()
}

func (t *PostgreSQLTransaction) Rollback() error {
	return t.tx.Rollback()
}

func (t *PostgreSQLTransaction) GetTx() *sqlx.Tx {
	return t.tx
}

// PostgreSQLTransactionManager реализация TransactionManager для PostgreSQL
type PostgreSQLTransactionManager struct {
	db *sqlx.DB
}

func NewPostgreSQLTransactionManager(db *sqlx.DB) TransactionManager {
	return &PostgreSQLTransactionManager{db: db}
}

func (tm *PostgreSQLTransactionManager) BeginTransaction(ctx context.Context) (Transaction, error) {
	tx, err := tm.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, err
	}

	return &PostgreSQLTransaction{tx: tx}, nil
}

func (tm *PostgreSQLTransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context, tx Transaction) error) error {
	tx, err := tm.BeginTransaction(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(ctx, tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			// Log rollback error, but return original error
			return err
		}
		return err
	}

	return tx.Commit()
}

// WithTransactionContext выполняет функцию в рамках транзакции, передавая её через context
func (tm *PostgreSQLTransactionManager) WithTransactionContext(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.BeginTransaction(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	// Добавляем транзакцию в контекст
	txCtx := WithTransaction(ctx, tx)

	if err := fn(txCtx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			// Log rollback error, but return original error
			return err
		}
		return err
	}

	return tx.Commit()
}
