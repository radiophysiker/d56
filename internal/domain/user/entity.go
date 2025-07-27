package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type UserID = uuid.UUID

type User struct {
	id           UserID
	login        string
	passwordHash string
	balance      Balance
	createdAt    time.Time
}

type Balance struct {
	Current   float64
	Withdrawn float64
}

func NewUser(login, passwordHash string) (*User, error) {
	if login == "" {
		return nil, errors.New("login cannot be empty")
	}
	if passwordHash == "" {
		return nil, errors.New("password hash cannot be empty")
	}

	return &User{
		id:           uuid.New(),
		login:        login,
		passwordHash: passwordHash,
		balance:      Balance{Current: 0, Withdrawn: 0},
		createdAt:    time.Now(),
	}, nil
}

func (u *User) ID() UserID {
	return u.id
}

func (u *User) Login() string {
	return u.login
}

func (u *User) PasswordHash() string {
	return u.passwordHash
}

func (u *User) Balance() Balance {
	return u.balance
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdateBalance(current, withdrawn float64) {
	u.balance.Current = current
	u.balance.Withdrawn = withdrawn
}

func (u *User) AddBalance(amount float64) {
	u.balance.Current += amount
}

func (u *User) WithdrawBalance(amount float64) error {
	if u.balance.Current < amount {
		return errors.New("insufficient funds")
	}
	u.balance.Current -= amount
	u.balance.Withdrawn += amount
	return nil
}

// NewUserFromRepository создает пользователя из данных репозитория
func NewUserFromRepository(id UserID, login, passwordHash string, balance Balance, createdAt time.Time) *User {
	return &User{
		id:           id,
		login:        login,
		passwordHash: passwordHash,
		balance:      balance,
		createdAt:    createdAt,
	}
}
