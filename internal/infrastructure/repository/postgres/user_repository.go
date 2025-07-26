package postgres

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/radiophysiker/d56/internal/domain/user"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	query := `
		INSERT INTO users (id, login, password_hash, current_balance, withdrawn_balance, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			current_balance = EXCLUDED.current_balance,
			withdrawn_balance = EXCLUDED.withdrawn_balance
	`
	balance := u.Balance()
	_, err := r.db.ExecContext(ctx, query, u.ID(), u.Login(), u.PasswordHash(), balance.Current, balance.Withdrawn, u.CreatedAt())
	return err
}

func (r *UserRepository) FindByLogin(ctx context.Context, login string) (*user.User, error) {
	query := `SELECT id, login, password_hash, current_balance, withdrawn_balance, created_at FROM users WHERE login = $1`

	var id user.UserID
	var userLogin, passwordHash string
	var balance user.Balance
	var createdAt time.Time

	err := r.db.QueryRowContext(ctx, query, login).Scan(
		&id, &userLogin, &passwordHash, &balance.Current, &balance.Withdrawn, &createdAt,
	)
	if err != nil {
		return nil, err
	}

	return user.NewUserFromRepository(id, userLogin, passwordHash, balance, createdAt), nil
}

func (r *UserRepository) FindByID(ctx context.Context, id user.UserID) (*user.User, error) {
	query := `SELECT id, login, password_hash, current_balance, withdrawn_balance, created_at FROM users WHERE id = $1`

	var userID user.UserID
	var login, passwordHash string
	var balance user.Balance
	var createdAt time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&userID, &login, &passwordHash, &balance.Current, &balance.Withdrawn, &createdAt,
	)
	if err != nil {
		return nil, err
	}

	return user.NewUserFromRepository(userID, login, passwordHash, balance, createdAt), nil
}
